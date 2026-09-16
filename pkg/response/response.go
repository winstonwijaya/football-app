// Package response defines the single JSON envelope used by every handler,
// success and error alike, so API consumers only ever parse one shape.
package response

import "github.com/gin-gonic/gin"

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type Meta struct {
	Page  int   `json:"page,omitempty"`
	Limit int   `json:"limit,omitempty"`
	Total int64 `json:"total,omitempty"`
}

// OK writes a successful envelope with no pagination metadata.
func OK(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Envelope{Success: true, Data: data})
}

// OKWithMeta writes a successful envelope including pagination metadata,
// for list endpoints.
func OKWithMeta(c *gin.Context, status int, data interface{}, meta Meta) {
	c.JSON(status, Envelope{Success: true, Data: data, Meta: &meta})
}

// Error writes a failed envelope. Handlers should not call this directly —
// return a *apperror.Error from the service layer and let
// middleware.ErrorHandler call this instead, so error formatting stays in
// one place.
func Error(c *gin.Context, status int, code, message string, fields map[string]string) {
	c.JSON(status, Envelope{Success: false, Error: &ErrorBody{Code: code, Message: message, Fields: fields}})
}
