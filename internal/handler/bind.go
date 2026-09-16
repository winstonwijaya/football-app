package handler

import (
	"errors"
	"fmt"
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
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be >= %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be <= %s", fe.Param())
	case "url":
		return "must be a valid URL"
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
