// CLASSIFICATION: UNCLASSIFIED
// Binary server is the entry point for svc-nato-adapter.
//
// Startup sequence:
//  1. Load configuration from environment variables
//  2. Initialize structured logging
//  3. Create and register noop gRPC handler
//  4. Start gRPC server
//  5. Start health HTTP server
//  6. Block until SIGINT/SIGTERM
//  7. Graceful shutdown
//
// Feature: FEAT-15 NATO Interoperability
// UC: UC011 NATO Adapter
// Requirements: CR-NATO-001, CR-NATO-002, CR-NATO-003, CR-NATO-004, CR-NATO-005
package main

import (
"context"
"errors"
"fmt"
"log/slog"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/server"
"go.uber.org/zap"
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

zapLogger.Info("svc-nato-adapter starting",
zap.String("service", cfg.ServiceName),
zap.String("grpc_addr", cfg.GRPCAddr),
zap.String("health_port", cfg.HealthPort),
zap.String("environment", cfg.Environment),
)

// ── 4. gRPC server ───────────────────────────────────────────────────────
grpcServer, lis, err := server.New(cfg, zapLogger)
if err != nil {
return fmt.Errorf("server.New: %w", err)
}

// ── 5. Health HTTP server ─────────────────────────────────────────────────
go startHealthServer(ctx, cfg.HealthPort, zapLogger)

// ── 6. gRPC serve ────────────────────────────────────────────────────────
go func() {
zapLogger.Info("gRPC server listening", zap.String("addr", cfg.GRPCAddr))
if serveErr := grpcServer.Serve(lis); serveErr != nil {
zapLogger.Error("gRPC server error", zap.Error(serveErr))
}
}()

// ── 7. Wait for shutdown ──────────────────────────────────────────────────
<-ctx.Done()
zapLogger.Info("shutdown signal received, draining in-flight requests")

stopped := make(chan struct{})
go func() {
grpcServer.GracefulStop()
close(stopped)
}()

select {
case <-stopped:
zapLogger.Info("gRPC server stopped gracefully")
case <-time.After(30 * time.Second):
zapLogger.Warn("graceful stop timed out, forcing shutdown")
grpcServer.Stop()
}

zapLogger.Info("svc-nato-adapter shutdown complete")
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
