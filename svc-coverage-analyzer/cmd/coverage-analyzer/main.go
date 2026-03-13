// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"time"

	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/shutdown"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/telemetry"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-coverage-analyzer/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-coverage-analyzer/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-coverage-analyzer/internal/producer"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()
	tp, err := telemetry.Init(ctx, telemetry.Config{
		ServiceName:    "svc-coverage-analyzer",
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		OTelEndpoint:   cfg.OTelEndpoint,
	})
	if err != nil {
		tmpLogger, _ := zap.NewProduction()
		tmpLogger.Fatal("failed to initialize telemetry", zap.Error(err))
	}
	logger := tp.Logger

	connOpts := redpanda.ConnectionOptions{
		Brokers:    cfg.RedpandaBrokers,
		TLSEnabled: cfg.RedpandaTLSEnabled,
		ClientID:   "svc-coverage-analyzer",
	}

	// 1. Create Producer
	rawProd, err := redpanda.NewProducer(ctx, redpanda.ProducerConfig{
		Connection:    connOpts,
		ServiceName:   "svc-coverage-analyzer",
		SchemaVersion: "1.0.0",
	})
	if err != nil {
		logger.Fatal("failed to create producer", zap.Error(err))
	}
	alertProd := producer.NewAlertProducer(rawProd, cfg.AlertTopic, logger)

	detector := domain.NewGapDetector(logger)

	// 2. Define Handler
	handler := func(ctx context.Context, record *kgo.Record) error {
		obs := &ingestionv1.SensorObservation{}
		if err := protojson.Unmarshal(record.Value, obs); err != nil {
			logger.Error("failed to unmarshal observation", zap.Error(err))
			return nil // Skip
		}

		alert, err := detector.Analyze(ctx, obs)
		if err != nil {
			logger.Error("analysis error", zap.Error(err))
			return nil
		}

		if alert != nil {
			if err := alertProd.Produce(ctx, alert); err != nil {
				logger.Error("failed to produce alert", zap.Error(err))
			}
		}
		return nil
	}

	// 3. Create Consumer
	rawConsumer, err := redpanda.NewConsumer(ctx, redpanda.ConsumerConfig{
		Connection:    connOpts,
		ConsumerGroup: "coverage-analyzer-group",
		Topics:        []string{cfg.InputTopic},
		Handler:       handler,
		Logger:        logger,
	})
	if err != nil {
		logger.Fatal("failed to create consumer", zap.Error(err))
	}

	// 4. Shutdown Manager
	sm := shutdown.NewManager(logger, 30*time.Second)
	sm.Register("consumer", func(ctx context.Context) error {
		return rawConsumer.Close()
	})
	sm.Register("producer", func(ctx context.Context) error {
		return rawProd.Close()
	})
	sm.Register("telemetry", tp.Shutdown)

	// 5. Start
	logger.Info("Coverage Analyzer starting", zap.String("input", cfg.InputTopic), zap.String("output", cfg.AlertTopic))
	go func() {
		if err := rawConsumer.Start(ctx); err != nil {
			logger.Error("consumer stopped with error", zap.Error(err))
		}
	}()

	if err := sm.Wait(); err != nil {
		logger.Error("shutdown completed with errors", zap.Error(err))
	}
}
