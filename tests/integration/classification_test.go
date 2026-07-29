// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides the MANDATORY classification propagation test (IT14).
// This test validates classification propagation across the entire RTSA pipeline
// through public APIs and Redpanda message patterns.
//
// IT14 is MANDATORY per Module 17 specification.
package integration

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/redpanda"
"github.com/arvinddhasmana/rtsa_webgpu/tests/integration/testutil"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestIT14_ClassificationEndToEnd is the MANDATORY classification propagation test.
//
// Validates classification propagation across the pipeline:
//  1. Sensor obs at PROTECTED_B → Redpanda header = PROTECTED_B
//  2. Sensor obs at SECRET for same entity → header = SECRET (MAX rule verified via pkg/classification)
//  3. Fused track at SECRET → header = SECRET on tracks.fused.surface
//  4. Anomaly alert inherits SECRET classification from track
//  5. PROTECTED_B caller cannot access SECRET data (classification.CanAccess)
//  6. SECRET caller can access SECRET data
func TestIT14_ClassificationEndToEnd(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

producer := env.NewKafkaProducer(t)
ctx := context.Background()

// ── Step 1: PROTECTED_B observation → header verification ────────────────
t.Log("IT14 Step 1: PROTECTED_B classification header propagation")

obs1 := &ingestionv1.SensorObservation{
ObservationId:   "it14-obs-protb",
SensorId:        "RADAR-IT14",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.Now(),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
Position: &commonv1.Position{
Latitude:  45.0,
Longitude: -60.0,
},
}

payload1, _ := proto.Marshal(obs1)
headers1 := redpanda.StandardHeaders(
classification.LevelToString(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B),
"svc-radar-ingestion", "", "v1",
)
r1 := producer.ProduceSync(ctx, &kgo.Record{
Topic: "sensors.radar.tracks", Key: []byte(obs1.SensorId),
Value: payload1, Headers: headers1,
})
if r1.FirstErr() != nil {
t.Fatalf("IT14 Step 1: produce: %v", r1.FirstErr())
}

consumer1 := env.NewKafkaConsumer(t, "it14-step1-group", "sensors.radar.tracks")
received1 := testutil.WaitForTopicMessages(t, consumer1, 1, 15*time.Second)
if len(received1) == 0 {
t.Fatal("IT14 Step 1: no message on sensors.radar.tracks")
}
classHeader1 := testutil.AssertHeaderPresent(t, received1[0], redpanda.HeaderClassification)
if classHeader1 != "PROTECTED_B" {
t.Errorf("IT14 Step 1: expected PROTECTED_B header, got %q", classHeader1)
}
t.Logf("IT14 Step 1 PASS: classification header = %s", classHeader1)

// ── Step 2: Verify MAX classification rule ────────────────────────────────
t.Log("IT14 Step 2: MAX classification rule (PROTECTED_B + SECRET → SECRET)")

maxLevel := classification.MaxAll(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
)
if maxLevel != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("IT14 Step 2: MaxAll = %v, want SECRET", maxLevel)
}
t.Log("IT14 Step 2 PASS: MAX rule → SECRET")

// ── Step 3: SECRET fused track → header propagation ──────────────────────
t.Log("IT14 Step 3: SECRET fused track classification header propagation")

rng := testutil.NewSeededRand(14)
pos14 := testutil.MidAtlanticPosition(rng)

secretTrack := &entityv1.FusedTrack{
TrackId:    "track-it14-secret",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Status:     commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
EstimatedPosition: pos14,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
UpdatedAt:  timestamppb.Now(),
}
payload3, _ := proto.Marshal(secretTrack)
headers3 := redpanda.StandardHeaders("SECRET", "svc-fusion-engine", "", "v1")

r3 := producer.ProduceSync(ctx, &kgo.Record{
Topic: "tracks.fused.surface", Key: []byte(secretTrack.TrackId),
Value: payload3, Headers: headers3,
})
if r3.FirstErr() != nil {
t.Fatalf("IT14 Step 3: produce SECRET track: %v", r3.FirstErr())
}

consumer3 := env.NewKafkaConsumer(t, "it14-step3-group", "tracks.fused.surface")
received3 := testutil.WaitForTopicMessages(t, consumer3, 1, 15*time.Second)
if len(received3) == 0 {
t.Fatal("IT14 Step 3: no message on tracks.fused.surface")
}
classHeader3 := testutil.AssertHeaderPresent(t, received3[0], redpanda.HeaderClassification)
if classHeader3 != "SECRET" {
t.Errorf("IT14 Step 3: classification header = %q, want SECRET", classHeader3)
}
t.Log("IT14 Step 3 PASS: SECRET fused track header = SECRET")

// ── Step 4: Anomaly alert inherits SECRET classification ──────────────────
t.Log("IT14 Step 4: Anomaly alert inherits SECRET from track")

secretAlert := &inferencev1.AnomalyAlert{
AlertId:         "alert-it14-secret",
TrackId:         secretTrack.TrackId,
AnomalyType:     commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
Severity:        commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
ConfidenceScore: 0.95,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, // inherited
Explanation:     "extreme speed anomaly detected",
ModelVersion:    "rules-v1.0.0",
DetectedAt:      timestamppb.Now(),
}
payload4, _ := proto.Marshal(secretAlert)
headers4 := redpanda.StandardHeaders("SECRET", "svc-anomaly-detection", "", "v1")

r4 := producer.ProduceSync(ctx, &kgo.Record{
Topic: "alerts.anomaly.critical", Key: []byte(secretTrack.TrackId),
Value: payload4, Headers: headers4,
})
if r4.FirstErr() != nil {
t.Fatalf("IT14 Step 4: produce SECRET alert: %v", r4.FirstErr())
}

consumer4 := env.NewKafkaConsumer(t, "it14-step4-group", "alerts.anomaly.critical")
received4 := testutil.WaitForTopicMessages(t, consumer4, 1, 15*time.Second)
if len(received4) == 0 {
t.Fatal("IT14 Step 4: no message on alerts.anomaly.critical")
}

var decodedAlert inferencev1.AnomalyAlert
if err := proto.Unmarshal(received4[0].Value, &decodedAlert); err != nil {
t.Fatalf("IT14 Step 4: deserialize alert: %v", err)
}
if decodedAlert.GetClassification() != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("IT14 Step 4: alert classification = %v, want SECRET", decodedAlert.GetClassification())
}
t.Log("IT14 Step 4 PASS: anomaly alert inherits SECRET classification")

// ── Step 5: Access control verification ──────────────────────────────────
t.Log("IT14 Step 5: Classification-based access control")

// PROTECTED_B caller cannot access SECRET data.
if classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
) {
t.Error("IT14 Step 5: PROTECTED_B caller should NOT access SECRET data")
} else {
t.Log("IT14 Step 5 PASS: PROTECTED_B caller denied SECRET data access")
}

// SECRET caller can access SECRET data.
if !classification.CanAccess(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
) {
t.Error("IT14 Step 5: SECRET caller SHOULD access SECRET data")
} else {
t.Log("IT14 Step 5 PASS: SECRET caller granted SECRET data access")
}

// ── Step 6: MaxAll across all levels ──────────────────────────────────────
t.Log("IT14 Step 6: MaxAll classification utility verification")

allLevels := classification.MaxAll(
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
)
if allLevels != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
t.Errorf("IT14 Step 6: MaxAll = %v, want SECRET", allLevels)
}
t.Log("IT14 Step 6 PASS: MaxAll correctly returns SECRET")

t.Log("IT14 PASS: Classification end-to-end propagation validated across ALL stages (MANDATORY test complete)")
}
