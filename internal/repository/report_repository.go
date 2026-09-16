package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// MatchReportRow is the flat result of the match-report query — one row
// covering the match, both teams' names, the match's top scorer (nullable:
// a 0-0 draw or an own-goal-only match has none), and each team's
// cumulative win count as of this match's date.
type MatchReportRow struct {
	MatchID            int64
	MatchDatetime      time.Time
	HomeTeamID         int64
	HomeTeamName       string
	AwayTeamID         int64
	AwayTeamName       string
	HomeScore          int16
	AwayScore          int16
	HomeCumulativeWins int64
	AwayCumulativeWins int64
	TopScorerPlayerID  *int64
	TopScorerName      *string
	TopScorerGoals     *int64
}

type ReportRepository interface {
	// MatchReport returns gorm.ErrRecordNotFound if the match doesn't
	// exist or hasn't been played yet (result not reported).
	MatchReport(ctx context.Context, matchID int64) (*MatchReportRow, error)
	// ListMatchReports returns one row per played match, newest first is
	// not implied — ordered by match_datetime ascending, same as the
	// window functions computing cumulative wins.
	ListMatchReports(ctx context.Context, page, limit int) ([]MatchReportRow, int64, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// matchReportCTE computes, for every played match: each team's cumulative
// win count up to and including that match's date (via a window function
// over a per-team-per-match "did this team win" row), and the match's top
// scorer (via ROW_NUMBER over goals scored per player, excluding own
// goals, tie-broken by lowest player_id for determinism).
const matchReportCTE = `
WITH team_results AS (
    SELECT id AS match_id, home_team_id AS team_id, match_datetime,
           (home_score > away_score) AS is_win
    FROM matches
    WHERE status = 'Played' AND deleted_at IS NULL
    UNION ALL
    SELECT id AS match_id, away_team_id AS team_id, match_datetime,
           (away_score > home_score) AS is_win
    FROM matches
    WHERE status = 'Played' AND deleted_at IS NULL
),
cumulative AS (
    SELECT match_id, team_id,
           SUM(CASE WHEN is_win THEN 1 ELSE 0 END) OVER (
               PARTITION BY team_id ORDER BY match_datetime, match_id
               ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
           ) AS cumulative_wins
    FROM team_results
),
goal_counts AS (
    SELECT match_id, player_id, COUNT(*) AS goals
    FROM match_logs
    WHERE action = 'GOAL' AND is_own_goal = false
    GROUP BY match_id, player_id
),
ranked_scorers AS (
    SELECT match_id, player_id, goals,
           ROW_NUMBER() OVER (PARTITION BY match_id ORDER BY goals DESC, player_id ASC) AS rn
    FROM goal_counts
)
SELECT
    m.id AS match_id,
    m.match_datetime,
    m.home_team_id, ht.name AS home_team_name,
    m.away_team_id, at.name AS away_team_name,
    m.home_score, m.away_score,
    ch.cumulative_wins AS home_cumulative_wins,
    ca.cumulative_wins AS away_cumulative_wins,
    rs.player_id AS top_scorer_player_id,
    p.name AS top_scorer_name,
    rs.goals AS top_scorer_goals
FROM matches m
JOIN teams ht ON ht.id = m.home_team_id
JOIN teams at ON at.id = m.away_team_id
JOIN cumulative ch ON ch.match_id = m.id AND ch.team_id = m.home_team_id
JOIN cumulative ca ON ca.match_id = m.id AND ca.team_id = m.away_team_id
LEFT JOIN ranked_scorers rs ON rs.match_id = m.id AND rs.rn = 1
LEFT JOIN players p ON p.id = rs.player_id
WHERE m.status = 'Played' AND m.deleted_at IS NULL
`

func (r *reportRepository) MatchReport(ctx context.Context, matchID int64) (*MatchReportRow, error) {
	var row MatchReportRow
	err := r.db.WithContext(ctx).
		Raw(matchReportCTE+" AND m.id = ?", matchID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.MatchID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *reportRepository) ListMatchReports(ctx context.Context, page, limit int) ([]MatchReportRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Table("matches").
		Where("status = ? AND deleted_at IS NULL", "Played").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	var rows []MatchReportRow
	err := r.db.WithContext(ctx).
		Raw(matchReportCTE+" ORDER BY m.match_datetime, m.id LIMIT ? OFFSET ?", limit, offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
