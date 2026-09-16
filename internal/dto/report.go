package dto

import "time"

const (
	OutcomeHomeWin = "HOME_WIN"
	OutcomeAwayWin = "AWAY_WIN"
	OutcomeDraw    = "DRAW"
)

type TeamRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TopScorer struct {
	PlayerID int64  `json:"player_id"`
	Name     string `json:"name"`
	Goals    int64  `json:"goals"`
}

// MatchReportResponse — GET /api/v1/reports/matches/:id and the elements of
// GET /api/v1/reports/matches. TopScorer is nil for a 0-0 draw or a match
// decided entirely by own goals. Cumulative wins are each team's win count
// as of this match's date, not their current overall total.
type MatchReportResponse struct {
	MatchID                int64      `json:"match_id"`
	MatchDatetime          time.Time  `json:"match_datetime"`
	HomeTeam               TeamRef    `json:"home_team"`
	AwayTeam               TeamRef    `json:"away_team"`
	HomeScore              int16      `json:"home_score"`
	AwayScore              int16      `json:"away_score"`
	Outcome                string     `json:"outcome"`
	TopScorer              *TopScorer `json:"top_scorer,omitempty"`
	HomeTeamCumulativeWins int64      `json:"home_team_cumulative_wins"`
	AwayTeamCumulativeWins int64      `json:"away_team_cumulative_wins"`
}

// ListMatchReportsQuery — GET /api/v1/reports/matches?page=&limit=
type ListMatchReportsQuery struct {
	PaginationQuery
}
