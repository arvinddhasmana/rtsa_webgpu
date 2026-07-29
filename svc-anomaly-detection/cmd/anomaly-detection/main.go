// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/health"
	pkgshutdown "github.com/arvinddhasmana/rtsa_webgpu/pkg/shutdown"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/config"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/consumer"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain/detectors"
	pipeline "github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/handler"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/producer"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/state"
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
	// ── Load Configuration ────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config.Load: %w", err)
	}

	// ── Structured Logger ─────────────────────────────────────────────────────
	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: logLevel}
	if cfg.LogFormat == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger := slog.New(handler).With("service", cfg.ServiceName)

	// --- Dummy gRPC Health Server ---
	grpcLis, _ := net.Listen("tcp", ":50051")
	grpcServerDummy := grpc.NewServer()
	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServerDummy, grpchealth_server)
	go grpcServerDummy.Serve(grpcLis)
	defer grpcServerDummy.GracefulStop()


	slog.SetDefault(logger)

	logger.Info("starting anomaly detection service",
		"service", cfg.ServiceName,
		"model_version", cfg.ModelVersion,
	)

	// ── Context & Graceful Shutdown ───────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Track History ─────────────────────────────────────────────────────────
	trackHistory := state.NewTrackHistory(cfg.TrackHistoryMaxEntries, cfg.TrackHistoryMaxAge)

	// Periodic cleanup of stale history entries.
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				trackHistory.Cleanup()
				logger.Debug("track history cleanup completed", "track_count", trackHistory.Count())
			case <-ctx.Done():
				return
			}
		}
	}()

	// ── Exclusion Zones ───────────────────────────────────────────────────────
	exclusionZones := make([]domain.ExclusionZone, 0, len(cfg.ExclusionZones))
	for _, z := range cfg.ExclusionZones {
		exclusionZones = append(exclusionZones, domain.ExclusionZone{
			Name:      z.Name,
			CenterLat: z.CenterLat,
			CenterLon: z.CenterLon,
			RadiusNM:  z.RadiusNM,
		})
	}

	// ── Feature Extractor ─────────────────────────────────────────────────────
	featureExtractor := domain.NewFeatureExtractor(trackHistory, exclusionZones)

	// ── Redpanda Producer ─────────────────────────────────────────────────────
	alertProducerClient, err := producer.NewRedpandaMessageProducer(ctx, cfg.RedpandaBrokers, cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("producer.NewRedpandaMessageProducer: %w", err)
	}
	defer func() {
		if closeErr := alertProducerClient.Close(); closeErr != nil {
			logger.Error("alert producer close error", "error", closeErr)
		}
	}()

	alertProd := producer.NewAlertProducer(alertProducerClient, logger)

	// ── Detectors ─────────────────────────────────────────────────────────────
	pipelineCfg := pipeline.DetectionPipelineConfig{
		ModelVersion:       cfg.ModelVersion,
		SpeedDetector:      detectors.NewSpeedDetector(cfg.Detectors.SpeedSigmaThresh),
		RouteDetector:      detectors.NewRouteDeviationDetector(cfg.Detectors.RouteDeviationDeg, cfg.Detectors.RouteSustainedN),
		AISDetector:        detectors.NewAISManipulationDetector(cfg.Detectors.AISDiscrepancyNM),
		BehavioralDetector: detectors.NewBehavioralDetector(cfg.Detectors.BehavioralConfThr),
		TemporalDetector:   detectors.NewTemporalDetector(cfg.Detectors.TemporalPValue),
		ProximityDetector:  detectors.NewProximityDetector(),
		SpeedEnabled:       cfg.Detectors.SpeedEnabled,
		RouteEnabled:       cfg.Detectors.RouteEnabled,
		AISEnabled:         cfg.Detectors.AISEnabled,
		BehavioralEnabled:  cfg.Detectors.BehavioralEnabled,
		TemporalEnabled:    cfg.Detectors.TemporalEnabled,
		ProximityEnabled:   cfg.Detectors.ProximityEnabled,
	}

	pipeline := pipeline.NewDetectionPipeline(featureExtractor, pipelineCfg, alertProd, logger)

	// ── Redpanda Consumer ─────────────────────────────────────────────────────
	consumerClient := consumer.NewRedpandaMessageConsumer(cfg.RedpandaBrokers, cfg.ConsumerGroup, logger)
	defer func() {
		if closeErr := consumerClient.Close(); closeErr != nil {
			logger.Error("consumer close error", "error", closeErr)
		}
	}()

	trackConsumer := consumer.NewTrackConsumer(consumerClient, logger)

	// ── Health Server ─────────────────────────────────────────────────────────
	healthServer := health.NewHTTPServer(logger)
	go func() {
		if err := healthServer.ListenAndServe(ctx, cfg.HealthAddr); err != nil {
			logger.Error("health server error", "error", err)
		}
	}()

	// ── Signal Handler ────────────────────────────────────────────────────────
	go pkgshutdown.WaitForSignal(cancel, logger)

	logger.Info("anomaly detection service ready",
		"topics", cfg.InputTopics,
		"consumer_group", cfg.ConsumerGroup,
		"health_addr", cfg.HealthAddr,
	)

	// ── Main Loop ─────────────────────────────────────────────────────────────
	return trackConsumer.Start(ctx, cfg.InputTopics, pipeline.HandleTrack)
}
