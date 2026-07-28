package webtransport
// CLASSIFICATION: UNCLASSIFIED
// Binary webtransport is the entry point for svc-webtransport — the RTSA
// hot-path fan-out service.
//
// Startup sequence:
//  1. Load configuration from RTSA_* environment variables
//  2. Initialise OpenTelemetry (zap logger + meter + tracer)
//  3. Build the Redpanda-backed TrackSource (fused tracks -> 128-byte records)
//  4. Build the WebTransport server (QUIC datagrams to browser COP clients)
//  5. Start the source consumer, the WebTransport server, health and metrics
//  6. Block until SIGINT/SIGTERM, then shut down gracefully
//
// Feature: FEAT-13 Situational Awareness UI (hot path)
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-UI-002, CR-SEC-001, NFR-PERF-001, NFR-AVAIL-001
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/telemetry"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/webtransport"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-webtransport/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-webtransport/internal/source"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── 1. Configuration ──────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "svc-webtransport: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Telemetry (logger + meter + tracer) ────────────────────────────────
	tp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: version,
		Environment:    cfg.Environment,
		OTelEndpoint:   cfg.OTelEndpoint,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "svc-webtransport: telemetry init: %v\n", err)
		os.Exit(1)
	}
	logger := tp.Logger
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()
	meter := tp.MeterProvider.Meter(cfg.ServiceName)

	logger.Info("svc-webtransport starting",
		zap.String("version", version),
		zap.String("listen_addr", cfg.WTListenAddr),
		zap.Strings("topics", cfg.Topics),
		zap.String("consumer_group", cfg.ConsumerGroup),
		zap.String("health_port", cfg.HealthPort),
		zap.String("metrics_port", cfg.MetricsPort),
	)

	// ── 3. Track source (Redpanda -> 128-byte records) ────────────────────────
	src, err := source.New(ctx, source.Config{
		Brokers:              cfg.RedpandaBrokers,
		Topics:               cfg.Topics,
		ConsumerGroup:        cfg.ConsumerGroup,
		StartOffset:          cfg.StartOffset,
		SubscriberBufferSize: cfg.SubscriberBufferSize,
		ClientID:             cfg.ServiceName,
		TLSEnabled:           cfg.RedpandaTLSEnabled,
		TLSCAFile:            cfg.TLSCACert,
		TLSCertFile:          cfg.TLSCert,
		TLSKeyFile:           cfg.TLSKey,
	}, logger, meter)
	if err != nil {
		logger.Error("failed to create track source", zap.Error(err))
		os.Exit(1)
	}
	defer func() { _ = src.Close() }()

	// ── 4. WebTransport server ─────────────────────────────────────────────────
	// RULE-SEC-02: WebTransport mandates TLS; operator sessions are authenticated
	// with signed JWTs. The signing secret is injected from Key Vault, never code.
	if len(cfg.AllowedOrigins) == 0 {
		logger.Warn("RTSA_WT_ALLOWED_ORIGINS is empty — origin checking disabled (dev only)")
	}
	wtServer, err := webtransport.New(webtransport.Config{
		ListenAddr:        cfg.WTListenAddr,
		TLSCert:           cfg.TLSCert,
		TLSKey:            cfg.TLSKey,
		JWTSecret:         cfg.JWTSecret,
		AllowedOrigins:    cfg.AllowedOrigins,
		MaxSessions:       cfg.MaxSessions,
		DatagramBatchSize: cfg.DatagramBatchSize,
	}, src, meter, logger)
	if err != nil {
		logger.Error("failed to create webtransport server", zap.Error(err))
		os.Exit(1)
	}

	// ── 5. Start consumers and servers ─────────────────────────────────────────
	go func() {
		if runErr := src.Run(ctx); runErr != nil && ctx.Err() == nil {
			logger.Error("track source stopped unexpectedly", zap.Error(runErr))
			stop()
		}
	}()

	go func() {
		logger.Info("webtransport server listening", zap.String("addr", cfg.WTListenAddr))
		if serveErr := wtServer.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey); serveErr != nil &&
			!errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("webtransport server error", zap.Error(serveErr))
			stop()
		}
	}()

	healthSrv := startHealthServer(cfg.HealthPort, src, logger)
	metricsSrv := startMetricsServer(cfg.MetricsPort, logger)

	// ── 6. Wait for shutdown ───────────────────────────────────────────────────
	<-ctx.Done()
	logger.Info("shutdown signal received, draining")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := wtServer.Close(); err != nil {
		logger.Warn("webtransport close error", zap.Error(err))
	}
	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("health server shutdown error", zap.Error(err))
	}
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("metrics server shutdown error", zap.Error(err))
	}
	if err := src.Close(); err != nil {
		logger.Warn("source close error", zap.Error(err))
	}

	logger.Info("svc-webtransport stopped")
}

// startHealthServer starts the liveness (/healthz) and readiness (/readyz)
// endpoints. Readiness reflects broker connectivity.
func startHealthServer(addr string, src *source.Source, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if !src.Healthy(ctx) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	go func() {
		logger.Info("health server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("health server error", zap.Error(err))
		}
	}()
	return srv
}

// startMetricsServer serves Prometheus metrics from the default registry, which
// the OpenTelemetry Prometheus exporter feeds.
func startMetricsServer(addr string, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	go func() {
		logger.Info("metrics server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()
	return srv
}
