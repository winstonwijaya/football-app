package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"football-app/internal/middleware"
	"football-app/pkg/apperror"
)

// bindJSON binds and validates the request body, returning a ready-to-use
// *apperror.Error with field-level messages on failure so handlers never
// touch gin's raw binding errors.
func bindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return toValidationError(c, err)
	}
	return nil
}

// bindQuery binds and validates query-string parameters the same way
// bindJSON does for the body.
func bindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return toValidationError(c, err)
	}
	return nil
}

func toValidationError(c *gin.Context, err error) error {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fe.Field()] = fieldErrorMessage(fe)
		}
		return apperror.Validation("validation failed", fields)
	}

	// Not a field-level rule failure — malformed JSON, wrong type for a
	// field, empty body, etc. These never reached field validation, so
	// there's no fields map to build; log the real cause server-side
	// (the client only gets a message, per apperror.Internal's pattern —
	// except here the message is genuinely safe to share, since it just
	// describes what's wrong with the client's own input).
	log.Printf("[%s] malformed request body: %v", middleware.RequestIDFrom(c), err)
	return apperror.Validation(jsonErrorMessage(err), nil)
}

// jsonErrorMessage turns encoding/json's error types into a message safe
// and useful to return to the client — these describe malformed input, not
// internal state, so unlike apperror.Internal they aren't hidden.
func jsonErrorMessage(err error) string {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return fmt.Sprintf("malformed JSON in request body (at byte offset %d) — note JSON does not support // comments", syntaxErr.Offset)
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return fmt.Sprintf("field %q must be of type %s", typeErr.Field, typeErr.Type)
	}
	if errors.Is(err, io.EOF) {
		return "request body is empty"
	}
	return "invalid request body"
}

func fieldErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "max":
		if isNumericKind(fe.Kind()) {
			return fmt.Sprintf("must be at most %s", fe.Param())
		}
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "min":
		if isNumericKind(fe.Kind()) {
			return fmt.Sprintf("must be at least %s", fe.Param())
		}
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("must be >= %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be <= %s", fe.Param())
	case "url":
		return "must be a valid URL"
	case "nefield":
		return fmt.Sprintf("must be different from %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", strings.Join(splitOneOfParam(fe.Param()), ", "))
	case "datetime":
		return fmt.Sprintf("must match the date format %s", humanDateFormat(fe.Param()))
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}

// splitOneOfParam splits a validator "oneof" tag param on whitespace,
// respecting single-quoted multi-word values (e.g. "a b 'c d' e" ->
// ["a", "b", "c d", "e"]), matching validator's own oneof quoting syntax.
func splitOneOfParam(param string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range param {
		switch {
		case r == '\'':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// humanDateFormat translates Go's reference-date layout (e.g. "2006-01-02")
// into a conventional placeholder (e.g. "YYYY-MM-DD"), so validation
// messages don't expose a Go-specific idiom to non-Go API consumers.
func humanDateFormat(goLayout string) string {
	replacer := strings.NewReplacer(
		"2006", "YYYY",
		"01", "MM",
		"02", "DD",
		"15", "hh",
		"04", "mm",
		"05", "ss",
	)
	return replacer.Replace(goLayout)
}

func isNumericKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
