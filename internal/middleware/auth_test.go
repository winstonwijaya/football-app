package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"football-app/internal/middleware"
	"football-app/pkg/token"
)

func newTestRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	protected := r.Group("/protected")
	protected.Use(middleware.RequireAuth(secret))
	protected.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetInt64(middleware.ContextUserIDKey)})
	})
	return r
}

func TestRequireAuth(t *testing.T) {
	secret := "test-secret"
	router := newTestRouter(secret)

	validToken, _, err := token.Generate(secret, 42, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	expiredToken, _, err := token.Generate(secret, 42, -time.Hour)
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}
	wrongSecretToken, _, err := token.Generate("other-secret", 42, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"missing header", "", http.StatusUnauthorized},
		{"malformed header", "Token abc123", http.StatusUnauthorized},
		{"invalid token", "Bearer not-a-real-token", http.StatusUnauthorized},
		{"expired token", "Bearer " + expiredToken, http.StatusUnauthorized},
		{"wrong secret", "Bearer " + wrongSecretToken, http.StatusUnauthorized},
		{"valid token", "Bearer " + validToken, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
