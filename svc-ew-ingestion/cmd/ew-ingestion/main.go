// CLASSIFICATION: UNCLASSIFIED
package main

import (
"context"
"fmt"
"net"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/health"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/interceptors"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/shutdown"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/telemetry"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/mapper"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/producer"
"go.uber.org/zap"
"google.golang.org/grpc"
)

func main() {
// 1. Load config
cfg := config.MustLoad()

// 2. Initialize telemetry (tracing + metrics + logger)
ctx := context.Background()
tp, err := telemetry.Init(ctx, telemetry.Config{
ServiceName:    "svc-ew-ingestion",
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
ClientID:   "svc-ew-ingestion",
}

// 4. Create Redpanda producer (for sensors.ew.intercepts)
prod, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
Connection:    connOpts,
ServiceName:   "svc-ew-ingestion",
SchemaVersion: "1.0.0",
})
if err != nil {
logger.Fatal("failed to create Redpanda producer", zap.Error(err))
}

// 5. Create DLQ producer
dlqProd, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
Connection:    connOpts,
ServiceName:   "svc-ew-ingestion",
SchemaVersion: "1.0.0",
})
if err != nil {
logger.Fatal("failed to create DLQ producer", zap.Error(err))
}

// 6. Create audit emitter
auditEmitter := audit.NewEmitter(prod, "svc-ew-ingestion", logger)

// 7. Create health checker
healthChecker := health.NewChecker()
healthChecker.Register("redpanda")
healthChecker.Register("grpc")

// 8. Create domain components
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
enricher := mapper.NewEnricher("svc-ew-ingestion", guard)

// Create observation producers
obsProd := producer.NewObservationProducer(prod, cfg.OutputTopic)
dlqObsProd := producer.NewObservationProducer(dlqProd, cfg.DLQTopic)

// 9. Create gRPC handler
	ingestionHandler := handler.NewIngestionHandler(
		validator, normalizer, enricher,
		obsProd, dlqObsProd, auditEmitter, logger, cfg.Coverage)

// 10. Create gRPC server with interceptor chain
meter := tp.MeterProvider.Meter("svc-ew-ingestion")
chainCfg := interceptors.ChainConfig{
Logger:              logger,
Meter:               meter,
ClassificationGuard: guard,
ServiceName:         "svc-ew-ingestion",
}
srv := grpc.NewServer(
grpc.ChainUnaryInterceptor(interceptors.BuildUnaryServerInterceptors(chainCfg)...),
grpc.ChainStreamInterceptor(interceptors.BuildStreamServerInterceptors(chainCfg)...),
)

// 11. Register services
ingestionv1.RegisterIngestionServiceServer(srv, ingestionHandler)
commonv1.RegisterHealthServiceServer(srv, health.NewServer(healthChecker))

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
zap.String("service", "svc-ew-ingestion"),
zap.String("version", cfg.ServiceVersion))

// 15. Wait for shutdown signal
if err := sm.Wait(); err != nil {
logger.Error("shutdown completed with errors", zap.Error(err))
}
}
