package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"football-app/internal/model"
	"football-app/internal/service"
	"football-app/pkg/apperror"
	"football-app/pkg/hash"
)

// fakeUserRepository is an in-memory stand-in for repository.UserRepository,
// used to unit-test AuthService.Login without a real database.
type fakeUserRepository struct {
	usersByUsername map[string]*model.User
}

func (f *fakeUserRepository) FindByUsername(_ context.Context, username string) (*model.User, error) {
	user, ok := f.usersByUsername[username]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (f *fakeUserRepository) Create(_ context.Context, user *model.User) error {
	f.usersByUsername[user.Username] = user
	return nil
}

func TestAuthService_Login(t *testing.T) {
	passwordHash, err := hash.Hash("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &fakeUserRepository{usersByUsername: map[string]*model.User{
		"admin": {ID: 1, Username: "admin", FirstName: "Admin", LastName: "User", PasswordHash: passwordHash},
	}}
	svc := service.NewAuthService(repo, "test-secret", time.Hour)

	t.Run("valid credentials returns a token", func(t *testing.T) {
		resp, err := svc.Login(context.Background(), "admin", "correct-password")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.AccessToken == "" {
			t.Error("expected a non-empty access token")
		}
		if resp.User.ID != 1 {
			t.Errorf("got user id %d, want 1", resp.User.ID)
		}
	})

	t.Run("wrong password is unauthorized", func(t *testing.T) {
		_, err := svc.Login(context.Background(), "admin", "wrong-password")
		assertUnauthorized(t, err)
	})

	t.Run("unknown username is unauthorized", func(t *testing.T) {
		_, err := svc.Login(context.Background(), "nobody", "whatever")
		assertUnauthorized(t, err)
	})
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeUnauthorized {
		t.Errorf("got code %s, want %s", appErr.Code, apperror.CodeUnauthorized)
	}
}
