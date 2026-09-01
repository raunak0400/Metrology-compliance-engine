// Centralized error handling. Produces the consistent error envelope:
//
//	{ success: false, error: { code, message, details, requestId } }
//
// Stack traces / internal messages are never exposed in production.
package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/raunak0400/metrology-backend/internal/apperror"
)

type errorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Details   []apperror.Detail `json:"details,omitempty"`
	RequestID string            `json:"requestId,omitempty"`
}

// mapError converts any error (validation, pgx, or an already-typed AppError)
// into a normalized *apperror.AppError.
func mapError(err error) *apperror.AppError {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NotFound("Record")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return apperror.Conflict("A record with this value already exists (" + pgErr.ConstraintName + ")")
		case "23503": // foreign_key_violation
			return apperror.Conflict("Referenced record does not exist or is still in use")
		case "23514": // check_violation
			return apperror.BadRequest("Invalid value: " + pgErr.ConstraintName)
		default:
			return apperror.BadRequest("Database error (" + string(pgErr.Code) + ")")
		}
	}

	return apperror.Internal(err.Error())
}

// ErrorHandler must be registered last. It reads any error attached via
// c.Error(...) during the request and writes the standard envelope.
func ErrorHandler(isProd bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		appErr := mapError(err)
		requestID := c.GetString("requestId")

		if !appErr.IsOperational || appErr.StatusCode >= 500 {
			slog.Error("unhandled error", "err", err, "requestId", requestID)
		} else {
			slog.Warn("handled error", "code", appErr.ErrCode, "msg", appErr.Msg, "requestId", requestID)
		}

		msg := appErr.Msg
		if isProd && appErr.StatusCode >= 500 {
			msg = "Internal server error"
		}

		c.JSON(appErr.StatusCode, gin.H{
			"success": false,
			"error": errorBody{
				Code:      string(appErr.ErrCode),
				Message:   msg,
				Details:   appErr.Details,
				RequestID: requestID,
			},
		})
	}
}

// NotFoundHandler responds with the standard 404 envelope for unmatched routes.
func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error": errorBody{
			Code:    string(apperror.CodeNotFound),
			Message: "Route " + c.Request.Method + " " + c.Request.URL.Path + " not found",
		},
	})
}
