package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"football-app/pkg/apperror"
	"football-app/pkg/response"
)

// ErrorHandler is the single place that turns a service-layer error into an
// HTTP status and the JSON error envelope. Handlers call c.Error(err) and
// return; they never call response.Error or set a status code themselves.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			if appErr.Err != nil {
				log.Printf("[%s] request error [%s]: %v (cause: %v)", RequestIDFrom(c), appErr.Code, appErr.Message, appErr.Err)
			}
			response.Error(c, statusForCode(appErr.Code), string(appErr.Code), appErr.Message, appErr.Fields)
			return
		}

		log.Printf("[%s] unhandled error: %v", RequestIDFrom(c), err)
		response.Error(c, http.StatusInternalServerError, string(apperror.CodeInternal), "internal server error", nil)
	}
}

func statusForCode(code apperror.Code) int {
	switch code {
	case apperror.CodeValidation:
		return http.StatusUnprocessableEntity
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
