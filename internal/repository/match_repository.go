package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"football-app/internal/model"
)

type MatchFilter struct {
	Status string
	TeamID *int64
	From   *time.Time
	To     *time.Time
	Page   int
	Limit  int
}

type MatchRepository interface {
	// CreateWithConflictCheck creates match inside a transaction that also
	// enforces "a team plays at most one match per calendar day" — this
	// can't be a simple unique index since a team can appear in either the
	// home or away column, so the check-then-insert must be atomic.
	// Returns ErrConflict if either team already has a non-cancelled match
	// that day.
	CreateWithConflictCheck(ctx context.Context, match *model.Match) error
	// FindByID returns gorm.ErrRecordNotFound if no active match matches.
	FindByID(ctx context.Context, id int64) (*model.Match, error)
	List(ctx context.Context, filter MatchFilter) ([]model.Match, int64, error)
	// UpdateWithConflictCheck is the update equivalent of
	// CreateWithConflictCheck, excluding the match being updated from its
	// own conflict check. Updates schedule fields and status only — never
	// scores, which are exclusively written by result reporting.
	UpdateWithConflictCheck(ctx context.Context, match *model.Match) error
	// SoftDelete returns gorm.ErrRecordNotFound if no active match matches id.
	SoftDelete(ctx context.Context, id int64, deletedBy int64) error
	// ReportResult writes match's home_score/away_score/status and all
	// logs rows in one transaction, row-locking the match first so two
	// concurrent reports on the same match can't both succeed. Returns
	// ErrConflict if the match has already been reported.
	ReportResult(ctx context.Context, match *model.Match, logs []model.MatchLog, userID int64) error
}

type matchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) CreateWithConflictCheck(ctx context.Context, match *model.Match) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		conflict, err := hasScheduleConflict(tx, match.HomeTeamID, match.AwayTeamID, match.MatchDatetime, 0)
		if err != nil {
			return err
		}
		if conflict {
			return ErrConflict
		}
		return tx.Create(match).Error
	})
}

func (r *matchRepository) FindByID(ctx context.Context, id int64) (*model.Match, error) {
	var match model.Match
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) List(ctx context.Context, filter MatchFilter) ([]model.Match, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Match{}).Where("deleted_at IS NULL")
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.TeamID != nil {
		query = query.Where("home_team_id = ? OR away_team_id = ?", *filter.TeamID, *filter.TeamID)
	}
	if filter.From != nil {
		query = query.Where("match_datetime >= ?", *filter.From)
	}
	if filter.To != nil {
		query = query.Where("match_datetime <= ?", *filter.To)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var matches []model.Match
	err := paginate(query.Order("match_datetime"), filter.Page, filter.Limit).Find(&matches).Error
	if err != nil {
		return nil, 0, err
	}
	return matches, total, nil
}

func (r *matchRepository) UpdateWithConflictCheck(ctx context.Context, match *model.Match) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		conflict, err := hasScheduleConflict(tx, match.HomeTeamID, match.AwayTeamID, match.MatchDatetime, match.ID)
		if err != nil {
			return err
		}
		if conflict {
			return ErrConflict
		}

		result := tx.Model(&model.Match{}).
			Where("id = ? AND deleted_at IS NULL", match.ID).
			Select("home_team_id", "away_team_id", "match_datetime", "status", "updated_at", "updated_by").
			Updates(match)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *matchRepository) SoftDelete(ctx context.Context, id int64, deletedBy int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Match{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *matchRepository) ReportResult(ctx context.Context, match *model.Match, logs []model.MatchLog, userID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.Match
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", match.ID).
			First(&current).Error
		if err != nil {
			return err
		}
		if current.Status == model.MatchStatusPlayed {
			return ErrConflict
		}

		if err := tx.Model(&model.Match{}).
			Where("id = ?", match.ID).
			Select("home_score", "away_score", "status", "updated_at", "updated_by").
			Updates(match).Error; err != nil {
			return err
		}

		for i := range logs {
			logs[i].CreatedBy = &userID
		}
		if len(logs) > 0 {
			if err := tx.Create(&logs).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// hasScheduleConflict reports whether homeTeamID or awayTeamID already has
// a non-cancelled match on the same calendar date as matchDatetime,
// excluding excludeMatchID (0 on create, since ids start at 1). Must run
// inside the same transaction as the write it's guarding.
func hasScheduleConflict(tx *gorm.DB, homeTeamID, awayTeamID int64, matchDatetime time.Time, excludeMatchID int64) (bool, error) {
	var count int64
	err := tx.Model(&model.Match{}).
		Where("deleted_at IS NULL").
		Where("status <> ?", model.MatchStatusCancelled).
		Where("id <> ?", excludeMatchID).
		Where("DATE(match_datetime) = DATE(?)", matchDatetime).
		Where("home_team_id IN (?, ?) OR away_team_id IN (?, ?)", homeTeamID, awayTeamID, homeTeamID, awayTeamID).
		Count(&count).Error
	return count > 0, err
}
