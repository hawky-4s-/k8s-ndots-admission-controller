package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Clean env
	os.Clearenv()

	t.Run("defaults", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 8443, cfg.Port)
		assert.Equal(t, 2, cfg.NdotsValue)
		assert.Equal(t, "change-ndots", cfg.AnnotationKey)
		assert.Equal(t, "opt-out", cfg.AnnotationMode)
		assert.Len(t, cfg.NamespaceExclude, 3) // kube-system, kube-public, kube-node-lease
		assert.Equal(t, 10*time.Second, cfg.Timeout)
		// New fields
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, "json", cfg.LogFormat)
		assert.Equal(t, 8080, cfg.MetricsPort)
	})

	t.Run("from env", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		os.Setenv("NDOTS_VALUE", "5")
		os.Setenv("ANNOTATION_MODE", "opt-in")
		os.Setenv("NAMESPACE_INCLUDE", "prod,staging")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("LOG_FORMAT", "text")
		os.Setenv("METRICS_PORT", "9090")

		defer os.Clearenv()

		cfg, err := Load()
		require.NoError(t, err)
		assert.Equal(t, 9090, cfg.Port)
		assert.Equal(t, 5, cfg.NdotsValue)
		assert.Equal(t, "opt-in", cfg.AnnotationMode)
		assert.Equal(t, []string{"prod", "staging"}, cfg.NamespaceInclude)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, "text", cfg.LogFormat)
		assert.Equal(t, 9090, cfg.MetricsPort)
	})

	t.Run("bad env", func(t *testing.T) {
		os.Setenv("PORT", "invalid")
		defer os.Clearenv()

		cfg, err := Load()
		// If int parsing fails, we usually ignore or error. Plan didn't specify strict fail on parse, but Config.Load logic usually implies it behaves.
		// If Load ignores errors (as mostly using strconv.AtoI and ignoring err in snippet), then it keeps default.
		// Wait, snippet said "if port, err := strconv.Atoi(v); err == nil". So it ignores invalid ints.
		require.NoError(t, err)
		assert.Equal(t, 8443, cfg.Port) // Default remains
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cfg := DefaultConfig
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid port", func(t *testing.T) {
		cfg := DefaultConfig
		cfg.Port = 90000
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "port")
	})

	t.Run("invalid ndots", func(t *testing.T) {
		cfg := DefaultConfig
		cfg.NdotsValue = 16
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ndots")
	})

	t.Run("invalid annot mode", func(t *testing.T) {
		cfg := DefaultConfig
		cfg.AnnotationMode = "foo"
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotationMode")
	})
}
