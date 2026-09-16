// Package apperror defines the typed error the service layer returns.
// Handlers never construct HTTP status codes themselves, they propagate this error and middleware.ErrorHandler maps Code to a status.
package apperror

// Code identifies the class of failure. It doubles as the machine-readable
type Code string

const (
	CodeValidation      Code = "VALIDATION_ERROR"
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeTooManyRequests Code = "TOO_MANY_REQUESTS"
	CodeInternal        Code = "INTERNAL_ERROR"
)

type Error struct {
	Code    Code
	Message string
	// Fields carries field-level validation messages, e.g. {"name": "required"}.
	Fields map[string]string
	// Err is the wrapped underlying error (DB error, etc.); never
	// serialized to the client, kept for logging via errors.Unwrap.
	Err error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NotFound(message string) *Error {
	return &Error{Code: CodeNotFound, Message: message}
}

func Conflict(message string) *Error {
	return &Error{Code: CodeConflict, Message: message}
}

func Validation(message string, fields map[string]string) *Error {
	return &Error{Code: CodeValidation, Message: message, Fields: fields}
}

func Unauthorized(message string) *Error {
	return &Error{Code: CodeUnauthorized, Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Code: CodeForbidden, Message: message}
}

func TooManyRequests(message string) *Error {
	return &Error{Code: CodeTooManyRequests, Message: message}
}

// Internal wraps an unexpected error (e.g. a DB failure). The wrapped error
// is logged but never sent to the client — Message is a generic string.
func Internal(err error) *Error {
	return &Error{Code: CodeInternal, Message: "internal server error", Err: err}
}
