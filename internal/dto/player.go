package dto

import "time"

// playerPositionOneOf must be kept in sync with the PlayerPosition consts
// in internal/model/player.go and the player_positions Postgres enum.
const playerPositionOneOf = "Forward Midfielder Defender Goalkeeper"

// CreatePlayerRequest — POST /api/v1/players
type CreatePlayerRequest struct {
	TeamID      int64    `json:"team_id" binding:"required"`
	Name        string   `json:"name" binding:"required,max=150"`
	HeightCM    *float64 `json:"height_cm" binding:"omitempty,gt=0,lte=300"`
	WeightKG    *float64 `json:"weight_kg" binding:"omitempty,gt=0,lte=250"`
	Position    string   `json:"position" binding:"required,oneof=Forward Midfielder Defender Goalkeeper"`
	SquadNumber int16    `json:"squad_number" binding:"required,gte=1,lte=99"`
}

// UpdatePlayerRequest — PUT /api/v1/players/:id. Full-replace semantics,
// including team_id: players can be transferred between teams. match_logs
// stores team_id explicitly (not derived from the player's current team)
// precisely so historical matches survive a transfer.
type UpdatePlayerRequest struct {
	TeamID      int64    `json:"team_id" binding:"required"`
	Name        string   `json:"name" binding:"required,max=150"`
	HeightCM    *float64 `json:"height_cm" binding:"omitempty,gt=0,lte=300"`
	WeightKG    *float64 `json:"weight_kg" binding:"omitempty,gt=0,lte=250"`
	Position    string   `json:"position" binding:"required,oneof=Forward Midfielder Defender Goalkeeper"`
	SquadNumber int16    `json:"squad_number" binding:"required,gte=1,lte=99"`
}

// ListPlayersQuery — GET /api/v1/players?team_id=&position=&page=&limit=
type ListPlayersQuery struct {
	TeamID   *int64 `form:"team_id"`
	Position string `form:"position" binding:"omitempty,oneof=Forward Midfielder Defender Goalkeeper"`
	PaginationQuery
}

type PlayerResponse struct {
	ID          int64     `json:"id"`
	TeamID      int64     `json:"team_id"`
	Name        string    `json:"name"`
	HeightCM    *float64  `json:"height_cm,omitempty"`
	WeightKG    *float64  `json:"weight_kg,omitempty"`
	Position    string    `json:"position"`
	SquadNumber int16     `json:"squad_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
