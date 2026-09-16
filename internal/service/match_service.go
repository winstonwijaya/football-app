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

const scheduleConflictMessage = "a team in this match already has a match scheduled on this date"

type MatchService struct {
	matches repository.MatchRepository
	teams   repository.TeamRepository
}

func NewMatchService(matches repository.MatchRepository, teams repository.TeamRepository) *MatchService {
	return &MatchService{matches: matches, teams: teams}
}

func (s *MatchService) Create(ctx context.Context, req dto.CreateMatchRequest, userID int64) (*dto.MatchResponse, error) {
	if err := s.requireTeam(ctx, req.HomeTeamID); err != nil {
		return nil, err
	}
	if err := s.requireTeam(ctx, req.AwayTeamID); err != nil {
		return nil, err
	}

	match := &model.Match{
		HomeTeamID:    req.HomeTeamID,
		AwayTeamID:    req.AwayTeamID,
		MatchDatetime: req.MatchDatetime,
		Status:        model.MatchStatusScheduled,
	}
	match.CreatedBy = &userID
	match.UpdatedBy = &userID

	if err := s.matches.CreateWithConflictCheck(ctx, match); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, apperror.Conflict(scheduleConflictMessage)
		}
		return nil, apperror.Internal(err)
	}

	resp := matchToResponse(*match)
	return &resp, nil
}

func (s *MatchService) Get(ctx context.Context, id int64) (*dto.MatchResponse, error) {
	match, err := s.matches.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NotFound("match not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := matchToResponse(*match)
	return &resp, nil
}

func (s *MatchService) List(ctx context.Context, query dto.ListMatchesQuery) ([]dto.MatchResponse, int64, error) {
	matches, total, err := s.matches.List(ctx, repository.MatchFilter{
		Status: query.Status,
		TeamID: query.TeamID,
		From:   query.From,
		To:     query.To,
		Page:   query.Page,
		Limit:  query.Limit,
	})
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}

	items := make([]dto.MatchResponse, len(matches))
	for i, m := range matches {
		items[i] = matchToResponse(m)
	}
	return items, total, nil
}

func (s *MatchService) Update(ctx context.Context, id int64, req dto.UpdateMatchRequest, userID int64) (*dto.MatchResponse, error) {
	existing, err := s.matches.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NotFound("match not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if existing.Status == model.MatchStatusPlayed {
		return nil, apperror.Conflict("cannot modify a match that has already been played")
	}

	if err := s.requireTeam(ctx, req.HomeTeamID); err != nil {
		return nil, err
	}
	if err := s.requireTeam(ctx, req.AwayTeamID); err != nil {
		return nil, err
	}

	match := &model.Match{
		ID:            id,
		HomeTeamID:    req.HomeTeamID,
		AwayTeamID:    req.AwayTeamID,
		MatchDatetime: req.MatchDatetime,
		Status:        model.MatchStatus(req.Status),
	}
	match.UpdatedBy = &userID

	if err := s.matches.UpdateWithConflictCheck(ctx, match); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("match not found")
		}
		if errors.Is(err, repository.ErrConflict) {
			return nil, apperror.Conflict(scheduleConflictMessage)
		}
		return nil, apperror.Internal(err)
	}

	updated, err := s.matches.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := matchToResponse(*updated)
	return &resp, nil
}

func (s *MatchService) Delete(ctx context.Context, id int64, userID int64) error {
	existing, err := s.matches.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("match not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	if existing.Status == model.MatchStatusPlayed {
		return apperror.Conflict("cannot delete a match that has already been played")
	}

	err = s.matches.SoftDelete(ctx, id, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("match not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *MatchService) requireTeam(ctx context.Context, teamID int64) error {
	_, err := s.teams.FindByID(ctx, teamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("team not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func matchToResponse(m model.Match) dto.MatchResponse {
	return dto.MatchResponse{
		ID:            m.ID,
		HomeTeamID:    m.HomeTeamID,
		AwayTeamID:    m.AwayTeamID,
		MatchDatetime: m.MatchDatetime,
		HomeScore:     m.HomeScore,
		AwayScore:     m.AwayScore,
		Status:        string(m.Status),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
