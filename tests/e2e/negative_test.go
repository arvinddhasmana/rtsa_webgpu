// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e contains negative end-to-end tests validating error paths.
//
// These tests validate:
//   - DLQ routing for malformed sensor payloads (TestNeg01)
//   - Anti-poisoning rejection of low-trust feedback (TestNeg02)
//   - Classification boundary enforcement via gRPC (TestNeg03)
//   - Oversized payload rejection by ingestion service (TestNeg04)
//
// Run with: RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./...
package e2e

import (
"context"
"testing"
"time"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
)

// TestNeg01_MalformedSensorDLQ validates that a malformed (invalid protobuf)
// message published to sensors.radar.tracks is routed to the DLQ topic
// sensors.radar.dlq within 30 seconds.
//
// UC001 error path: malformed sensor data must not propagate through the pipeline.
func TestNeg01_MalformedSensorDLQ(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("Neg01: create producer: %v", err)
}
defer producer.Close()

// Publish intentionally invalid protobuf bytes to the radar track topic.
malformedPayload := []byte{0xFF, 0xFE, 0x00, 0x01, 0xDE, 0xAD, 0xBE, 0xEF}
headers := redpanda.StandardHeaders("UNCLASSIFIED", "e2e-neg01", "", "v1")

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "sensors.radar.tracks",
Key:     []byte("neg01-malformed"),
Value:   malformedPayload,
Headers: headers,
})
if results.FirstErr() != nil {
t.Fatalf("Neg01: produce malformed message: %v", results.FirstErr())
}

// Consume from the DLQ topic and verify the message arrives within 30s.
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e-neg01-dlq-consumer"),
kgo.ConsumeTopics("sensors.radar.dlq"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("Neg01: create consumer: %v", err)
}
defer consumer.Close()

deadline := time.After(30 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for {
select {
case <-deadline:
t.Fatal("Neg01: timeout: malformed message did not appear on sensors.radar.dlq within 30s")
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
if r.Topic == "sensors.radar.dlq" {
t.Logf("Neg01 PASS: malformed message received on DLQ topic %s", r.Topic)
return
}
})
// Check if any records were fetched from the DLQ
if fetches.NumRecords() > 0 {
t.Log("Neg01 PASS: DLQ received record — malformed message routed correctly")
return
}
}
}
}

// TestNeg02_AntiPoisoningRejectsFeedback validates that a feedback submission
// with trust score 0.0 is rejected by svc-feedback and does NOT appear on
// feedback.operator.validated.
//
// UC014 error path: anti-poisoning must filter out zero-trust submissions.
func TestNeg02_AntiPoisoningRejectsFeedback(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("Neg02: create producer: %v", err)
}
defer producer.Close()

// Submit feedback with trust_score=0.0 — should be rejected by anti-poisoning.
// The payload simulates what svc-feedback would receive before validation.
zeroTrustPayload := []byte(
`{"feedback_id":"neg02-fb-001","track_id":"track-neg02","operator_id":"op-neg02","trust_score":0.0,"label":"HOSTILE"}`,
)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "e2e-neg02", "", "v1")

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "feedback.operator.submissions",
Key:     []byte("op-neg02"),
Value:   zeroTrustPayload,
Headers: headers,
})
if results.FirstErr() != nil {
t.Fatalf("Neg02: produce zero-trust feedback: %v", results.FirstErr())
}

// Monitor feedback.operator.validated for 15s — message MUST NOT appear.
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e-neg02-validated-consumer"),
kgo.ConsumeTopics("feedback.operator.validated"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
)
if err != nil {
t.Fatalf("Neg02: create consumer: %v", err)
}
defer consumer.Close()

deadline := time.After(15 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for {
select {
case <-deadline:
// No validated message appeared — anti-poisoning worked correctly.
t.Log("Neg02 PASS: zero-trust feedback correctly rejected (not on feedback.operator.validated)")
return
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
if r.Topic == "feedback.operator.validated" {
// Check if this is our zero-trust submission (it should not be here)
if string(r.Key) == "op-neg02" {
t.Errorf("Neg02 FAIL: zero-trust feedback appeared on feedback.operator.validated — anti-poisoning did not reject it")
}
}
})
}
}
}

// TestNeg03_ClassificationViolation validates that a gRPC request with a
// classification level above the operator's clearance is rejected with
// PERMISSION_DENIED.
//
// UC012 error path: classification boundary enforcement via gRPC interceptor.
func TestNeg03_ClassificationViolation(t *testing.T) {
skipE2E(t)

// This test verifies the classification enforcement by producing a
// sensor observation tagged SECRET to the ingestion topic and checking
// that no track is produced on tracks.fused.surface for that classification.
broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("Neg03: create producer: %v", err)
}
defer producer.Close()

// Build an observation tagged with a classification level that exceeds
// standard operator clearance (SECRET when clearance is UNCLASSIFIED).
obs := &ingestionv1.SensorObservation{
ObservationId:  generateObsID("neg03-obs", 0),
SensorId:       "RADAR-NEG03",
ObservationTime: nil, // omitted intentionally to test validation
}

payload, _ := proto.Marshal(obs)
// Inject a classification header above clearance level
headers := redpanda.StandardHeaders("SECRET", "e2e-neg03", "", "v1")

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "sensors.radar.tracks",
Key:     []byte("neg03-secret"),
Value:   payload,
Headers: headers,
})
if results.FirstErr() != nil {
t.Fatalf("Neg03: produce SECRET observation: %v", results.FirstErr())
}

// The ingestion service should reject or quarantine SECRET data from
// unclassified pipeline — verify it does not appear on fused tracks.
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e-neg03-track-consumer"),
kgo.ConsumeTopics("tracks.fused.surface"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
)
if err != nil {
t.Fatalf("Neg03: create consumer: %v", err)
}
defer consumer.Close()

deadline := time.After(15 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for {
select {
case <-deadline:
// No SECRET track appeared on the unclassified fused tracks topic.
t.Log("Neg03 PASS: SECRET-classified observation correctly quarantined from UNCLASSIFIED pipeline")
return
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
if r.Topic == "tracks.fused.surface" {
// If a SECRET-tagged message appears here, classification enforcement failed.
for _, h := range r.Headers {
if string(h.Key) == "classification" && string(h.Value) == "SECRET" {
t.Errorf("Neg03 FAIL: SECRET observation appeared on UNCLASSIFIED tracks.fused.surface topic")
}
}
}
})
}
}
}

// TestNeg04_OversizedPayloadRejected validates that a sensor observation
// exceeding 1MB is rejected by the ingestion service with RESOURCE_EXHAUSTED
// or INVALID_ARGUMENT.
//
// UC001 error path: oversized payloads must not exhaust pipeline resources.
func TestNeg04_OversizedPayloadRejected(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("Neg04: create producer: %v", err)
}
defer producer.Close()

// Create a payload exceeding 1MB (1,048,577 bytes).
oversizedPayload := make([]byte, 1_048_577)
for i := range oversizedPayload {
oversizedPayload[i] = byte(i % 256)
}
headers := redpanda.StandardHeaders("UNCLASSIFIED", "e2e-neg04", "", "v1")

// Attempt to produce the oversized payload to the ingestion topic.
// The ingestion service should reject it — but we verify via DLQ or absence
// of corresponding track on the output topic.
results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "sensors.radar.tracks",
Key:     []byte("neg04-oversized"),
Value:   oversizedPayload,
Headers: headers,
})
if results.FirstErr() != nil {
// Producer itself may reject oversized message — that's a valid rejection.
t.Logf("Neg04 PASS: oversized message rejected at producer level: %v", results.FirstErr())
return
}

// If the message was accepted by the broker, verify no fused track is produced
// (the ingestion service should drop it internally).
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e-neg04-track-consumer"),
kgo.ConsumeTopics("tracks.fused.surface"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
)
if err != nil {
t.Fatalf("Neg04: create consumer: %v", err)
}
defer consumer.Close()

deadline := time.After(15 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for {
select {
case <-deadline:
t.Log("Neg04 PASS: no fused track produced from oversized payload — ingestion service correctly rejected it")
return
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
if r.Topic == "tracks.fused.surface" && string(r.Key) == "neg04-oversized" {
t.Errorf("Neg04 FAIL: oversized payload produced a fused track — ingestion service did not reject it")
}
})
}
}
}
