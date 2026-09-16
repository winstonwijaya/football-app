package repository

import (
	"context"

	"gorm.io/gorm"

	"football-app/internal/model"
)

// UserRepository is the interface services depend on, so auth logic can be unit-tested against a mock instead of a real database.
type UserRepository interface {
	// Check if a user with the given username exists. Returns gorm.ErrRecordNotFound if not found.
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	// Create a user, currently only used through seeding since the app doesn't have a registration endpoint.
	Create(ctx context.Context, user *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db
	}
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).
		Create(user).
		Error
}
