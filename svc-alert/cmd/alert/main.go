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

	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/consumer"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/handler"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger := newLogger("info")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err.Error())
		os.Exit(1)
	}

	logger = newLogger(cfg.LogLevel)
	logger.Info("svc-alert starting", "service", cfg.ServiceName)

	// ─── Prometheus metrics ───────────────────────────────────────────────────
	reg := prometheus.NewRegistry()
	m := metrics.New(reg)

	// ─── Domain — priority queue ───────────────────────────────────────────────
	queue := domain.NewAlertQueue(cfg.MaxQueueSize)

	// ─── Domain — acknowledger ─────────────────────────────────────────────────
	ackMetrics := &domain.AcknowledgerMetrics{
		TimeToAcknowledge: m.TimeToAcknowledge,
	}
	ack := domain.NewAcknowledger(queue, ackMetrics, logger)

	// ─── gRPC handlers ────────────────────────────────────────────────────────
	streamMetrics := &handler.StreamMetrics{
		StreamClients:        m.StreamClients,
		AlertsUnacknowledged: m.AlertsUnacknowledged,
	}
	streamH := handler.NewStreamHandler(queue, streamMetrics, logger)
	ackH := handler.NewAcknowledgeHandler(ack, logger)
	detailsH := handler.NewDetailsHandler(queue, logger)
	assigner := domain.NewAssigner(queue, logger)
	assignH := handler.NewAssignHandler(assigner, logger)
	alertServer := handler.NewAlertServer(streamH, ackH, detailsH, assignH)

	// ─── Redpanda consumer ────────────────────────────────────────────────────
	consumerMetrics := &consumer.ConsumerMetrics{
		AlertsReceived: m.AlertsReceived,
		QueueSize:      m.QueueSize,
	}

	franzClient, err := consumer.NewFranzConsumerClient(
		cfg.RedpandaBrokers,
		cfg.ConsumerGroup,
		logger,
	)
	if err != nil {
		logger.Error("failed to create Redpanda consumer", "error", err.Error())
		os.Exit(1)
	}

	alertConsumer := consumer.NewAlertConsumer(
		franzClient,
		queue,
		cfg.Topics,
		consumerMetrics,
		logger,
	)

	// ─── gRPC server ──────────────────────────────────────────────────────────
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Error("failed to listen on gRPC port", "port", cfg.GRPCPort, "error", err.Error())
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, grpchealth_server)

	inferencev1.RegisterAlertServiceServer(grpcServer, alertServer)

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

	// ─── Metrics HTTP server ──────────────────────────────────────────────────
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	metricsSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler:      metricsMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// ─── Graceful shutdown ────────────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start consumer in background.
	consumerErrCh := make(chan error, 1)
	go func() {
		consumerErrCh <- alertConsumer.Start(ctx)
	}()

	// Start health server.
	go func() {
		logger.Info("health server listening", "port", cfg.HealthPort)
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health server error", "error", err.Error())
		}
	}()

	// Start metrics server.
	go func() {
		logger.Info("metrics server listening", "port", cfg.MetricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", "error", err.Error())
		}
	}()

	// Start gRPC server.
	go func() {
		logger.Info("gRPC server listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("gRPC server error", "error", err.Error())
		}
	}()

	// Wait for shutdown signal.
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Graceful gRPC shutdown.
	grpcServer.GracefulStop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("health server shutdown error", "error", err.Error())
	}
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("metrics server shutdown error", "error", err.Error())
	}
	if err := alertConsumer.Close(); err != nil {
		logger.Warn("consumer close error", "error", err.Error())
	}

	// Wait for consumer goroutine.
	select {
	case err := <-consumerErrCh:
		if err != nil {
			logger.Warn("consumer exited with error", "error", err.Error())
		}
	case <-shutdownCtx.Done():
		logger.Warn("timed out waiting for consumer to stop")
	}

	logger.Info("svc-alert stopped")
}

// newLogger creates a structured slog.Logger at the given log level.
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
