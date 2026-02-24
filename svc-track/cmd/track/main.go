// CLASSIFICATION: UNCLASSIFIED
// Binary track is the entry point for svc-track.
//
// Startup sequence:
//  1. Load configuration from environment variables
//  2. Initialize structured logging
//  3. Create Prometheus metrics registry
//  4. Create in-memory TrackCache
//  5. Create gRPC handler and wire onChange callback
//  6. Start Redpanda consumer goroutine
//  7. Start gRPC server
//  8. Start health + metrics HTTP servers
//  9. Block until SIGINT/SIGTERM
//
// 10. Graceful shutdown
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-UI-002, CR-SEC-001, NFR-AVAIL-001
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/consumer"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/handler"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	// ── 1. Signal context for graceful shutdown ──────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── 2. Configuration ─────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		// log.Fatal is acceptable in main init path only.
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ── 3. Structured logging ────────────────────────────────────────────────
	logger := buildLogger(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(logger)

	logger.Info("svc-track starting",
		slog.String("service", cfg.ServiceName),
		slog.String("grpc_addr", cfg.GRPCAddr),
		slog.String("health_port", cfg.HealthPort),
	)

	// ── 4. Metrics ───────────────────────────────────────────────────────────
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	// ── 5. Track cache ───────────────────────────────────────────────────────
	cache := domain.NewTrackCache(cfg.HistoryMaxPoints)

	// ── 6. gRPC handlers ─────────────────────────────────────────────────────
	filter := &domain.FilterEngine{}
	streamH := handler.NewStreamHandler(cache, filter, m, logger, cfg.StreamChannelBufferSize)
	detailsH := handler.NewDetailsHandler(cache)
	historyH := handler.NewHistoryHandler(cache)

	trackSvc := &trackServiceServer{
		stream:  streamH,
		details: detailsH,
		history: historyH,
	}

	// ── 7. gRPC server ───────────────────────────────────────────────────────
	// TODO: Replace insecure credentials with mTLS when certificates are available.
	// RULE-SEC-02: production deployments MUST use mTLS.
	var grpcCreds grpc.ServerOption
	if cfg.TLSEnabled {
		tlsCreds, tlsErr := loadTLSCredentials(cfg)
		if tlsErr != nil {
			logger.Error("failed to load TLS credentials", slog.String("error", tlsErr.Error()))
			os.Exit(1)
		}
		grpcCreds = tlsCreds
	} else {
		logger.Warn("TLS disabled — mTLS required for production (RTSA_TLS_ENABLED=true)")
		grpcCreds = grpc.Creds(insecure.NewCredentials())
	}

	grpcServer := grpc.NewServer(grpcCreds)
	entityv1.RegisterTrackServiceServer(grpcServer, trackSvc)
	reflection.Register(grpcServer) // dev convenience

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		logger.Error("failed to listen on gRPC address",
			slog.String("addr", cfg.GRPCAddr),
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// ── 8. Redpanda consumer ─────────────────────────────────────────────────
	fusedConsumer, err := consumer.NewFusedTrackConsumer(
		cfg.RedpandaBrokers, cfg.ConsumerGroupID, cache, logger)
	if err != nil {
		logger.Error("failed to create fused track consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer fusedConsumer.Close()

	go func() {
		if runErr := fusedConsumer.Run(ctx); runErr != nil {
			logger.Error("fused track consumer error", slog.String("error", runErr.Error()))
		}
	}()

	// ── 9. Health + metrics HTTP servers ─────────────────────────────────────
	go startHealthServer(ctx, cfg.HealthPort, logger)
	go startMetricsServer(ctx, cfg.MetricsPort, reg, logger)

	// ── 10. gRPC serve ───────────────────────────────────────────────────────
	go func() {
		logger.Info("gRPC server listening", slog.String("addr", cfg.GRPCAddr))
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			logger.Error("gRPC server error", slog.String("error", serveErr.Error()))
		}
	}()

	// ── 11. Wait for shutdown ─────────────────────────────────────────────────
	<-ctx.Done()
	logger.Info("shutdown signal received, draining in-flight requests")

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Info("gRPC server stopped gracefully")
	case <-time.After(30 * time.Second):
		logger.Warn("graceful stop timed out, forcing shutdown")
		grpcServer.Stop()
	}

	logger.Info("svc-track shutdown complete")
}

// trackServiceServer composes the individual handlers into a single gRPC server.
type trackServiceServer struct {
	entityv1.UnimplementedTrackServiceServer
	stream  *handler.StreamHandler
	details *handler.DetailsHandler
	history *handler.HistoryHandler
}

func (s *trackServiceServer) StreamTracks(req *entityv1.StreamTracksRequest, stream grpc.ServerStreamingServer[entityv1.TrackUpdate]) error {
	return s.stream.StreamTracks(req, stream)
}

func (s *trackServiceServer) GetTrackDetails(ctx context.Context, req *entityv1.GetTrackDetailsRequest) (*entityv1.FusedTrack, error) {
	return s.details.GetTrackDetails(ctx, req)
}

func (s *trackServiceServer) GetTrackHistory(ctx context.Context, req *entityv1.GetTrackHistoryRequest) (*entityv1.TrackHistoryResponse, error) {
	return s.history.GetTrackHistory(ctx, req)
}

// buildLogger creates a slog.Logger based on config.
func buildLogger(level, format string) *slog.Logger {
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
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

// startHealthServer runs a simple HTTP health check server.
func startHealthServer(ctx context.Context, addr string, logger *slog.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"READY"}`))
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			logger.Error("health server shutdown error", slog.String("error", err.Error()))
		}
	}()
	logger.Info("health server listening", slog.String("addr", addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("health server error", slog.String("error", err.Error()))
	}
}

// startMetricsServer runs a Prometheus /metrics HTTP server.
func startMetricsServer(ctx context.Context, addr string, reg *prometheus.Registry, logger *slog.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			logger.Error("metrics server shutdown error", slog.String("error", err.Error()))
		}
	}()
	logger.Info("metrics server listening", slog.String("addr", addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("metrics server error", slog.String("error", err.Error()))
	}
}

// loadTLSCredentials loads mTLS credentials from the configured certificate paths.
func loadTLSCredentials(cfg *config.Config) (grpc.ServerOption, error) {
	return nil, fmt.Errorf("loadTLSCredentials: mTLS not yet implemented — set RTSA_TLS_ENABLED=false for development")
}
