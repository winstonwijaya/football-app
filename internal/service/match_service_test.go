package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"football-app/internal/dto"
	"football-app/internal/model"
	"football-app/internal/repository"
	"football-app/internal/service"
	"football-app/pkg/apperror"
)

// fakePlayerRepository is an in-memory stand-in for repository.PlayerRepository.
type fakePlayerRepository struct {
	playersByID map[int64]*model.Player
	nextID      int64
}

func newFakePlayerRepository() *fakePlayerRepository {
	return &fakePlayerRepository{playersByID: map[int64]*model.Player{}, nextID: 1}
}

func (f *fakePlayerRepository) Create(_ context.Context, p *model.Player) error {
	p.ID = f.nextID
	f.nextID++
	f.playersByID[p.ID] = p
	return nil
}

func (f *fakePlayerRepository) FindByID(_ context.Context, id int64) (*model.Player, error) {
	p, ok := f.playersByID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *p
	return &cp, nil
}

func (f *fakePlayerRepository) List(_ context.Context, _ repository.PlayerFilter) ([]model.Player, int64, error) {
	players := make([]model.Player, 0, len(f.playersByID))
	for _, p := range f.playersByID {
		players = append(players, *p)
	}
	return players, int64(len(players)), nil
}

func (f *fakePlayerRepository) Update(_ context.Context, p *model.Player) error {
	if _, ok := f.playersByID[p.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.playersByID[p.ID] = p
	return nil
}

func (f *fakePlayerRepository) SoftDelete(_ context.Context, id int64, _ int64) error {
	if _, ok := f.playersByID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.playersByID, id)
	return nil
}

// fakeMatchRepository is an in-memory stand-in for repository.MatchRepository.
type fakeMatchRepository struct {
	matchesByID map[int64]*model.Match
	logs        []model.MatchLog
	nextID      int64
}

func newFakeMatchRepository() *fakeMatchRepository {
	return &fakeMatchRepository{matchesByID: map[int64]*model.Match{}, nextID: 1}
}

func (f *fakeMatchRepository) CreateWithConflictCheck(_ context.Context, match *model.Match) error {
	match.ID = f.nextID
	f.nextID++
	f.matchesByID[match.ID] = match
	return nil
}

func (f *fakeMatchRepository) FindByID(_ context.Context, id int64) (*model.Match, error) {
	m, ok := f.matchesByID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	// Return a copy, like a real DB fetch would — callers mutating the
	// returned struct must not leak into the fake's stored state.
	cp := *m
	return &cp, nil
}

func (f *fakeMatchRepository) List(_ context.Context, _ repository.MatchFilter) ([]model.Match, int64, error) {
	matches := make([]model.Match, 0, len(f.matchesByID))
	for _, m := range f.matchesByID {
		matches = append(matches, *m)
	}
	return matches, int64(len(matches)), nil
}

func (f *fakeMatchRepository) UpdateWithConflictCheck(_ context.Context, match *model.Match) error {
	if _, ok := f.matchesByID[match.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	f.matchesByID[match.ID] = match
	return nil
}

func (f *fakeMatchRepository) SoftDelete(_ context.Context, id int64, _ int64) error {
	if _, ok := f.matchesByID[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.matchesByID, id)
	return nil
}

func (f *fakeMatchRepository) ReportResult(_ context.Context, match *model.Match, logs []model.MatchLog, userID int64) error {
	current, ok := f.matchesByID[match.ID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if current.Status == model.MatchStatusPlayed {
		return repository.ErrConflict
	}
	for i := range logs {
		logs[i].CreatedBy = &userID
	}
	f.matchesByID[match.ID] = match
	f.logs = append(f.logs, logs...)
	return nil
}

// setupResultTest builds a MatchService with a scheduled match between two
// teams, one home player (id 1) and one away player (id 2).
func setupResultTest(t *testing.T) (*service.MatchService, *fakeMatchRepository, int64) {
	t.Helper()

	teams := newFakeTeamRepository()
	homeTeam := &model.Team{Name: "Home FC"}
	awayTeam := &model.Team{Name: "Away FC"}
	if err := teams.Create(context.Background(), homeTeam); err != nil {
		t.Fatalf("failed to create home team: %v", err)
	}
	if err := teams.Create(context.Background(), awayTeam); err != nil {
		t.Fatalf("failed to create away team: %v", err)
	}

	players := newFakePlayerRepository()
	homePlayer := &model.Player{TeamID: homeTeam.ID, Name: "Home Player", Position: model.PositionForward, SquadNumber: 9}
	awayPlayer := &model.Player{TeamID: awayTeam.ID, Name: "Away Player", Position: model.PositionForward, SquadNumber: 9}
	if err := players.Create(context.Background(), homePlayer); err != nil {
		t.Fatalf("failed to create home player: %v", err)
	}
	if err := players.Create(context.Background(), awayPlayer); err != nil {
		t.Fatalf("failed to create away player: %v", err)
	}

	matches := newFakeMatchRepository()
	match := &model.Match{
		HomeTeamID:    homeTeam.ID,
		AwayTeamID:    awayTeam.ID,
		MatchDatetime: time.Now(),
		Status:        model.MatchStatusScheduled,
	}
	if err := matches.CreateWithConflictCheck(context.Background(), match); err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	svc := service.NewMatchService(matches, teams, players)
	return svc, matches, match.ID
}

func TestMatchService_ReportResult(t *testing.T) {
	tests := []struct {
		name       string
		buildReq   func(homePlayerID, awayPlayerID int64, homeTeamID, awayTeamID int64) dto.ReportResultRequest
		wantErr    bool
		wantCode   apperror.Code
		wantFields []string // field keys expected in the validation error, if any
	}{
		{
			name: "score matches goal count - succeeds",
			buildReq: func(homePlayerID, awayPlayerID, homeTeamID, awayTeamID int64) dto.ReportResultRequest {
				return dto.ReportResultRequest{
					HomeScore: 1,
					AwayScore: 1,
					Goals: []dto.GoalInput{
						{TeamID: homeTeamID, PlayerID: homePlayerID, Minute: 10},
						{TeamID: awayTeamID, PlayerID: awayPlayerID, Minute: 20},
					},
				}
			},
			wantErr: false,
		},
		{
			name: "own goal credited to opposing team - succeeds",
			buildReq: func(homePlayerID, awayPlayerID, homeTeamID, awayTeamID int64) dto.ReportResultRequest {
				return dto.ReportResultRequest{
					HomeScore: 1,
					AwayScore: 0,
					Goals: []dto.GoalInput{
						// away player scores an own goal, credited to home team
						{TeamID: homeTeamID, PlayerID: awayPlayerID, Minute: 30, IsOwnGoal: true},
					},
				}
			},
			wantErr: false,
		},
		{
			name: "goal count does not match stated score",
			buildReq: func(homePlayerID, awayPlayerID, homeTeamID, awayTeamID int64) dto.ReportResultRequest {
				return dto.ReportResultRequest{
					HomeScore: 2, // only one home goal submitted
					AwayScore: 0,
					Goals: []dto.GoalInput{
						{TeamID: homeTeamID, PlayerID: homePlayerID, Minute: 10},
					},
				}
			},
			wantErr:    true,
			wantCode:   apperror.CodeValidation,
			wantFields: []string{"home_score"},
		},
		{
			name: "goal team_id not part of this match",
			buildReq: func(homePlayerID, awayPlayerID, homeTeamID, awayTeamID int64) dto.ReportResultRequest {
				return dto.ReportResultRequest{
					HomeScore: 1,
					AwayScore: 0,
					Goals: []dto.GoalInput{
						{TeamID: 999, PlayerID: homePlayerID, Minute: 10},
					},
				}
			},
			wantErr:    true,
			wantCode:   apperror.CodeValidation,
			wantFields: []string{"goals[0].team_id"},
		},
		{
			name: "non-own-goal scored by player from the other team",
			buildReq: func(homePlayerID, awayPlayerID, homeTeamID, awayTeamID int64) dto.ReportResultRequest {
				return dto.ReportResultRequest{
					HomeScore: 1,
					AwayScore: 0,
					// away player credited to home team, but not flagged as an own goal
					Goals: []dto.GoalInput{
						{TeamID: homeTeamID, PlayerID: awayPlayerID, Minute: 10, IsOwnGoal: false},
					},
				}
			},
			wantErr:    true,
			wantCode:   apperror.CodeValidation,
			wantFields: []string{"goals[0].player_id"},
		},
		{
			name: "unknown player_id",
			buildReq: func(homePlayerID, awayPlayerID, homeTeamID, awayTeamID int64) dto.ReportResultRequest {
				return dto.ReportResultRequest{
					HomeScore: 1,
					AwayScore: 0,
					Goals: []dto.GoalInput{
						{TeamID: homeTeamID, PlayerID: 999, Minute: 10},
					},
				}
			},
			wantErr:    true,
			wantCode:   apperror.CodeValidation,
			wantFields: []string{"goals[0].player_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, matches, matchID := setupResultTest(t)
			match := matches.matchesByID[matchID]
			req := tt.buildReq(1, 2, match.HomeTeamID, match.AwayTeamID)

			resp, err := svc.ReportResult(context.Background(), matchID, req, 1)

			if tt.wantErr {
				var appErr *apperror.Error
				if !errors.As(err, &appErr) {
					t.Fatalf("expected an *apperror.Error, got %T (%v)", err, err)
				}
				if appErr.Code != tt.wantCode {
					t.Errorf("got code %s, want %s", appErr.Code, tt.wantCode)
				}
				for _, key := range tt.wantFields {
					if _, ok := appErr.Fields[key]; !ok {
						t.Errorf("expected field %q in validation errors, got %v", key, appErr.Fields)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Status != string(model.MatchStatusPlayed) {
				t.Errorf("got status %q, want %q", resp.Status, model.MatchStatusPlayed)
			}
		})
	}
}

func TestMatchService_ReportResult_AlreadyReported(t *testing.T) {
	svc, _, matchID := setupResultTest(t)

	req := dto.ReportResultRequest{
		HomeScore: 0,
		AwayScore: 0,
	}

	if _, err := svc.ReportResult(context.Background(), matchID, req, 1); err != nil {
		t.Fatalf("unexpected error on first report: %v", err)
	}

	_, err := svc.ReportResult(context.Background(), matchID, req, 1)
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperror.Error, got %T (%v)", err, err)
	}
	if appErr.Code != apperror.CodeConflict {
		t.Errorf("got code %s, want %s", appErr.Code, apperror.CodeConflict)
	}
}
