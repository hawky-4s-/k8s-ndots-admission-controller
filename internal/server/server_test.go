package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cfg := Config{
		Port:        8443,
		TLSCertPath: "test.crt", // These won't be read in this test
		TLSKeyPath:  "test.key",
	}

	srv, err := New(cfg, http.NewServeMux())
	require.NoError(t, err)
	assert.NotNil(t, srv)
	assert.NotNil(t, srv.httpServer)
	assert.Equal(t, ":8443", srv.httpServer.Addr)
	assert.NotNil(t, srv.httpServer.TLSConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), srv.httpServer.TLSConfig.MinVersion)
}

func TestServer_HealthEndpoints(t *testing.T) {
	// Create a dummy server for testing handlers directly
	s := &Server{}

	tests := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{
			name:     "healthz returns 200 ok",
			path:     "/healthz",
			wantCode: http.StatusOK,
			wantBody: "ok",
		},
		{
			name:     "readyz returns 200 ok",
			path:     "/readyz",
			wantCode: http.StatusOK,
			wantBody: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			// We need to implement these handlers
			if tt.path == "/healthz" {
				s.HandleHealthz(w, req)
			} else {
				s.HandleReadyz(w, req)
			}

			resp := w.Result()
			assert.Equal(t, tt.wantCode, resp.StatusCode)

			// Check body
			body := w.Body.String()
			assert.Equal(t, tt.wantBody, body)
		})
	}
}
