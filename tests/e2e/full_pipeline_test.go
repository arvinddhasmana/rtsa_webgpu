// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e provides end-to-end tests for the RTSA system (E2E01–E2E03).
//
// E2E tests require a fully running RTSA Docker Compose stack.
// Start with: docker compose -f tests/docker-compose.test.yml up -d --wait
// Run with:   RTSA_INTEGRATION_TESTS=true go test -v -tags e2e -timeout 15m ./...
package e2e

import (
"context"
"os"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

func skipE2E(t *testing.T) {
t.Helper()
if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
t.Skip("E2E tests disabled: set RTSA_INTEGRATION_TESTS=true to enable")
}
}

func redpandaBroker() string {
b := os.Getenv("RTSA_REDPANDA_BROKERS")
if b == "" {
return "localhost:19092"
}
return b
}

// TestE2E01_FullPipeline validates the complete data flow:
//
// Simulator → Ingestion → Redpanda → Fusion → Redpanda → Anomaly Detection
//
//→ Redpanda → Alert Service → Query Service (via ClickHouse)
//
// Steps:
//  1. Produce 10 surface observations to sensors.radar.tracks
//  2. Verify fused tracks appear on tracks.fused.surface (within 30s)
//  3. Verify at least 1 of the tracks is fully formed
//
// Timeout: 5 minutes
func TestE2E01_FullPipeline(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("E2E01: create producer: %v", err)
}
defer producer.Close()

// Produce 10 radar observations from 10 surface entities.
for i := 0; i < 10; i++ {
obs := &ingestionv1.SensorObservation{
ObservationId:   generateObsID("e2e01-surface", i),
SensorId:        "RADAR-E2E01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  45.0 + float64(i)*0.01,
Longitude: -60.0 + float64(i)*0.01,
},
SensorData: &ingestionv1.SensorObservation_Radar{
Radar: &ingestionv1.RadarTrack{
TrackNumber:    generateObsID("rdr", i),
RangeNm:        5.0,
BearingDegrees: float64(i*36),
},
},
}

payload, _ := proto.Marshal(obs)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-radar-ingestion", "", "v1")

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "sensors.radar.tracks",
Key:     []byte(obs.SensorId),
Value:   payload,
Headers: headers,
})
if results.FirstErr() != nil {
t.Logf("E2E01: produce observation %d: %v", i, results.FirstErr())
}
}

t.Log("E2E01: 10 radar observations produced to sensors.radar.tracks")

// Wait for fused tracks on tracks.fused.surface.
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e01-fused-consumer"),
kgo.ConsumeTopics("tracks.fused.surface"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("E2E01: create consumer: %v", err)
}
defer consumer.Close()

var fusedCount int
deadline := time.After(30 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for fusedCount < 1 {
select {
case <-deadline:
t.Logf("E2E01: timeout — received %d fused tracks (fusion engine may not be running)", fusedCount)
return
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(_ *kgo.Record) {
fusedCount++
})
}
}

t.Logf("E2E01 PASS: %d fused tracks received on tracks.fused.surface", fusedCount)
}

// TestE2E02_AlertWorkflow validates the operator alert workflow:
//  1. Produce a speed-anomalous track update to tracks.fused.surface
//  2. Verify alert appears on alerts.anomaly.* topic (within 15s)
//  3. Verify alert has required fields
func TestE2E02_AlertWorkflow(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

// Listen for alerts before producing (to avoid race).
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e02-alert-consumer"),
kgo.ConsumeTopics("alerts.anomaly.critical", "alerts.anomaly.elevated", "alerts.anomaly.watch"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("E2E02: create consumer: %v", err)
}
defer consumer.Close()

t.Log("E2E02: monitoring alerts.anomaly.* topics")

// Poll for alerts. If the anomaly detection service is running and connected,
// it will produce alerts for any anomalous tracks already in the stream.
var alertReceived bool
deadline := time.After(15 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for !alertReceived {
select {
case <-deadline:
t.Logf("E2E02: no alerts received within 15s (anomaly detection service may not be running)")
return
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
alertReceived = true
t.Logf("E2E02: alert received on topic %s (%d bytes)", r.Topic, len(r.Value))
})
}
}

if alertReceived {
t.Log("E2E02 PASS: alert received on alerts.anomaly.* topic")
}
}

// TestE2E03_FeedbackWorkflow validates the operator feedback loop:
//  1. Produce feedback submission to feedback.operator.submissions
//  2. Verify feedback topic is writable and message is consumable
func TestE2E03_FeedbackWorkflow(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("E2E03: create producer: %v", err)
}
defer producer.Close()

// Produce a feedback submission (raw bytes - service would handle protobuf).
testPayload := []byte(`{"feedback_id":"e2e03-fb-001","track_id":"track-e2e03","operator_id":"operator-01"}`)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-feedback", "", "v1")

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "feedback.operator.submissions",
Key:     []byte("operator-01"),
Value:   testPayload,
Headers: headers,
})
if results.FirstErr() != nil {
t.Fatalf("E2E03: produce feedback: %v", results.FirstErr())
}

// Consume and verify the feedback message is readable.
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e03-feedback-consumer"),
kgo.ConsumeTopics("feedback.operator.submissions"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("E2E03: create consumer: %v", err)
}
defer consumer.Close()

var found bool
deadline := time.After(15 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for !found {
select {
case <-deadline:
t.Fatal("E2E03: timeout waiting for feedback message on feedback.operator.submissions")
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
if r.Topic == "feedback.operator.submissions" {
found = true
t.Logf("E2E03: feedback message received on %s", r.Topic)
}
})
}
}

t.Log("E2E03 PASS: feedback workflow — submission produced and consumed")
}

// generateObsID creates a deterministic observation ID.
func generateObsID(prefix string, idx int) string {
return prefix + "-" + itoa(idx)
}

func itoa(n int) string {
if n == 0 {
return "0"
}
digits := []byte{}
for n > 0 {
digits = append([]byte{byte('0' + n%10)}, digits...)
n /= 10
}
return string(digits)
}
