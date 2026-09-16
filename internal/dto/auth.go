package dto

import "time"

// LoginRequest — POST /api/v1/auth/login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	ExpiresAt   time.Time   `json:"expires_at"`
	User        UserSummary `json:"user"`
}

// UserSummary is the minimal user info returned alongside a token —
// never includes PasswordHash.
type UserSummary struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
