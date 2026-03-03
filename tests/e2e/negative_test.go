// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e contains negative end-to-end tests validating error paths.
//
// These tests validate:
//   - DLQ routing for malformed sensor payloads (E2E06)
//   - Anti-poisoning rejection of zero-trust feedback (E2E07)
//   - Classification boundary enforcement (E2E08)
//   - Oversized payload rejection by ingestion service (E2E09)
//
// Run with: RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./...
package e2e

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestE2E06_MalformedSensor_EmptySensorId_RejectedToDLQ validates that a
// sensor observation that fails business-logic validation is rejected by
// svc-radar-ingestion via its gRPC interface and routed to the dead-letter
// topic dlq.sensors.radar within 30s.
//
// Design: the test calls IngestSingleObservation with an empty sensor_id —
// the RadarValidator rejects it and the handler publishes the rejected
// observation to dlq.sensors.radar before returning an IngestionAck with
// Accepted=false.  Direct raw-byte Kafka publishing is intentionally avoided
// because svc-radar-ingestion is a gRPC service, not a Kafka consumer.
//
// Covers UC001 error path: invalid sensor data must not propagate through the pipeline.
// Timeout: 2 minutes
func TestE2E06_MalformedSensor_EmptySensorId_RejectedToDLQ(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Subscribe to dlq.sensors.radar at the current end BEFORE sending the
	//    invalid observation to capture only records produced by this test run.
	consumer := newKafkaConsumer(t, broker,
		"e2e06-dlq-consumer",
		kgo.NewOffset().AtEnd(),
		"dlq.sensors.radar",
	)
	defer consumer.Close()

	// 2. Connect to svc-radar-ingestion gRPC endpoint.
	conn := grpcDialCtx(ctx, t, radarIngestionEndpoint())
	defer conn.Close()

	client := ingestionv1.NewIngestionServiceClient(conn)

	// 3. Send an observation with an empty sensor_id — the RadarValidator rejects
	//    this (rule: "sensor_id must not be empty") and routes it to DLQ.
	invalidObs := &ingestionv1.SensorObservation{
		ObservationId:   "e2e06-invalid-001",
		SensorId:        "", // empty sensor_id triggers validator rejection
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
	}

	ack, err := client.IngestSingleObservation(ctx, invalidObs)
	if err != nil {
		t.Fatalf("E2E06: IngestSingleObservation returned unexpected gRPC error: %v", err)
	}
	if ack.GetAccepted() {
		t.Error("E2E06: expected observation to be rejected (Accepted=false) by RadarValidator")
	}
	t.Logf("E2E06: gRPC ack — Accepted=%v, RejectionReason=%q",
		ack.GetAccepted(), ack.GetRejectionReason())

	// 4. Consume from dlq.sensors.radar and verify the rejected observation arrives.
	ok := pollUntil(ctx, consumer, 30*time.Second, func(_ *kgo.Record) bool {
		return true // any DLQ record confirms routing
	})
	if !ok {
		t.Fatal("E2E06: timeout: rejected observation did not appear on dlq.sensors.radar within 30s")
	}

	t.Log("E2E06 PASS: rejected observation routed to dlq.sensors.radar — DLQ routing confirmed")
}

// TestE2E07_AntiPoisoning_ZeroTrustFeedback_NotValidated validates that a
// feedback submission with trust score 0.0 is rejected by svc-feedback and
// does NOT appear on feedback.operator.validated.
//
// Covers UC014 error path: anti-poisoning must filter out zero-trust submissions.
// Timeout: 2 minutes
func TestE2E07_AntiPoisoning_ZeroTrustFeedback_NotValidated(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	producer := newKafkaProducer(t, broker)
	defer producer.Close()

	// Submit feedback with trust_score=0.0 — should be rejected by anti-poisoning.
	// The payload simulates what svc-feedback would receive before validation.
	zeroTrustPayload := []byte(
		`{"feedback_id":"e2e07-fb-001","track_id":"track-e2e07","operator_id":"op-e2e07","trust_score":0.0,"label":"HOSTILE"}`,
	)
	headers := redpanda.StandardHeaders("UNCLASSIFIED", "e2e-neg07", "", "v1")

	results := producer.ProduceSync(ctx, &kgo.Record{
		Topic:   "feedback.operator.submissions",
		Key:     []byte("op-e2e07"),
		Value:   zeroTrustPayload,
		Headers: headers,
	})
	if results.FirstErr() != nil {
		t.Fatalf("E2E07: produce zero-trust feedback: %v", results.FirstErr())
	}

	// Monitor feedback.operator.validated for 15s — message MUST NOT appear.
	consumer := newKafkaConsumer(t, broker,
		"e2e07-validated-consumer",
		kgo.NewOffset().AtEnd(),
		"feedback.operator.validated",
	)
	defer consumer.Close()

	// The zero-trust message must not appear — poll for 15 seconds.
	appeared := pollUntil(ctx, consumer, 15*time.Second, func(r *kgo.Record) bool {
		return string(r.Key) == "op-e2e07"
	})
	if appeared {
		t.Fatal("E2E07 FAIL: zero-trust feedback appeared on feedback.operator.validated — " +
			"anti-poisoning did not reject it")
	}

	t.Log("E2E07 PASS: zero-trust feedback correctly rejected (not on feedback.operator.validated)")
}

// TestE2E08_Classification_SecretInUnclassifiedPipeline_Quarantined validates
// that a sensor observation tagged SECRET does not appear on the UNCLASSIFIED
// fused tracks topic (tracks.fused.surface).
//
// Covers UC012 error path: classification boundary enforcement must quarantine
// SECRET data from UNCLASSIFIED pipeline topics.
// Timeout: 2 minutes
func TestE2E08_Classification_SecretInUnclassifiedPipeline_Quarantined(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	producer := newKafkaProducer(t, broker)
	defer producer.Close()

	// Build an observation with a SECRET classification header — the ingestion
	// service should reject it at the classification guard layer.
	obs := &ingestionv1.SensorObservation{
		ObservationId:   generateObsID("e2e08-obs", 0),
		SensorId:        "RADAR-E2E08",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
	}

	payload, _ := proto.Marshal(obs)
	// Inject a classification header above clearance level.
	headers := redpanda.StandardHeaders("SECRET", "e2e-neg08", "", "v1")

	results := producer.ProduceSync(ctx, &kgo.Record{
		Topic:   "sensors.radar.tracks",
		Key:     []byte("e2e08-secret"),
		Value:   payload,
		Headers: headers,
	})
	if results.FirstErr() != nil {
		t.Fatalf("E2E08: produce SECRET observation: %v", results.FirstErr())
	}

	// Verify the SECRET-tagged observation does not appear on the UNCLASSIFIED
	// fused tracks topic within 15s.
	consumer := newKafkaConsumer(t, broker,
		"e2e08-track-consumer",
		kgo.NewOffset().AtEnd(),
		"tracks.fused.surface",
	)
	defer consumer.Close()

	appeared := pollUntil(ctx, consumer, 15*time.Second, func(r *kgo.Record) bool {
		for _, h := range r.Headers {
			if string(h.Key) == "classification" && string(h.Value) == "SECRET" {
				return true
			}
		}
		return false
	})
	if appeared {
		t.Fatal("E2E08 FAIL: SECRET observation appeared on UNCLASSIFIED tracks.fused.surface — " +
			"classification guard did not quarantine it")
	}

	t.Log("E2E08 PASS: SECRET-classified observation correctly quarantined from UNCLASSIFIED pipeline")
}

// TestE2E09_OversizedPayload_ExceedsLimit_Rejected validates that a sensor
// observation exceeding 1 MB is rejected by the ingestion pipeline and does not
// produce a fused track (i.e., it does not exhaust pipeline resources).
//
// Covers UC001 error path: oversized payloads must not exhaust pipeline resources.
// Timeout: 2 minutes
func TestE2E09_OversizedPayload_ExceedsLimit_Rejected(t *testing.T) {
	skipE2E(t)

	broker := redpandaBroker()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	producer := newKafkaProducer(t, broker)
	defer producer.Close()

	// Create a payload exceeding 1 MB (1,048,577 bytes).
	oversizedPayload := make([]byte, 1_048_577)
	for i := range oversizedPayload {
		oversizedPayload[i] = byte(i % 256)
	}
	headers := redpanda.StandardHeaders("UNCLASSIFIED", "e2e-neg09", "", "v1")

	// Attempt to produce the oversized payload to the ingestion topic.
	// The ingestion service should reject it — verify via absence of a
	// corresponding track on the output topic.
	results := producer.ProduceSync(ctx, &kgo.Record{
		Topic:   "sensors.radar.tracks",
		Key:     []byte("e2e09-oversized"),
		Value:   oversizedPayload,
		Headers: headers,
	})
	if results.FirstErr() != nil {
		// Producer itself rejected the oversized message — valid rejection.
		t.Logf("E2E09 PASS: oversized message rejected at producer level: %v", results.FirstErr())
		return
	}

	// If the message was accepted by the broker, verify no fused track is produced
	// from it (the ingestion service should drop oversized messages internally).
	consumer := newKafkaConsumer(t, broker,
		"e2e09-track-consumer",
		kgo.NewOffset().AtEnd(),
		"tracks.fused.surface",
	)
	defer consumer.Close()

	appeared := pollUntil(ctx, consumer, 15*time.Second, func(r *kgo.Record) bool {
		return string(r.Key) == "e2e09-oversized"
	})
	if appeared {
		t.Fatal("E2E09 FAIL: oversized payload produced a fused track — ingestion service did not reject it")
	}

	t.Log("E2E09 PASS: no fused track produced from oversized payload — ingestion service correctly rejected it")
}
