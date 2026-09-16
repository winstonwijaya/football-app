package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"football-app/internal/middleware"
	"football-app/pkg/apperror"
)

// idParam parses the ":id" path parameter shared by every resource route.
func idParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, apperror.Validation("invalid id", map[string]string{"id": "must be a positive integer"})
	}
	return id, nil
}

// userID reads the authenticated user's ID set by middleware.RequireAuth,
// for use as created_by/updated_by/deleted_by on mutations.
func userID(c *gin.Context) int64 {
	return c.GetInt64(middleware.ContextUserIDKey)
}
