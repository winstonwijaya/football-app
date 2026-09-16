package service_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/model"
	"football-app/internal/repository"
	"football-app/internal/service"
	"football-app/pkg/apperror"
)

// fakeTeamRepository is an in-memory stand-in for repository.TeamRepository.
type fakeTeamRepository struct {
	teamsByID map[int64]*model.Team
	nextID    int64
}

func newFakeTeamRepository() *fakeTeamRepository {
	return &fakeTeamRepository{teamsByID: map[int64]*model.Team{}, nextID: 1}
}

func (f *fakeTeamRepository) Create(_ context.Context, team *model.Team) error {
	team.ID = f.nextID
	f.nextID++
	f.teamsByID[team.ID] = team
	return nil
}

func (f *fakeTeamRepository) FindByID(_ context.Context, id int64) (*model.Team, error) {
	team, ok := f.teamsByID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return team, nil
}

func (f *fakeTeamRepository) List(_ context.Context, _ repository.TeamFilter) ([]model.Team, int64, error) {
	teams := make([]model.Team, 0, len(f.teamsByID))
	for _, t := range f.teamsByID {
		teams = append(teams, *t)
	}
	return teams, int64(len(teams)), nil
}

func (f *fakeTeamRepository) Update(_ context.Context, team *model.Team) error {
	existing, ok := f.teamsByID[team.ID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	existing.Name = team.Name
	existing.LogoURL = team.LogoURL
	existing.EstablishedYear = team.EstablishedYear
	existing.Address = team.Address
	existing.City = team.City
	existing.UpdatedBy = team.UpdatedBy
	return nil
}

func (f *fakeTeamRepository) SoftDelete(_ context.Context, id int64, _ int64) error {
	if _, ok := f.teamsByID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.teamsByID, id)
	return nil
}

func TestTeamService_CreateAndGet(t *testing.T) {
	svc := service.NewTeamService(newFakeTeamRepository())

	created, err := svc.Create(context.Background(), dto.CreateTeamRequest{Name: "FC Testing"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected a non-zero ID")
	}

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "FC Testing" {
		t.Errorf("got name %q, want %q", got.Name, "FC Testing")
	}
}

func TestTeamService_GetNotFound(t *testing.T) {
	svc := service.NewTeamService(newFakeTeamRepository())

	_, err := svc.Get(context.Background(), 999)
	assertNotFound(t, err)
}

func TestTeamService_UpdateNotFound(t *testing.T) {
	svc := service.NewTeamService(newFakeTeamRepository())

	_, err := svc.Update(context.Background(), 999, dto.UpdateTeamRequest{Name: "X"}, 1)
	assertNotFound(t, err)
}

func TestTeamService_DeleteNotFound(t *testing.T) {
	svc := service.NewTeamService(newFakeTeamRepository())

	err := svc.Delete(context.Background(), 999, 1)
	assertNotFound(t, err)
}

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Errorf("got code %s, want %s", appErr.Code, apperror.CodeNotFound)
	}
}
