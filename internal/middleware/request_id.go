package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	HeaderRequestID     = "X-Request-ID"
	ContextRequestIDKey = "request_id"
)

// RequestID assigns a unique ID to every request — reused from an incoming
// X-Request-ID header if the caller supplied one, otherwise generated.
// Echoed back as a response header so a client can correlate its own logs
// with server-side ones, and stashed in the context so log lines emitted
// while handling this request can be tagged with it.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ContextRequestIDKey, id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}

// RequestIDFrom reads the request ID set by RequestID, for use in log lines
// emitted elsewhere in the middleware chain.
func RequestIDFrom(c *gin.Context) string {
	if id, ok := c.Get(ContextRequestIDKey); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}
