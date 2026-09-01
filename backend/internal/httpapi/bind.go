// Package httpapi holds small helpers shared by every module's Gin handlers:
// UUID path-param parsing and structured request-body binding errors.
package httpapi

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
)

// ParseUUIDParam reads and validates a UUID path parameter (e.g. ":id").
func ParseUUIDParam(c *gin.Context, name string) (uuid.UUID, error) {
	raw := c.Param(name)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, apperror.Validation("Invalid "+name, apperror.Detail{
			Field: name, Message: "must be a valid UUID",
		})
	}
	return id, nil
}

// BindJSON parses and validates the request body into dst, translating
// go-playground/validator errors into the standard VALIDATION_ERROR envelope.
func BindJSON(c *gin.Context, dst interface{}) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return translateBindErr(err)
	}
	return nil
}

// Bind parses and validates the request body using content-type-aware binding
// (JSON, form, or multipart). Use this instead of BindJSON on routes that
// accept an optional file upload alongside JSON-only fields.
func Bind(c *gin.Context, dst interface{}) error {
	if err := c.ShouldBind(dst); err != nil {
		return translateBindErr(err)
	}
	return nil
}

// BindQuery parses and validates query parameters into dst.
func BindQuery(c *gin.Context, dst interface{}) error {
	if err := c.ShouldBindQuery(dst); err != nil {
		return translateBindErr(err)
	}
	return nil
}

func translateBindErr(err error) error {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		details := make([]apperror.Detail, 0, len(verrs))
		for _, fe := range verrs {
			details = append(details, apperror.Detail{
				Field:   toCamel(fe.Field()),
				Message: humanizeTag(fe),
				Code:    fe.Tag(),
			})
		}
		return apperror.Validation("Validation failed", details...)
	}
	return apperror.Validation(err.Error())
}

func humanizeTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "uuid4", "uuid":
		return "must be a valid UUID"
	case "min":
		return "must be at least " + fe.Param() + " characters/items"
	case "max":
		return "must be at most " + fe.Param() + " characters/items"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "is invalid (" + fe.Tag() + ")"
	}
}

func toCamel(field string) string {
	if field == "" {
		return field
	}
	return strings.ToLower(field[:1]) + field[1:]
}
