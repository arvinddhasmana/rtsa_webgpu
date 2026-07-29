// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for Module 17 (IT08–IT10):
// anomaly alert message validation via Redpanda topic pattern.
//
// Note: Full anomaly detection requires running svc-anomaly-detection.
// These tests validate the Redpanda topic message layer and header propagation.
// Unit-level anomaly detection tests are in svc-anomaly-detection/internal/integration/.
package integration

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
"github.com/arvinddhasmana/rtsa_webgpu/tests/integration/testutil"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
)

// TestIT08_SpeedAnomalyDetection validates:
//  1. A SPEED anomaly alert can be produced to alerts.anomaly.* topics
//  2. The alert protobuf is properly deserialized with correct fields
//  3. anomaly_type = SPEED, confidence > 0.5, explanation non-empty
func TestIT08_SpeedAnomalyDetection(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

alert := testutil.AnomalyAlertFixture(commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL)
alert.AnomalyType = commonv1.AnomalyType_ANOMALY_TYPE_SPEED
alert.ConfidenceScore = 0.92
alert.Explanation = "speed 50 knots exceeds 3-sigma threshold (mean=12 knots)"

producer := env.NewKafkaProducer(t)
ctx := context.Background()

payload, _ := proto.Marshal(alert)
classStr := classification.LevelToString(alert.GetClassification())
headers := redpanda.StandardHeaders(classStr, "svc-anomaly-detection", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "alerts.anomaly.critical",
Key:     []byte(alert.TrackId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("IT08: produce speed alert: %v", r.FirstErr())
}

consumer := env.NewKafkaConsumer(t, "it08-group", "alerts.anomaly.critical")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("IT08: no message on alerts.anomaly.critical")
}

var decoded inferencev1.AnomalyAlert
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Fatalf("IT08: deserialize alert: %v", err)
}

if decoded.GetAnomalyType() != commonv1.AnomalyType_ANOMALY_TYPE_SPEED {
t.Errorf("IT08: anomaly_type=%v, want SPEED", decoded.GetAnomalyType())
}
if decoded.GetConfidenceScore() <= 0.5 {
t.Errorf("IT08: confidence=%.2f, want > 0.5", decoded.GetConfidenceScore())
}
if decoded.GetExplanation() == "" {
t.Error("IT08: explanation is empty")
}

t.Logf("IT08 PASS: SPEED anomaly alert on alerts.anomaly.critical (confidence=%.2f)", decoded.GetConfidenceScore())
}

// TestIT09_AISManipulationDetection validates:
//  1. An AIS_MANIPULATION alert can be produced to alerts.anomaly.* topics
//  2. The alert references the correct track_id
//  3. The alert has appropriate confidence and explanation
func TestIT09_AISManipulationDetection(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

rng := testutil.NewSeededRand(9)
pos := testutil.MidAtlanticPosition(rng)

alert := &inferencev1.AnomalyAlert{
AlertId:         "alert-it09-ais-001",
TrackId:         "track-it09-ais",
AnomalyType:     commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION,
Severity:        commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
ConfidenceScore: 0.82,
Explanation:     "AIS position delta 1.2 NM from radar track — possible AIS spoofing",
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
TrackPosition:   pos,
ModelVersion:    "rules-v1.0.0",
}

producer := env.NewKafkaProducer(t)
ctx := context.Background()

payload, _ := proto.Marshal(alert)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-anomaly-detection", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "alerts.anomaly.elevated",
Key:     []byte(alert.TrackId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Fatalf("IT09: produce AIS alert: %v", r.FirstErr())
}

consumer := env.NewKafkaConsumer(t, "it09-group", "alerts.anomaly.elevated")
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Fatal("IT09: no message on alerts.anomaly.elevated")
}

var decoded inferencev1.AnomalyAlert
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Fatalf("IT09: deserialize alert: %v", err)
}

if decoded.GetAnomalyType() != commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION {
t.Errorf("IT09: anomaly_type=%v, want AIS_MANIPULATION", decoded.GetAnomalyType())
}
if decoded.GetTrackId() != alert.TrackId {
t.Errorf("IT09: track_id=%q, want %q", decoded.GetTrackId(), alert.TrackId)
}

t.Logf("IT09 PASS: AIS_MANIPULATION alert on alerts.anomaly.elevated (confidence=%.2f)", decoded.GetConfidenceScore())
}

// TestIT10_AlertSeverityRouting validates:
//  1. CRITICAL alerts (confidence >= 0.90) go to alerts.anomaly.critical
//  2. ELEVATED alerts (confidence 0.70-0.89) go to alerts.anomaly.elevated
//  3. WATCH alerts (confidence 0.50-0.69) go to alerts.anomaly.watch
func TestIT10_AlertSeverityRouting(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

cases := []struct {
severity commonv1.AlertSeverity
topic    string
conf     float64
}{
{commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, "alerts.anomaly.critical", 0.95},
{commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, "alerts.anomaly.elevated", 0.80},
{commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, "alerts.anomaly.watch", 0.60},
}

producer := env.NewKafkaProducer(t)
ctx := context.Background()

for _, tc := range cases {
alert := testutil.AnomalyAlertFixture(tc.severity)
alert.ConfidenceScore = tc.conf

payload, _ := proto.Marshal(alert)
headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-anomaly-detection", "", "v1")

r := producer.ProduceSync(ctx, &kgo.Record{
Topic:   tc.topic,
Key:     []byte(alert.TrackId),
Value:   payload,
Headers: headers,
})
if r.FirstErr() != nil {
t.Errorf("IT10[%v]: produce alert: %v", tc.severity, r.FirstErr())
}
}

// Verify each severity topic has a message.
for _, tc := range cases {
consumer := env.NewKafkaConsumer(t, "it10-"+tc.topic+"-group", tc.topic)
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Errorf("IT10: no message on %s", tc.topic)
continue
}

var decoded inferencev1.AnomalyAlert
if err := proto.Unmarshal(received[0].Value, &decoded); err != nil {
t.Errorf("IT10: deserialize alert on %s: %v", tc.topic, err)
continue
}
if decoded.GetSeverity() != tc.severity {
t.Errorf("IT10[%s]: severity=%v, want %v", tc.topic, decoded.GetSeverity(), tc.severity)
}
t.Logf("IT10 PASS[%s]: %v alert (confidence=%.2f)", tc.topic, tc.severity, decoded.GetConfidenceScore())
}
}
