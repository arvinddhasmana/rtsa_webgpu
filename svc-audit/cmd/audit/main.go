// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/consumer"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/handler"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "svc-audit: failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, err := buildLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "svc-audit: failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("svc-audit starting", zap.String("service", cfg.ServiceName))

	// ─── ClickHouse repository ────────────────────────────────────────────────
	repo, err := repository.NewAuditRepository(cfg.ClickHouseDSN)
	if err != nil {
		logger.Fatal("failed to create audit repository", zap.Error(err))
	}
	defer func() { _ = repo.Close() }()

	// ─── Audit consumer ──────────────────────────────────────────────────────
	kClient, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.RedpandaBrokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics("audit.events"),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		logger.Fatal("failed to create Kafka client", zap.Error(err))
	}

	auditConsumer := consumer.NewAuditConsumer(
		kClient,
		repo,
		cfg.BatchSize,
		cfg.FlushPeriodSec,
		logger,
	)
	defer auditConsumer.Close()

	// ─── Domain guardrail ─────────────────────────────────────────────────────
	guardrail := domain.NewQueryGuardrail(
		cfg.MaxQueryRangeDays,
		cfg.MaxResultRows,
		cfg.QueryTimeoutSec,
	)

	// ─── gRPC handler ─────────────────────────────────────────────────────────
	auditSrv := handler.NewAuditServer(repo, guardrail, cfg.DefaultPageSize, logger)

	// ─── gRPC server ──────────────────────────────────────────────────────────
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen on gRPC port",
			zap.Int("port", cfg.GRPCPort), zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, grpchealth_server)

	auditv1.RegisterAuditServiceServer(grpcServer, auditSrv)

	// ─── Health check HTTP server ─────────────────────────────────────────────
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	healthSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HealthPort),
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// ─── Graceful shutdown ────────────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("health server listening", zap.Int("port", cfg.HealthPort))
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health server error", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("gRPC server listening", zap.Int("port", cfg.GRPCPort))
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	consumerErrCh := make(chan error, 1)
	go func() {
		consumerErrCh <- auditConsumer.Start(ctx)
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("health server shutdown error", zap.Error(err))
	}

	select {
	case err := <-consumerErrCh:
		if err != nil {
			logger.Warn("consumer exited with error", zap.Error(err))
		}
	case <-shutdownCtx.Done():
		logger.Warn("timed out waiting for consumer to stop")
	}

	logger.Info("svc-audit stopped")
}

func buildLogger(level string) (*zap.Logger, error) {
	var cfg zap.Config
	switch level {
	case "debug":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()
	}
	return cfg.Build()
}
