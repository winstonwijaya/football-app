package handler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"football-app/pkg/apperror"
)

// bindJSON binds and validates the request body, returning a ready-to-use
// *apperror.Error with field-level messages on failure so handlers never
// touch gin's raw binding errors.
func bindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return toValidationError(err)
	}
	return nil
}

// bindQuery binds and validates query-string parameters the same way
// bindJSON does for the body.
func bindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) error {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return apperror.Validation("invalid request body", nil)
	}

	fields := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		fields[fe.Field()] = fieldErrorMessage(fe)
	}
	return apperror.Validation("validation failed", fields)
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
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", strings.Join(splitOneOfParam(fe.Param()), ", "))
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
