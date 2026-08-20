// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"fmt"
	"net"
	"time"

	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/audit"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/health"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/interceptors"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/shutdown"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/telemetry"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/config"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/handler"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/mapper"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/producer"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// 1. Load configuration
	cfg := config.MustLoad()

	// 2. Initialize telemetry (tracing + metrics + logger)
	ctx := context.Background()
	tp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    "svc-ais-ingestion",
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		OTelEndpoint:   cfg.OTelEndpoint,
	})
	if err != nil {
		tmpLogger, _ := zap.NewProduction()
		tmpLogger.Fatal("failed to initialize telemetry", zap.Error(err))
	}
	logger := tp.Logger

	// 3. Create classification guard
	guard := classification.NewGuard(classification.StringToLevel(cfg.MaxClassification))

	// Build connection options
	connOpts := redpanda.ConnectionOptions{
		Brokers:    cfg.RedpandaBrokers,
		TLSEnabled: cfg.RedpandaTLSEnabled,
		ClientID:   "svc-ais-ingestion",
	}

	// 4. Create Redpanda producer (for sensors.ew.intercepts)
	prod, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
		Connection:    connOpts,
		ServiceName:   "svc-ais-ingestion",
		SchemaVersion: "1.0.0",
	})
	if err != nil {
		logger.Fatal("failed to create Redpanda producer", zap.Error(err))
	}

	// 5. Create DLQ producer
	dlqProd, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
		Connection:    connOpts,
		ServiceName:   "svc-ais-ingestion",
		SchemaVersion: "1.0.0",
	})
	if err != nil {
		logger.Fatal("failed to create DLQ producer", zap.Error(err))
	}

	// 6. Create audit emitter
	auditEmitter := audit.NewEmitter(prod, "svc-ais-ingestion", logger)

	// 7. Create health checker
	healthChecker := health.NewChecker()
	healthChecker.Register("redpanda")
	healthChecker.Register("grpc")

	// 8. Create domain components
	validator := domain.NewValidator()
	normalizer := domain.NewNormalizer()
	enricher := mapper.NewEnricher("svc-ais-ingestion", guard)

	// Create observation producers
	obsProd := producer.NewObservationProducer(prod, cfg.OutputTopic)
	dlqObsProd := producer.NewObservationProducer(dlqProd, cfg.DLQTopic)

	// 9. Create gRPC handler
	ingestionHandler := handler.NewIngestionHandler(
		validator, normalizer, enricher,
		obsProd, dlqObsProd, auditEmitter, logger, cfg.Coverage)

	// 10. Create gRPC server with interceptor chain
	meter := tp.MeterProvider.Meter("svc-ais-ingestion")
	chainCfg := interceptors.ChainConfig{
		Logger:              logger,
		Meter:               meter,
		ClassificationGuard: guard,
		ServiceName:         "svc-ais-ingestion",
	}
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.BuildUnaryServerInterceptors(chainCfg)...),
		grpc.ChainStreamInterceptor(interceptors.BuildStreamServerInterceptors(chainCfg)...),
	)

	// 11. Register services
	ingestionv1.RegisterIngestionServiceServer(srv, ingestionHandler)
	health.RegisterAll(srv, healthChecker)

	// 12. Create shutdown manager
	sm := shutdown.NewManager(logger, 30*time.Second)
	sm.Register("grpc-server", func(ctx context.Context) error {
		srv.GracefulStop()
		return nil
	})
	sm.Register("producer", func(ctx context.Context) error {
		return prod.Close()
	})
	sm.Register("dlq-producer", func(ctx context.Context) error {
		return dlqProd.Close()
	})
	sm.Register("telemetry", tp.Shutdown)

	// 13. Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.Int("port", cfg.GRPCPort), zap.Error(err))
	}

	go func() {
		logger.Info("gRPC server starting", zap.Int("port", cfg.GRPCPort))
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	// 14. Set health to SERVING
	healthChecker.SetStatus("grpc", health.StatusServing)
	healthChecker.SetStatus("redpanda", health.StatusServing)
	logger.Info("service ready",
		zap.String("service", "svc-ais-ingestion"),
		zap.String("version", cfg.ServiceVersion))

	// 15. Wait for shutdown signal
	if err := sm.Wait(); err != nil {
		logger.Error("shutdown completed with errors", zap.Error(err))
	}
}
