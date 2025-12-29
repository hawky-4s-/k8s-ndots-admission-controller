package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Port        int
	TLSCertPath string
	TLSKeyPath  string
	Timeout     time.Duration
}

type Server struct {
	httpServer *http.Server
	config     Config
}

// New creates a new webhook server.
func New(cfg Config, handler http.Handler) (*Server, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		TLSConfig:         tlsConfig,
		ReadTimeout:       cfg.Timeout,
		WriteTimeout:      cfg.Timeout,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return &Server{
		config:     cfg,
		httpServer: httpServer,
	}, nil
}

// Start starts the server.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServeTLS(s.config.TLSCertPath, s.config.TLSKeyPath)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// HandleHealthz handles liveness probes.
func (s *Server) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// HandleReadyz handles readiness probes.
func (s *Server) HandleReadyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
