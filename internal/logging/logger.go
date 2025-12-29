// Package logging provides structured logging utilities.
package logging

import (
	"io"
	"log/slog"
	"strings"
)

// NewLogger creates a new slog.Logger with the specified level and format.
// Supported levels: debug, info, warn, error (case insensitive).
// Supported formats: json, text.
// Invalid values default to info level and json format.
func NewLogger(level, format string, w io.Writer) *slog.Logger {
	lvl := ParseLevel(level)

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	var handler slog.Handler
	if strings.ToLower(format) == "text" {
		handler = slog.NewTextHandler(w, opts)
	} else {
		handler = slog.NewJSONHandler(w, opts)
	}

	return slog.New(handler)
}

// ParseLevel parses a log level string into slog.Level.
// Supported values: debug, info, warn, warning, error (case insensitive).
// Invalid values default to info.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
