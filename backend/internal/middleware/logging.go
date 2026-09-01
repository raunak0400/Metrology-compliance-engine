// Structured, request-scoped logging: assigns/propagates x-request-id and logs
// method, path, status and duration for every request (matching pino-http).
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ready" {
			c.Next()
			return
		}

		requestID := c.GetHeader("x-request-id")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("requestId", requestID)
		c.Writer.Header().Set("x-request-id", requestID)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		userID := ""
		if u, ok := CurrentUser(c); ok {
			userID = u.ID
		}

		attrs := []any{
			"requestId", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"durationMs", duration.Milliseconds(),
			"userId", userID,
		}

		switch {
		case status >= 500:
			slog.Error("request", attrs...)
		case status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}
