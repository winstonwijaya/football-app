package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"football-app/pkg/apperror"
	"football-app/pkg/token"
)

// ContextUserIDKey is the gin.Context key RequireAuth stores the
// authenticated user's ID under.
const ContextUserIDKey = "user_id"

// RequireAuth validates the Bearer token on every route it wraps and
// stores the caller's user ID in the request context, for handlers/services
// to use as created_by/updated_by on mutations.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.Error(apperror.Unauthorized("missing bearer token"))
			c.Abort()
			return
		}

		claims, err := token.Parse(jwtSecret, strings.TrimPrefix(header, prefix))
		if err != nil {
			c.Error(apperror.Unauthorized("invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Next()
	}
}
