// Package apperror defines the application error hierarchy. Handlers and
// services return these; the centralized error middleware (internal/middleware)
// maps them to the consistent JSON error envelope.
package apperror

import "net/http"

type Code string

const (
	CodeValidation           Code = "VALIDATION_ERROR"
	CodeUnauthorized         Code = "UNAUTHORIZED"
	CodeForbidden            Code = "FORBIDDEN"
	CodeNotFound             Code = "NOT_FOUND"
	CodeConflict             Code = "CONFLICT"
	CodeRateLimited          Code = "RATE_LIMITED"
	CodeBadRequest           Code = "BAD_REQUEST"
	CodeUnsupportedMediaType Code = "UNSUPPORTED_MEDIA_TYPE"
	CodePayloadTooLarge      Code = "PAYLOAD_TOO_LARGE"
	CodeInternal             Code = "INTERNAL_ERROR"
	CodeServiceUnavailable   Code = "SERVICE_UNAVAILABLE"
)

// Detail describes one field-level validation failure.
type Detail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// AppError is the error type every handler/service should return for
// operational (expected) failures. IsOperational=false marks unexpected bugs
// whose messages should never reach clients in production.
type AppError struct {
	StatusCode    int
	ErrCode       Code
	Msg           string
	Details       []Detail
	IsOperational bool
}

func (e *AppError) Error() string { return e.Msg }

func New(status int, code Code, msg string, details ...Detail) *AppError {
	return &AppError{StatusCode: status, ErrCode: code, Msg: msg, Details: details, IsOperational: true}
}

func Validation(msg string, details ...Detail) *AppError {
	if msg == "" {
		msg = "Validation failed"
	}
	return New(http.StatusUnprocessableEntity, CodeValidation, msg, details...)
}

func BadRequest(msg string) *AppError {
	if msg == "" {
		msg = "Bad request"
	}
	return New(http.StatusBadRequest, CodeBadRequest, msg)
}

func Unauthorized(msg string) *AppError {
	if msg == "" {
		msg = "Authentication required"
	}
	return New(http.StatusUnauthorized, CodeUnauthorized, msg)
}

func Forbidden(msg string) *AppError {
	if msg == "" {
		msg = "You do not have permission to perform this action"
	}
	return New(http.StatusForbidden, CodeForbidden, msg)
}

func NotFound(resource string) *AppError {
	if resource == "" {
		resource = "Resource"
	}
	return New(http.StatusNotFound, CodeNotFound, resource+" not found")
}

func Conflict(msg string) *AppError {
	if msg == "" {
		msg = "Resource already exists"
	}
	return New(http.StatusConflict, CodeConflict, msg)
}

func UnsupportedMediaType(msg string) *AppError {
	if msg == "" {
		msg = "Unsupported media type"
	}
	return New(http.StatusUnsupportedMediaType, CodeUnsupportedMediaType, msg)
}

func PayloadTooLarge(msg string) *AppError {
	if msg == "" {
		msg = "Payload too large"
	}
	return New(http.StatusRequestEntityTooLarge, CodePayloadTooLarge, msg)
}

func ServiceUnavailable(msg string) *AppError {
	if msg == "" {
		msg = "Upstream service unavailable"
	}
	return New(http.StatusServiceUnavailable, CodeServiceUnavailable, msg)
}

func Internal(msg string) *AppError {
	if msg == "" {
		msg = "Internal server error"
	}
	return &AppError{StatusCode: http.StatusInternalServerError, ErrCode: CodeInternal, Msg: msg, IsOperational: false}
}
