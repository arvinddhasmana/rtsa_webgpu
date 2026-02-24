// CLASSIFICATION: UNCLASSIFIED

// Package main is the entry point for svc-query.
// It loads configuration, wires dependencies, and starts the gRPC server.
package main

import (
"context"
"log/slog"
"os"
"os/signal"
"strings"
"syscall"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/repository"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/server"
)

func main() {
// Structured JSON logging — never log classified data or raw payloads
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

slog.Info("svc-query starting", "version", "1.0.0")

// Load configuration from environment variables
cfg, err := config.Load()
if err != nil {
slog.Error("config load failed", "error", err)
os.Exit(1)
}

// Build query guardrail
guard, err := domain.NewQueryGuardrail(
cfg.MaxQueryRangeDays,
cfg.MaxResultRows,
cfg.QueryTimeoutSec,
cfg.DefaultPageSize,
cfg.MaxPageSize,
)
if err != nil {
slog.Error("guardrail init failed", "error", err)
os.Exit(1)
}

// Build ClickHouse client
chClient, err := repository.NewClickHouseClient(cfg)
if err != nil {
slog.Error("ClickHouse client init failed", "error", err)
os.Exit(1)
}
defer func() {
if err := chClient.Close(); err != nil {
slog.Error("ClickHouse close error", "error", err)
}
}()

// Build classification filter
classFilter := security.NewClassificationFilter()

// Build repositories
tracksRepo := repository.NewTracksRepository(chClient, classFilter, guard)
anomalyRepo := repository.NewAnomalyRepository(chClient, classFilter, guard)
auditRepo := repository.NewAuditRepository(chClient, classFilter, guard)

// Build audit emitter
brokers := strings.Split(cfg.RedpandaBrokers, ",")
auditEmitter, err := audit.NewRedpandaEmitter(brokers, cfg.ServiceName)
if err != nil {
slog.Error("audit emitter init failed", "error", err)
os.Exit(1)
}
defer func() {
if err := auditEmitter.Close(); err != nil {
slog.Error("audit emitter close error", "error", err)
}
}()

// Build handler
h := handler.New(tracksRepo, anomalyRepo, auditRepo, auditEmitter, guard, cfg.ServiceName)

// Build gRPC server with mTLS
grpcServer, err := server.NewGRPCServer(h, cfg.GRPCPort, cfg.TLSCertFile, cfg.TLSKeyFile, cfg.TLSCAFile)
if err != nil {
slog.Error("gRPC server init failed", "error", err)
os.Exit(1)
}

// Graceful shutdown on SIGTERM / SIGINT
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

errCh := make(chan error, 1)
go func() {
errCh <- grpcServer.Start()
}()

select {
case <-ctx.Done():
slog.Info("shutdown signal received")
grpcServer.GracefulStop()
case err := <-errCh:
if err != nil {
slog.Error("gRPC server error", "error", err)
os.Exit(1)
}
}

slog.Info("svc-query stopped")
}
