package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"football-app/internal/model"
)

type PlayerFilter struct {
	TeamID   *int64
	Position string
	Page     int
	Limit    int
}

type PlayerRepository interface {
	// Create returns ErrConflict if the squad number is already taken
	// within the team.
	Create(ctx context.Context, player *model.Player) error
	// FindByID returns gorm.ErrRecordNotFound if no active player matches.
	FindByID(ctx context.Context, id int64) (*model.Player, error)
	List(ctx context.Context, filter PlayerFilter) ([]model.Player, int64, error)
	// Update returns gorm.ErrRecordNotFound if no active player matches
	// player.ID, or ErrConflict on a squad number collision.
	Update(ctx context.Context, player *model.Player) error
	// SoftDelete returns gorm.ErrRecordNotFound if no active player matches id.
	SoftDelete(ctx context.Context, id int64, deletedBy int64) error
}

type playerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) PlayerRepository {
	return &playerRepository{db: db}
}

func (r *playerRepository) Create(ctx context.Context, player *model.Player) error {
	err := r.db.WithContext(ctx).Create(player).Error
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

func (r *playerRepository) FindByID(ctx context.Context, id int64) (*model.Player, error) {
	var player model.Player
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) List(ctx context.Context, filter PlayerFilter) ([]model.Player, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Player{}).Where("deleted_at IS NULL")
	if filter.TeamID != nil {
		query = query.Where("team_id = ?", *filter.TeamID)
	}
	if filter.Position != "" {
		query = query.Where("position = ?", filter.Position)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var players []model.Player
	err := paginate(query.Order("id"), filter.Page, filter.Limit).Find(&players).Error
	if err != nil {
		return nil, 0, err
	}
	return players, total, nil
}

func (r *playerRepository) Update(ctx context.Context, player *model.Player) error {
	result := r.db.WithContext(ctx).
		Model(&model.Player{}).
		Where("id = ? AND deleted_at IS NULL", player.ID).
		Select("team_id", "name", "height", "weight", "position", "squad_number", "updated_at", "updated_by").
		Updates(player)
	if isUniqueViolation(result.Error) {
		return ErrConflict
	}
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *playerRepository) SoftDelete(ctx context.Context, id int64, deletedBy int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Player{}).
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
