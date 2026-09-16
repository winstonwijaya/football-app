package dto

import "time"

// CreateMatchRequest — POST /api/v1/matches
type CreateMatchRequest struct {
	HomeTeamID    int64     `json:"home_team_id" binding:"required"`
	AwayTeamID    int64     `json:"away_team_id" binding:"required,nefield=HomeTeamID"`
	MatchDatetime time.Time `json:"match_datetime" binding:"required"`
}

// UpdateMatchRequest — PUT /api/v1/matches/:id. Full-replace semantics for
// schedule fields. Status may only move between Scheduled/Cancelled here —
// transitioning to Played happens exclusively via POST /matches/:id/result,
// which also writes match_logs and reconciles the score.
type UpdateMatchRequest struct {
	HomeTeamID    int64     `json:"home_team_id" binding:"required"`
	AwayTeamID    int64     `json:"away_team_id" binding:"required,nefield=HomeTeamID"`
	MatchDatetime time.Time `json:"match_datetime" binding:"required"`
	Status        string    `json:"status" binding:"required,oneof=Scheduled Cancelled"`
}

// ListMatchesQuery — GET /api/v1/matches?status=&team_id=&from=&to=&page=&limit=
type ListMatchesQuery struct {
	Status string     `form:"status" binding:"omitempty,oneof=Scheduled Played Cancelled"`
	TeamID *int64     `form:"team_id"`
	From   *time.Time `form:"from" time_format:"2006-01-02"`
	To     *time.Time `form:"to" time_format:"2006-01-02"`
	PaginationQuery
}

type MatchResponse struct {
	ID            int64     `json:"id"`
	HomeTeamID    int64     `json:"home_team_id"`
	AwayTeamID    int64     `json:"away_team_id"`
	MatchDatetime time.Time `json:"match_datetime"`
	HomeScore     *int16    `json:"home_score,omitempty"`
	AwayScore     *int16    `json:"away_score,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
