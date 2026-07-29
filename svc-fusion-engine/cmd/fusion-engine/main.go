// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/audit"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/config"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/consumer"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/handler"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/producer"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	cfg := config.Load()

	logger, err := buildLogger(cfg.LogLevel)

	// --- Dummy gRPC Health Server ---
	grpcLis, _ := net.Listen("tcp", ":50051")
	grpcServerDummy := grpc.NewServer()
	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServerDummy, grpchealth_server)
	go grpcServerDummy.Serve(grpcLis)
	defer grpcServerDummy.GracefulStop()


	if err != nil {
		os.Stderr.WriteString("failed to build logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck

	logger.Info("fusion engine starting",
		zap.String("service", cfg.ServiceName),
		zap.String("version", cfg.ServiceVersion),
		zap.String("environment", cfg.Environment),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Wire domain components
	kf := domain.NewKalmanFilter()
	manager := domain.NewTrackManager(kf)
	gating := domain.NewGatingFilter(
		domain.GatingConfig{
			MaxDistanceNM: cfg.GateSurfaceDistanceNM,
			MaxTimeDelta:  time.Duration(cfg.GateSurfaceTimeSec) * time.Second,
		},
		domain.GatingConfig{
			MaxDistanceNM: cfg.GateAirDistanceNM,
			MaxTimeDelta:  time.Duration(cfg.GateAirTimeSec) * time.Second,
		},
		domain.GatingConfig{
			MaxDistanceNM: cfg.GateSubDistanceNM,
			MaxTimeDelta:  time.Duration(cfg.GateSubTimeSec) * time.Second,
		},
	)
	scorer := domain.NewCorrelationScorer(
		cfg.WeightPosition, cfg.WeightVelocity, cfg.WeightType, cfg.WeightTemporal,
		cfg.AutoCorrelateThreshold, cfg.TentativeCorrelateThreshold,
	)

	// Metrics
	metrics := handler.NewFusionMetrics(prometheus.DefaultRegisterer)

	// Audit emitter
	auditEmitter := audit.NewLogEmitter(logger)

	// Track producer
	trackProd, err := producer.NewTrackProducer(ctx, cfg.RedpandaBrokers, cfg.OutputTopicPrefix, logger)
	if err != nil {
		logger.Fatal("failed to create track producer", zap.Error(err))
	}
	defer trackProd.Close()

	// Fusion pipeline
	defaultGate := domain.GatingConfig{
		MaxDistanceNM: cfg.GateSurfaceDistanceNM,
		MaxTimeDelta:  time.Duration(cfg.GateSurfaceTimeSec) * time.Second,
	}
	pipeline := handler.NewFusionPipeline(
		gating, scorer, manager, trackProd, auditEmitter, logger, metrics,
		cfg.AutoCorrelateThreshold, cfg.TentativeCorrelateThreshold, defaultGate,
	)

	// Stale monitor
	staleMonitor := domain.NewStaleMonitor(
		manager,
		time.Duration(cfg.StaleTimeoutSec)*time.Second,
		time.Duration(cfg.DropTimeoutSec)*time.Second,
		time.Duration(cfg.StaleCheckInterval)*time.Second,
		pipeline.OnTrackDropped,
	)
	go staleMonitor.Start(ctx)

	// Sensor consumer
	sensorConsumer, err := consumer.NewSensorConsumer(cfg.RedpandaBrokers, cfg.ConsumerGroup, cfg.InputTopics, logger)
	if err != nil {
		logger.Fatal("failed to create sensor consumer", zap.Error(err))
	}
	defer sensorConsumer.Close()

	// Graceful shutdown on SIGTERM / SIGINT
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
	}()

	logger.Info("fusion engine running", zap.Strings("topics", cfg.InputTopics))
	if err := sensorConsumer.Run(ctx, pipeline.HandleObservation); err != nil {
		logger.Info("consumer stopped", zap.Error(err))
	}
	logger.Info("fusion engine stopped")
}

func buildLogger(level string) (*zap.Logger, error) {
	if level == "debug" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
