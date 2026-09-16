package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/repository"
	"football-app/pkg/apperror"
	"football-app/pkg/hash"
	"football-app/pkg/token"
)

type AuthService struct {
	users     repository.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthService(users repository.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, jwtTTL: jwtTTL}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*dto.LoginResponse, error) {
	user, err := s.users.FindByUsername(ctx, username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.Unauthorized("invalid username or password")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	if !hash.Compare(user.PasswordHash, password) {
		return nil, apperror.Unauthorized("invalid username or password")
	}

	accessToken, expiresAt, err := token.Generate(s.jwtSecret, user.ID, s.jwtTTL)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
		User: dto.UserSummary{
			ID:        user.ID,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
	}, nil
}
