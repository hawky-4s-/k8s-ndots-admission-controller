package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	NdotsValue       int
	AnnotationKey    string
	AnnotationMode   string
	NamespaceInclude []string
	NamespaceExclude []string
	Port             int
	TLSCertPath      string
	TLSKeyPath       string
	Timeout          time.Duration
	LogLevel         string
	LogFormat        string
	MetricsPort      int
}

var DefaultConfig = Config{
	Port:             8443,
	NdotsValue:       2,
	AnnotationKey:    "change-ndots",
	AnnotationMode:   "opt-out",
	NamespaceExclude: []string{"kube-system", "kube-public", "kube-node-lease"},
	Timeout:          10 * time.Second,
	TLSCertPath:      "/certs/tls.crt",
	TLSKeyPath:       "/certs/tls.key",
	LogLevel:         "info",
	LogFormat:        "json",
	MetricsPort:      8080,
}

func Load() (*Config, error) {
	cfg := DefaultConfig

	if v := os.Getenv("PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}
	if v := os.Getenv("NDOTS_VALUE"); v != "" {
		if ndots, err := strconv.Atoi(v); err == nil {
			cfg.NdotsValue = ndots
		}
	}
	if v := os.Getenv("ANNOTATION_KEY"); v != "" {
		cfg.AnnotationKey = v
	}
	if v := os.Getenv("ANNOTATION_MODE"); v != "" {
		cfg.AnnotationMode = v
	}
	if v := os.Getenv("NAMESPACE_INCLUDE"); v != "" {
		cfg.NamespaceInclude = splitAndTrim(v)
	}
	if v := os.Getenv("NAMESPACE_EXCLUDE"); v != "" {
		cfg.NamespaceExclude = splitAndTrim(v)
	}
	if v := os.Getenv("TLS_CERT_PATH"); v != "" {
		cfg.TLSCertPath = v
	}
	if v := os.Getenv("TLS_KEY_PATH"); v != "" {
		cfg.TLSKeyPath = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.LogFormat = v
	}
	if v := os.Getenv("METRICS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.MetricsPort = port
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if c.NdotsValue < 0 || c.NdotsValue > 15 {
		return errors.New("ndotsValue must be between 0 and 15")
	}

	validModes := map[string]bool{"always": true, "opt-in": true, "opt-out": true}
	if !validModes[c.AnnotationMode] {
		return errors.New("annotationMode must be 'always', 'opt-in', or 'opt-out'")
	}

	if c.TLSCertPath == "" {
		return errors.New("tlsCertPath is required")
	}
	if c.TLSKeyPath == "" {
		return errors.New("tlsKeyPath is required")
	}

	return nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
