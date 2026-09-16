package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/model"
	"football-app/internal/repository"
	"football-app/pkg/apperror"
)

const (
	scheduleConflictMessage = "a team in this match already has a match scheduled on this date"
	resultConflictMessage   = "match result has already been reported"
)

type MatchService struct {
	matches repository.MatchRepository
	teams   repository.TeamRepository
	players repository.PlayerRepository
}

func NewMatchService(matches repository.MatchRepository, teams repository.TeamRepository, players repository.PlayerRepository) *MatchService {
	return &MatchService{matches: matches, teams: teams, players: players}
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

// ReportResult is a state transition, not a field update: it validates the
// submitted goals reconcile with the score, then writes the match's final
// score/status and every match_logs row in one transaction.
func (s *MatchService) ReportResult(ctx context.Context, matchID int64, req dto.ReportResultRequest, userID int64) (*dto.MatchResponse, error) {
	match, err := s.matches.FindByID(ctx, matchID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NotFound("match not found")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if match.Status == model.MatchStatusPlayed {
		return nil, apperror.Conflict(resultConflictMessage)
	}
	if match.Status == model.MatchStatusCancelled {
		return nil, apperror.Conflict("cannot report a result for a cancelled match")
	}

	logs, err := s.buildAndValidateGoals(ctx, match, req)
	if err != nil {
		return nil, err
	}

	match.HomeScore = &req.HomeScore
	match.AwayScore = &req.AwayScore
	match.Status = model.MatchStatusPlayed
	match.UpdatedBy = &userID

	if err := s.matches.ReportResult(ctx, match, logs, userID); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, apperror.Conflict(resultConflictMessage)
		}
		return nil, apperror.Internal(err)
	}

	resp := matchToResponse(*match)
	return &resp, nil
}

// buildAndValidateGoals checks every goal against the match (team_id must
// be home or away) and the scorer (must exist, and must belong to the
// credited team unless it's an own goal — then it must belong to the
// opposing team), and that the goal counts reconcile with the submitted
// score. All problems are collected into one apperror.Validation instead
// of failing on the first one.
func (s *MatchService) buildAndValidateGoals(ctx context.Context, match *model.Match, req dto.ReportResultRequest) ([]model.MatchLog, error) {
	fields := map[string]string{}
	homeCount, awayCount := 0, 0
	logs := make([]model.MatchLog, 0, len(req.Goals))

	for i, g := range req.Goals {
		prefix := fmt.Sprintf("goals[%d]", i)

		if g.TeamID != match.HomeTeamID && g.TeamID != match.AwayTeamID {
			fields[prefix+".team_id"] = "must be the home or away team of this match"
			continue
		}

		player, err := s.players.FindByID(ctx, g.PlayerID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fields[prefix+".player_id"] = "player not found"
			continue
		}
		if err != nil {
			return nil, apperror.Internal(err)
		}

		expectedPlayerTeam := g.TeamID
		if g.IsOwnGoal {
			expectedPlayerTeam = otherTeam(match, g.TeamID)
		}
		if player.TeamID != expectedPlayerTeam {
			if g.IsOwnGoal {
				fields[prefix+".player_id"] = "own goal must be scored by a player on the opposing team"
			} else {
				fields[prefix+".player_id"] = "player does not belong to the credited team"
			}
			continue
		}

		if g.TeamID == match.HomeTeamID {
			homeCount++
		} else {
			awayCount++
		}

		logs = append(logs, model.MatchLog{
			MatchID:   match.ID,
			TeamID:    g.TeamID,
			PlayerID:  g.PlayerID,
			Action:    model.MatchLogActionGoal,
			Minute:    g.Minute,
			IsOwnGoal: g.IsOwnGoal,
		})
	}

	if homeCount != int(req.HomeScore) {
		fields["home_score"] = fmt.Sprintf("does not match number of goals credited to the home team (%d)", homeCount)
	}
	if awayCount != int(req.AwayScore) {
		fields["away_score"] = fmt.Sprintf("does not match number of goals credited to the away team (%d)", awayCount)
	}

	if len(fields) > 0 {
		return nil, apperror.Validation("goals do not reconcile with the submitted score", fields)
	}
	return logs, nil
}

func otherTeam(match *model.Match, teamID int64) int64 {
	if teamID == match.HomeTeamID {
		return match.AwayTeamID
	}
	return match.HomeTeamID
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
