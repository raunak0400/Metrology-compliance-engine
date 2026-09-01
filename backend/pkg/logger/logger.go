// Package logger provides structured, request-scoped logging built on the
// standard library's log/slog. JSON output in production; human-readable text
// in development.
package logger

import (
	"log/slog"
	"os"
)

// New builds the process-wide slog.Logger. level is one of debug/info/warn/error.
// In production it emits JSON; otherwise a readable text handler.
func New(level string, env string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With("service", "legal-metrology-api")
	slog.SetDefault(logger)
	return logger
}
