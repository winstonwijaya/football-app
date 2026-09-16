package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/repository"
	"football-app/internal/service"
	"football-app/pkg/apperror"
)

// fakeReportRepository is an in-memory stand-in for repository.ReportRepository.
type fakeReportRepository struct {
	rowsByMatchID map[int64]repository.MatchReportRow
	order         []int64
}

func newFakeReportRepository(rows ...repository.MatchReportRow) *fakeReportRepository {
	f := &fakeReportRepository{rowsByMatchID: map[int64]repository.MatchReportRow{}}
	for _, r := range rows {
		f.rowsByMatchID[r.MatchID] = r
		f.order = append(f.order, r.MatchID)
	}
	return f
}

func (f *fakeReportRepository) MatchReport(_ context.Context, matchID int64) (*repository.MatchReportRow, error) {
	row, ok := f.rowsByMatchID[matchID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (f *fakeReportRepository) ListMatchReports(_ context.Context, page, limit int) ([]repository.MatchReportRow, int64, error) {
	rows := make([]repository.MatchReportRow, 0, len(f.order))
	for _, id := range f.order {
		rows = append(rows, f.rowsByMatchID[id])
	}
	return rows, int64(len(rows)), nil
}

func int64Ptr(v int64) *int64    { return &v }
func stringPtr(v string) *string { return &v }

func TestReportService_MatchReport_Outcome(t *testing.T) {
	tests := []struct {
		name        string
		homeScore   int16
		awayScore   int16
		wantOutcome string
	}{
		{"home win", 2, 1, dto.OutcomeHomeWin},
		{"away win", 0, 3, dto.OutcomeAwayWin},
		{"draw", 1, 1, dto.OutcomeDraw},
		{"scoreless draw", 0, 0, dto.OutcomeDraw},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeReportRepository(repository.MatchReportRow{
				MatchID:       1,
				MatchDatetime: time.Now(),
				HomeTeamID:    1,
				HomeTeamName:  "Home",
				AwayTeamID:    2,
				AwayTeamName:  "Away",
				HomeScore:     tt.homeScore,
				AwayScore:     tt.awayScore,
			})
			svc := service.NewReportService(repo)

			resp, err := svc.MatchReport(context.Background(), 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Outcome != tt.wantOutcome {
				t.Errorf("got outcome %q, want %q", resp.Outcome, tt.wantOutcome)
			}
		})
	}
}

func TestReportService_MatchReport_TopScorer(t *testing.T) {
	t.Run("present when goals were scored", func(t *testing.T) {
		repo := newFakeReportRepository(repository.MatchReportRow{
			MatchID:           1,
			HomeTeamID:        1,
			HomeTeamName:      "Home",
			AwayTeamID:        2,
			AwayTeamName:      "Away",
			HomeScore:         2,
			AwayScore:         0,
			TopScorerPlayerID: int64Ptr(7),
			TopScorerName:     stringPtr("Star Player"),
			TopScorerGoals:    int64Ptr(2),
		})
		svc := service.NewReportService(repo)

		resp, err := svc.MatchReport(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TopScorer == nil {
			t.Fatal("expected a top scorer, got nil")
		}
		if resp.TopScorer.PlayerID != 7 || resp.TopScorer.Goals != 2 {
			t.Errorf("got top scorer %+v, want player_id=7 goals=2", resp.TopScorer)
		}
	})

	t.Run("nil for a scoreless draw", func(t *testing.T) {
		repo := newFakeReportRepository(repository.MatchReportRow{
			MatchID:      1,
			HomeTeamID:   1,
			HomeTeamName: "Home",
			AwayTeamID:   2,
			AwayTeamName: "Away",
			HomeScore:    0,
			AwayScore:    0,
			// TopScorerPlayerID left nil
		})
		svc := service.NewReportService(repo)

		resp, err := svc.MatchReport(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.TopScorer != nil {
			t.Errorf("expected nil top scorer, got %+v", resp.TopScorer)
		}
	})
}

func TestReportService_MatchReport_NotFound(t *testing.T) {
	repo := newFakeReportRepository()
	svc := service.NewReportService(repo)

	_, err := svc.MatchReport(context.Background(), 999)
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Errorf("got code %s, want %s", appErr.Code, apperror.CodeNotFound)
	}
}

func TestReportService_ListMatchReports_CumulativeWins(t *testing.T) {
	repo := newFakeReportRepository(
		repository.MatchReportRow{
			MatchID: 1, HomeTeamID: 1, HomeTeamName: "A", AwayTeamID: 2, AwayTeamName: "B",
			HomeScore: 1, AwayScore: 0, HomeCumulativeWins: 1, AwayCumulativeWins: 0,
		},
		repository.MatchReportRow{
			MatchID: 2, HomeTeamID: 1, HomeTeamName: "A", AwayTeamID: 2, AwayTeamName: "B",
			HomeScore: 0, AwayScore: 2, HomeCumulativeWins: 1, AwayCumulativeWins: 1,
		},
	)
	svc := service.NewReportService(repo)

	items, total, err := svc.ListMatchReports(context.Background(), dto.ListMatchReportsQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("got total %d, want 2", total)
	}
	if items[0].HomeTeamCumulativeWins != 1 || items[0].AwayTeamCumulativeWins != 0 {
		t.Errorf("match 1: got wins home=%d away=%d, want home=1 away=0", items[0].HomeTeamCumulativeWins, items[0].AwayTeamCumulativeWins)
	}
	if items[1].HomeTeamCumulativeWins != 1 || items[1].AwayTeamCumulativeWins != 1 {
		t.Errorf("match 2: got wins home=%d away=%d, want home=1 away=1", items[1].HomeTeamCumulativeWins, items[1].AwayTeamCumulativeWins)
	}
}
