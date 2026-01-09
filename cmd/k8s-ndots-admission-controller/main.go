package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/admission"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/config"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/logging"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/metrics"
	"github.com/hawky-4s-/k8s-ndots-admission-controller/internal/server"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// 2. Setup logging
	logger := logging.NewLogger(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	slog.SetDefault(logger)

	// 3. Setup metrics
	reg := prometheus.NewRegistry()
	metricsRecorder := metrics.NewRecorder(reg)

	// 4. Initialize Kubernetes client
	// Try in-cluster config first
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to local kubeconfig
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
		}
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			logger.Warn("failed to create kubernetes client config, proceeding without namespace support", "error", err)
		}
	}

	var kubeClient kubernetes.Interface
	if kubeConfig != nil {
		kubeClient, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			logger.Error("failed to create kubernetes client", "error", err)
			os.Exit(1)
		}
	}

	// 5. Initialize components
	mutator := admission.NewMutator(cfg, logger, kubeClient)
	handler := admission.NewHandlerWithMetrics(mutator, logger, metricsRecorder)

	// 6. Create server config
	srvCfg := server.Config{
		Port:        cfg.Port,
		TLSCertPath: cfg.TLSCertPath,
		TLSKeyPath:  cfg.TLSKeyPath,
		Timeout:     cfg.Timeout,
	}

	// 7. Create router
	mux := http.NewServeMux()

	// Register application routes
	mux.HandleFunc("/mutate", handler.HandleMutate)

	// Server setup
	srv, err := server.New(srvCfg, mux)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Register health endpoints
	mux.HandleFunc("/healthz", srv.HandleHealthz)
	mux.HandleFunc("/readyz", srv.HandleReadyz)

	// 8. Start metrics server
	metricsSrv := metrics.NewServer(reg, cfg.MetricsPort)
	go func() {
		logger.Info("starting metrics server", "port", cfg.MetricsPort)
		if err := metricsSrv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", "error", err)
		}
	}()

	// 9. Start webhook server in background
	go func() {
		logger.Info("starting webhook server", "port", cfg.Port)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	logger.Info("shutting down servers...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown both servers
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("webhook server shutdown failed", "error", err)
	}
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("metrics server shutdown failed", "error", err)
	}

	logger.Info("servers stopped")
}
