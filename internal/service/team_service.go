package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/model"
	"football-app/internal/repository"
	"football-app/pkg/apperror"
)

type TeamService struct {
	teams repository.TeamRepository
}

func NewTeamService(teams repository.TeamRepository) *TeamService {
	return &TeamService{teams: teams}
}

func (s *TeamService) Create(ctx context.Context, req dto.CreateTeamRequest, userID int64) (*dto.TeamResponse, error) {
	team := &model.Team{
		Name:            req.Name,
		LogoURL:         req.LogoURL,
		EstablishedYear: req.EstablishedYear,
		Address:         req.Address,
		City:            req.City,
	}
	team.CreatedBy = &userID
	team.UpdatedBy = &userID

	if err := s.teams.Create(ctx, team); err != nil {
		return nil, apperror.Internal(err)
	}

	resp := teamToResponse(*team)
	return &resp, nil
}

func (s *TeamService) Get(ctx context.Context, id int64) (*dto.TeamResponse, error) {
	team, err := s.teams.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NotFound("team not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := teamToResponse(*team)
	return &resp, nil
}

func (s *TeamService) List(ctx context.Context, query dto.ListTeamsQuery) ([]dto.TeamResponse, int64, error) {
	teams, total, err := s.teams.List(ctx, repository.TeamFilter{
		City:   query.City,
		Search: query.Search,
		Page:   query.Page,
		Limit:  query.Limit,
	})
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}

	items := make([]dto.TeamResponse, len(teams))
	for i, t := range teams {
		items[i] = teamToResponse(t)
	}
	return items, total, nil
}

func (s *TeamService) Update(ctx context.Context, id int64, req dto.UpdateTeamRequest, userID int64) (*dto.TeamResponse, error) {
	team := &model.Team{
		ID:              id,
		Name:            req.Name,
		LogoURL:         req.LogoURL,
		EstablishedYear: req.EstablishedYear,
		Address:         req.Address,
		City:            req.City,
	}
	team.UpdatedBy = &userID

	if err := s.teams.Update(ctx, team); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("team not found")
		}
		return nil, apperror.Internal(err)
	}

	updated, err := s.teams.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := teamToResponse(*updated)
	return &resp, nil
}

func (s *TeamService) Delete(ctx context.Context, id int64, userID int64) error {
	err := s.teams.SoftDelete(ctx, id, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("team not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func teamToResponse(t model.Team) dto.TeamResponse {
	return dto.TeamResponse{
		ID:              t.ID,
		Name:            t.Name,
		LogoURL:         t.LogoURL,
		EstablishedYear: t.EstablishedYear,
		Address:         t.Address,
		City:            t.City,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
