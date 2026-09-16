package dto

import "time"

// CreateTeamRequest — POST /api/v1/teams
type CreateTeamRequest struct {
	Name            string  `json:"name" binding:"required,max=150"`
	LogoURL         *string `json:"logo_url" binding:"omitempty,url"`
	EstablishedYear *int16  `json:"established_year" binding:"omitempty,gte=1850,lte=2100"`
	Address         *string `json:"address" binding:"omitempty,max=255"`
	City            *string `json:"city" binding:"omitempty,max=100"`
}

// UpdateTeamRequest — PUT /api/v1/teams/:id.
// Full-replace semantics: every field is written, so omitted optional fields clear the column.
type UpdateTeamRequest struct {
	Name            string  `json:"name" binding:"required,max=150"`
	LogoURL         *string `json:"logo_url" binding:"omitempty,url"`
	EstablishedYear *int16  `json:"established_year" binding:"omitempty,gte=1850,lte=2100"`
	Address         *string `json:"address" binding:"omitempty,max=255"`
	City            *string `json:"city" binding:"omitempty,max=100"`
}

// ListTeamsQuery — GET /api/v1/teams?city=&search=&page=&limit=
type ListTeamsQuery struct {
	City   string `form:"city"`
	Search string `form:"search"`
	PaginationQuery
}

type TeamResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	LogoURL         *string   `json:"logo_url,omitempty"`
	EstablishedYear *int16    `json:"established_year,omitempty"`
	Address         *string   `json:"address,omitempty"`
	City            *string   `json:"city,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
