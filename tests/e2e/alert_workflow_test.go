// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e provides the alert workflow end-to-end test (E2E02).
package e2e

import (
	"context"
	"testing"
	"time"

	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// TestE2E02_AlertWorkflow_ProduceAndConsume_AlertReceivedByOperator validates
// the operator alert workflow:
//  1. Produce an anomaly alert to alerts.anomaly.critical
//  2. Consume the alert as an operator would
//  3. Verify alert has expected fields (alert_id, confidence_score)
//  4. Verify new alert is not pre-acknowledged
//
// Covers UC003 (alert generation) and UC004 (operator alert review).
// Timeout: 2 minutes
func TestE2E02_AlertWorkflow_ProduceAndConsume_AlertReceivedByOperator(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	producer := newKafkaProducer(t, broker)
	defer producer.Close()

	alert := &inferencev1.AnomalyAlert{
		AlertId:         "e2e02-alert-001",
		TrackId:         "track-e2e02",
		ConfidenceScore: 0.95,
		Explanation:     "speed anomaly detected in E2E test",
		ModelVersion:    "rules-v1.0.0",
		Acknowledged:    false,
	}

	payload, _ := proto.Marshal(alert)
	headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-anomaly-detection", "", "v1")

	results := producer.ProduceSync(ctx, &kgo.Record{
		Topic:   "alerts.anomaly.critical",
		Key:     []byte(alert.TrackId),
		Value:   payload,
		Headers: headers,
	})
	if results.FirstErr() != nil {
		t.Fatalf("E2E02: produce alert: %v", results.FirstErr())
	}

	// Consume alert — simulating operator reading it.
	consumer := newKafkaConsumer(t, broker,
		"e2e02-operator-consumer",
		kgo.NewOffset().AtStart(),
		"alerts.anomaly.critical",
	)
	defer consumer.Close()

	var received *inferencev1.AnomalyAlert
	ok := pollUntil(ctx, consumer, 20*time.Second, func(r *kgo.Record) bool {
		var a inferencev1.AnomalyAlert
		if err := proto.Unmarshal(r.Value, &a); err == nil {
			if a.AlertId == alert.AlertId {
				received = &a
				return true
			}
		}
		return false
	})

	if !ok || received == nil {
		t.Fatal("E2E02: timeout: expected alert not received on alerts.anomaly.critical within 20s")
	}
	if received.GetAcknowledged() {
		t.Error("E2E02: new alert should not be pre-acknowledged")
	}
	if got := received.GetConfidenceScore(); got != 0.95 {
		t.Errorf("E2E02: expected confidence_score=0.95, got=%v", got)
	}

	t.Logf("E2E02 PASS: alert %q received by operator, ready for acknowledgement (confidence=%.2f)",
		received.GetAlertId(), received.GetConfidenceScore())
}
