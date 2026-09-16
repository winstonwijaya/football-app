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

type PlayerService struct {
	players repository.PlayerRepository
	teams   repository.TeamRepository
}

func NewPlayerService(players repository.PlayerRepository, teams repository.TeamRepository) *PlayerService {
	return &PlayerService{players: players, teams: teams}
}

func (s *PlayerService) Create(ctx context.Context, req dto.CreatePlayerRequest, userID int64) (*dto.PlayerResponse, error) {
	if err := s.requireTeam(ctx, req.TeamID); err != nil {
		return nil, err
	}

	player := &model.Player{
		TeamID:      req.TeamID,
		Name:        req.Name,
		HeightCM:    req.HeightCM,
		WeightKG:    req.WeightKG,
		Position:    model.PlayerPosition(req.Position),
		SquadNumber: req.SquadNumber,
	}
	player.CreatedBy = &userID
	player.UpdatedBy = &userID

	if err := s.players.Create(ctx, player); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, apperror.Conflict("squad number already in use for this team")
		}
		return nil, apperror.Internal(err)
	}

	resp := playerToResponse(*player)
	return &resp, nil
}

func (s *PlayerService) Get(ctx context.Context, id int64) (*dto.PlayerResponse, error) {
	player, err := s.players.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NotFound("player not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := playerToResponse(*player)
	return &resp, nil
}

func (s *PlayerService) List(ctx context.Context, query dto.ListPlayersQuery) ([]dto.PlayerResponse, int64, error) {
	players, total, err := s.players.List(ctx, repository.PlayerFilter{
		TeamID:   query.TeamID,
		Position: query.Position,
		Page:     query.Page,
		Limit:    query.Limit,
	})
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}

	items := make([]dto.PlayerResponse, len(players))
	for i, p := range players {
		items[i] = playerToResponse(p)
	}
	return items, total, nil
}

func (s *PlayerService) Update(ctx context.Context, id int64, req dto.UpdatePlayerRequest, userID int64) (*dto.PlayerResponse, error) {
	if err := s.requireTeam(ctx, req.TeamID); err != nil {
		return nil, err
	}

	player := &model.Player{
		ID:          id,
		TeamID:      req.TeamID,
		Name:        req.Name,
		HeightCM:    req.HeightCM,
		WeightKG:    req.WeightKG,
		Position:    model.PlayerPosition(req.Position),
		SquadNumber: req.SquadNumber,
	}
	player.UpdatedBy = &userID

	if err := s.players.Update(ctx, player); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("player not found")
		}
		if errors.Is(err, repository.ErrConflict) {
			return nil, apperror.Conflict("squad number already in use for this team")
		}
		return nil, apperror.Internal(err)
	}

	updated, err := s.players.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := playerToResponse(*updated)
	return &resp, nil
}

func (s *PlayerService) Delete(ctx context.Context, id int64, userID int64) error {
	err := s.players.SoftDelete(ctx, id, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("player not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *PlayerService) requireTeam(ctx context.Context, teamID int64) error {
	_, err := s.teams.FindByID(ctx, teamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFound("team not found")
	}
	if err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func playerToResponse(p model.Player) dto.PlayerResponse {
	return dto.PlayerResponse{
		ID:          p.ID,
		TeamID:      p.TeamID,
		Name:        p.Name,
		HeightCM:    p.HeightCM,
		WeightKG:    p.WeightKG,
		Position:    string(p.Position),
		SquadNumber: p.SquadNumber,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
