// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/handler"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger := newLogger("info")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err.Error())
		os.Exit(1)
	}

	zapLogger, err := newZapLogger(cfg.LogLevel)
	if err != nil {
		logger.Error("failed to create zap logger", "error", err.Error())
		os.Exit(1)
	}
	defer func() { _ = zapLogger.Sync() }()

	zapLogger.Info("svc-query starting", zap.String("service", cfg.ServiceName))

	// ─── ClickHouse client ───────────────────────────────────────────────────
	chClient, err := repository.NewClickHouseClient(cfg.ClickHouseDSN)
	if err != nil {
		zapLogger.Fatal("failed to create ClickHouse client", zap.Error(err))
	}
	defer func() { _ = chClient.Close() }()

	// ─── Repositories ────────────────────────────────────────────────────────
	tracksRepo := repository.NewTracksRepository(chClient)
	anomalyRepo := repository.NewAnomalyRepository(chClient)
	auditRepo := repository.NewAuditRepository(chClient)

	// ─── Domain guardrail ────────────────────────────────────────────────────
	guardrail := domain.NewQueryGuardrail(
		cfg.MaxQueryRangeDays,
		cfg.MaxResultRows,
		cfg.QueryTimeoutSec,
	)

	// ─── gRPC handler ────────────────────────────────────────────────────────
	querySrv := handler.NewQueryServer(
		tracksRepo, anomalyRepo, auditRepo,
		guardrail, cfg.DefaultPageSize, zapLogger,
	)

	// ─── gRPC server ─────────────────────────────────────────────────────────
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		zapLogger.Fatal("failed to listen on gRPC port",
			zap.Int("port", cfg.GRPCPort), zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	queryv1.RegisterQueryServiceServer(grpcServer, querySrv)

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
		zapLogger.Info("health server listening", zap.Int("port", cfg.HealthPort))
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Error("health server error", zap.Error(err))
		}
	}()

	go func() {
		zapLogger.Info("gRPC server listening", zap.Int("port", cfg.GRPCPort))
		if err := grpcServer.Serve(grpcLis); err != nil {
			zapLogger.Error("gRPC server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	zapLogger.Info("shutdown signal received")

	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Warn("health server shutdown error", zap.Error(err))
	}

	zapLogger.Info("svc-query stopped")
}

func newLogger(level string) *slog.Logger {
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
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func newZapLogger(level string) (*zap.Logger, error) {
	var cfg zap.Config
	switch level {
	case "debug":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()
	}
	return cfg.Build()
}
