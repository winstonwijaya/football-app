package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"football-app/internal/model"
)

type TeamFilter struct {
	City   string
	Search string
	Page   int
	Limit  int
}

type TeamRepository interface {
	Create(ctx context.Context, team *model.Team) error
	FindByID(ctx context.Context, id int64) (*model.Team, error)
	List(ctx context.Context, filter TeamFilter) ([]model.Team, int64, error)
	Update(ctx context.Context, team *model.Team) error
	SoftDelete(ctx context.Context, id int64, deletedBy int64) error
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(ctx context.Context, team *model.Team) error {
	return r.db.WithContext(ctx).
		Create(team).Error
}

func (r *teamRepository) FindByID(ctx context.Context, id int64) (*model.Team, error) {
	var team model.Team

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&team).Error
	if err != nil {
		return nil, err
	}

	return &team, nil
}

func (r *teamRepository) List(ctx context.Context, filter TeamFilter) ([]model.Team, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Where("deleted_at IS NULL")

	if filter.City != "" {
		query = query.Where("city = ?", filter.City)
	}
	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var teams []model.Team
	err := paginate(query.Order("id"), filter.Page, filter.Limit).Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}
	return teams, total, nil
}

func (r *teamRepository) Update(ctx context.Context, team *model.Team) error {
	result := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Where("id = ? AND deleted_at IS NULL", team.ID).
		Select("name", "logo_url", "established_year", "address", "city", "updated_at", "updated_by").
		Updates(team)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *teamRepository) SoftDelete(ctx context.Context, id int64, deletedBy int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Team{}).
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
