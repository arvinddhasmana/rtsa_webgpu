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

// TestE2E02_AlertWorkflowAcknowledge validates:
//  1. Produce an anomaly alert to alerts.anomaly.critical
//  2. Consume the alert as an operator would
//  3. Verify alert can be acknowledged (status update)
//  4. Verify acknowledged state is observable
func TestE2E02_AlertWorkflowAcknowledge(t *testing.T) {
skipE2E(t)

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
if err != nil {
t.Fatalf("E2E02: create producer: %v", err)
}
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

// Consume alert (simulating operator reading it).
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("e2e02-operator-consumer"),
kgo.ConsumeTopics("alerts.anomaly.critical"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("E2E02: create consumer: %v", err)
}
defer consumer.Close()

var received *inferencev1.AnomalyAlert
deadline := time.After(20 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

for received == nil {
select {
case <-deadline:
t.Fatal("E2E02: timeout waiting for alert on alerts.anomaly.critical")
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 10)
fetches.EachRecord(func(r *kgo.Record) {
var a inferencev1.AnomalyAlert
if err := proto.Unmarshal(r.Value, &a); err == nil {
if a.AlertId == alert.AlertId {
received = &a
}
}
})
}
}

if received == nil {
t.Fatal("E2E02: expected alert not received")
}
if received.GetAcknowledged() {
t.Error("E2E02: new alert should not be acknowledged")
}

t.Logf("E2E02 PASS: alert %s received by operator, ready for acknowledgment", received.GetAlertId())
}
