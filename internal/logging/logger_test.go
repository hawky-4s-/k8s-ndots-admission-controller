package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name       string
		level      string
		format     string
		logFunc    func(l *slog.Logger)
		wantLevel  slog.Level
		wantFormat string
	}{
		{
			name:       "default level and json format",
			level:      "",
			format:     "",
			logFunc:    func(l *slog.Logger) { l.Info("test message", "key", "value") },
			wantLevel:  slog.LevelInfo,
			wantFormat: "json",
		},
		{
			name:       "debug level",
			level:      "debug",
			format:     "json",
			logFunc:    func(l *slog.Logger) { l.Debug("test message", "key", "value") },
			wantLevel:  slog.LevelDebug,
			wantFormat: "json",
		},
		{
			name:       "warn level",
			level:      "warn",
			format:     "json",
			logFunc:    func(l *slog.Logger) { l.Warn("test message", "key", "value") },
			wantLevel:  slog.LevelWarn,
			wantFormat: "json",
		},
		{
			name:       "error level",
			level:      "error",
			format:     "json",
			logFunc:    func(l *slog.Logger) { l.Error("test message", "key", "value") },
			wantLevel:  slog.LevelError,
			wantFormat: "json",
		},
		{
			name:       "info level explicit",
			level:      "info",
			format:     "json",
			logFunc:    func(l *slog.Logger) { l.Info("test message", "key", "value") },
			wantLevel:  slog.LevelInfo,
			wantFormat: "json",
		},
		{
			name:       "text format",
			level:      "info",
			format:     "text",
			logFunc:    func(l *slog.Logger) { l.Info("test message", "key", "value") },
			wantLevel:  slog.LevelInfo,
			wantFormat: "text",
		},
		{
			name:       "case insensitive level",
			level:      "DEBUG",
			format:     "json",
			logFunc:    func(l *slog.Logger) { l.Debug("test message", "key", "value") },
			wantLevel:  slog.LevelDebug,
			wantFormat: "json",
		},
		{
			name:       "invalid level defaults to info",
			level:      "invalid",
			format:     "json",
			logFunc:    func(l *slog.Logger) { l.Info("test message", "key", "value") },
			wantLevel:  slog.LevelInfo,
			wantFormat: "json",
		},
		{
			name:       "invalid format defaults to json",
			level:      "info",
			format:     "invalid",
			logFunc:    func(l *slog.Logger) { l.Info("test message", "key", "value") },
			wantLevel:  slog.LevelInfo,
			wantFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.level, tt.format, &buf)

			require.NotNil(t, logger)

			// Log at the appropriate level for this test
			tt.logFunc(logger)

			output := buf.String()
			require.NotEmpty(t, output)

			if tt.wantFormat == "json" {
				var logEntry map[string]interface{}
				err := json.Unmarshal([]byte(output), &logEntry)
				require.NoError(t, err, "output should be valid JSON")
				assert.Equal(t, "test message", logEntry["msg"])
				assert.Equal(t, "value", logEntry["key"])
			} else {
				assert.True(t, strings.Contains(output, "test message"))
				assert.True(t, strings.Contains(output, "key=value"))
			}
		})
	}
}

func TestNewLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		logFunc   func(logger *slog.Logger)
		shouldLog bool
	}{
		{
			name:      "info level logs info",
			level:     "info",
			logFunc:   func(l *slog.Logger) { l.Info("test") },
			shouldLog: true,
		},
		{
			name:      "info level does not log debug",
			level:     "info",
			logFunc:   func(l *slog.Logger) { l.Debug("test") },
			shouldLog: false,
		},
		{
			name:      "debug level logs debug",
			level:     "debug",
			logFunc:   func(l *slog.Logger) { l.Debug("test") },
			shouldLog: true,
		},
		{
			name:      "error level does not log warn",
			level:     "error",
			logFunc:   func(l *slog.Logger) { l.Warn("test") },
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(tt.level, "json", &buf)

			tt.logFunc(logger)

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"", slog.LevelInfo},
		{"invalid", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
