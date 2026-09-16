package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/repository"
	"football-app/pkg/apperror"
)

type ReportService struct {
	reports repository.ReportRepository
}

func NewReportService(reports repository.ReportRepository) *ReportService {
	return &ReportService{reports: reports}
}

func (s *ReportService) MatchReport(ctx context.Context, matchID int64) (*dto.MatchReportResponse, error) {
	row, err := s.reports.MatchReport(ctx, matchID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.NotFound("match report not available")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}

	resp := reportRowToResponse(*row)
	return &resp, nil
}

func (s *ReportService) ListMatchReports(ctx context.Context, query dto.ListMatchReportsQuery) ([]dto.MatchReportResponse, int64, error) {
	rows, total, err := s.reports.ListMatchReports(ctx, query.Page, query.Limit)
	if err != nil {
		return nil, 0, apperror.Internal(err)
	}

	items := make([]dto.MatchReportResponse, len(rows))
	for i, row := range rows {
		items[i] = reportRowToResponse(row)
	}
	return items, total, nil
}

func reportRowToResponse(row repository.MatchReportRow) dto.MatchReportResponse {
	outcome := dto.OutcomeDraw
	if row.HomeScore > row.AwayScore {
		outcome = dto.OutcomeHomeWin
	} else if row.AwayScore > row.HomeScore {
		outcome = dto.OutcomeAwayWin
	}

	var topScorer *dto.TopScorer
	if row.TopScorerPlayerID != nil {
		topScorer = &dto.TopScorer{
			PlayerID: *row.TopScorerPlayerID,
			Name:     *row.TopScorerName,
			Goals:    *row.TopScorerGoals,
		}
	}

	return dto.MatchReportResponse{
		MatchID:                row.MatchID,
		MatchDatetime:          row.MatchDatetime,
		HomeTeam:               dto.TeamRef{ID: row.HomeTeamID, Name: row.HomeTeamName},
		AwayTeam:               dto.TeamRef{ID: row.AwayTeamID, Name: row.AwayTeamName},
		HomeScore:              row.HomeScore,
		AwayScore:              row.AwayScore,
		Outcome:                outcome,
		TopScorer:              topScorer,
		HomeTeamCumulativeWins: row.HomeCumulativeWins,
		AwayTeamCumulativeWins: row.AwayCumulativeWins,
	}
}
