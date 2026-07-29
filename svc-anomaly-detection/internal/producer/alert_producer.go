// CLASSIFICATION: UNCLASSIFIED
package producer

import (
	"context"
	"fmt"
	"log/slog"

	inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// MessageProducer abstracts alert publication for testability.
type MessageProducer interface {
	Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error
	Close() error
}

// AlertProducer produces AnomalyAlert messages to severity-specific topics.
type AlertProducer struct {
	producer MessageProducer
	logger   *slog.Logger
}

// NewAlertProducer creates a new AlertProducer.
func NewAlertProducer(producer MessageProducer, logger *slog.Logger) *AlertProducer {
	return &AlertProducer{
		producer: producer,
		logger:   logger,
	}
}

// Produce serialises and produces an AnomalyAlert to the specified topic.
func (ap *AlertProducer) Produce(ctx context.Context, topic string, alert *inferencev1.AnomalyAlert) error {
	if topic == "" {
		return fmt.Errorf("[producer.AlertProducer.Produce]: topic is empty (severity below threshold?)")
	}
	if alert == nil {
		return fmt.Errorf("[producer.AlertProducer.Produce]: alert is nil")
	}
	if alert.GetAlertId() == "" {
		return fmt.Errorf("[producer.AlertProducer.Produce]: alert_id is empty")
	}

	data, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(alert)
	if err != nil {
		return fmt.Errorf("[producer.AlertProducer.Produce](%s): marshal: %w", alert.GetAlertId(), err)
	}

	key := []byte(alert.GetTrackId())
	headers := map[string]string{
		"alert_id":      alert.GetAlertId(),
		"anomaly_type":  alert.GetAnomalyType().String(),
		"severity":      alert.GetSeverity().String(),
		"model_version": alert.GetModelVersion(),
	}

	if err := ap.producer.Produce(ctx, topic, key, data, headers); err != nil {
		return fmt.Errorf("[producer.AlertProducer.Produce](%s): %w", topic, err)
	}

	ap.logger.Info("alert produced",
		"alert_id", alert.GetAlertId(),
		"track_id", alert.GetTrackId(),
		"topic", topic,
		"severity", alert.GetSeverity().String(),
		"anomaly_type", alert.GetAnomalyType().String(),
		"confidence", alert.GetConfidenceScore(),
	)
	return nil
}

// Close shuts down the producer.
func (ap *AlertProducer) Close() error {
	return ap.producer.Close()
}
