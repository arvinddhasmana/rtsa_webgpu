// CLASSIFICATION: UNCLASSIFIED
// Binary server is the entry point for svc-training.
//
// Startup sequence:
//  1. Load configuration from environment variables
//  2. Initialize structured logging
//  3. Create Redpanda consumer for feedback.operator.validated
//  4. Create Redpanda producer for models.anomaly.candidates
//  5. Start health HTTP server
//  6. Start consume→produce noop loop
//  7. Wait for SIGINT/SIGTERM graceful shutdown
//
// Feature: FEAT-12 Training Pipeline
// UC: UC014, UC015
// Requirements: CR-FB-003, CR-FB-004
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

	"github.com/arvinddhasmana/rtsa_webgpu/svc-training/internal/config"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-training/internal/consumer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	// ── 1. Signal context for graceful shutdown ──────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── 2. Configuration ─────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		return fmt.Errorf("config.Load: %w", err)
	}

	// ── 3. Structured logging ────────────────────────────────────────────────
	zapLogger, err := buildLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return fmt.Errorf("buildLogger: %w", err)
	}
	defer func() { _ = zapLogger.Sync() }()

	zapLogger.Info("svc-training starting",
		zap.String("service", cfg.ServiceName),
		zap.String("input_topic", cfg.InputTopic),
		zap.String("output_topic", cfg.OutputTopic),
		zap.Strings("brokers", cfg.Brokers),
	)

	// --- Dummy gRPC Health Server ---
	grpcLis, _ := net.Listen("tcp", ":50051")
	grpcServerDummy := grpc.NewServer()
	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServerDummy, grpchealth_server)
	go grpcServerDummy.Serve(grpcLis)
	defer grpcServerDummy.GracefulStop()

	// ── 4. Redpanda consumer ─────────────────────────────────────────────────
	consumerClient, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(cfg.InputTopic),
	)
	if err != nil {
		return fmt.Errorf("kgo.NewClient (consumer): %w", err)
	}
	defer consumerClient.Close()

	// ── 5. Redpanda producer ─────────────────────────────────────────────────
	producerClient, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
	)
	if err != nil {
		return fmt.Errorf("kgo.NewClient (producer): %w", err)
	}
	defer producerClient.Close()

	// ── 6. Health HTTP server ─────────────────────────────────────────────────
	go startHealthServer(ctx, cfg.HealthPort, zapLogger)

	// ── 7. Training consumer loop ─────────────────────────────────────────────
	tc := consumer.New(consumerClient, producerClient, cfg.OutputTopic, zapLogger)

	zapLogger.Info("svc-training ready — consuming from topic",
		zap.String("topic", cfg.InputTopic),
		zap.String("consumer_group", cfg.ConsumerGroup),
	)

	if err := tc.Run(ctx); err != nil {
		return fmt.Errorf("TrainingConsumer.Run: %w", err)
	}

	zapLogger.Info("svc-training shutdown complete")
	return nil
}

// buildLogger creates a zap.Logger based on config.
func buildLogger(level, format string) (*zap.Logger, error) {
	var zapCfg zap.Config
	if format == "text" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	switch level {
	case "debug":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapCfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("buildLogger: %w", err)
	}
	return logger, nil
}

// startHealthServer runs a simple HTTP health check server.
func startHealthServer(ctx context.Context, addr string, logger *zap.Logger) {
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
			logger.Error("health server shutdown error", zap.Error(err))
		}
	}()
	logger.Info("health server listening", zap.String("addr", addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("health server error", zap.Error(err))
	}
}
