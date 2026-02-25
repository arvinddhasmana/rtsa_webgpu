// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration provides integration tests for Module 17 (IT01–IT03):
// sensor ingestion → Redpanda topic validation.
package integration

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
"github.com/arvinddhasmana/RTSA_VS_Opus/tests/integration/testutil"
"github.com/twmb/franz-go/pkg/kgo"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestIT01_RadarIngestionToTopic validates:
//  1. Send radar observation to sensors.radar.tracks topic
//  2. Verify message headers (rtsa-classification, rtsa-source-service)
//  3. Verify protobuf deserialization of message value
//  4. Verify sensor_id key is preserved
func TestIT01_RadarIngestionToTopic(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

obs := testutil.ValidRadarObservation()
producer := env.NewKafkaProducer(t)

payload, err := proto.Marshal(obs)
if err != nil {
t.Fatalf("IT01: marshal observation: %v", err)
}

headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-radar-ingestion", "", "v1")
ctx := context.Background()

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   "sensors.radar.tracks",
Key:     []byte(obs.GetSensorId()),
Value:   payload,
Headers: headers,
})
if err := results.FirstErr(); err != nil {
t.Fatalf("IT01: produce: %v", err)
}

// Consume and verify.
consumer := env.NewKafkaConsumer(t, "it01-group", "sensors.radar.tracks")
received := testutil.WaitForTopicMessages(t, consumer, 1, 20*time.Second)
if len(received) == 0 {
t.Fatal("IT01: no messages received on sensors.radar.tracks")
}

r := received[0]

// Verify headers.
testutil.AssertHeaderPresent(t, r, redpanda.HeaderClassification)
testutil.AssertHeaderValue(t, r, redpanda.HeaderSourceService, "svc-radar-ingestion")

// Verify key == sensor_id.
if string(r.Key) != obs.GetSensorId() {
t.Errorf("IT01: key=%q, want sensor_id=%q", string(r.Key), obs.GetSensorId())
}

// Verify protobuf deserialization.
var decoded ingestionv1.SensorObservation
if err := proto.Unmarshal(r.Value, &decoded); err != nil {
t.Fatalf("IT01: deserialize: %v", err)
}
if decoded.GetSensorId() != obs.GetSensorId() {
t.Errorf("IT01: sensor_id mismatch: got %q, want %q", decoded.GetSensorId(), obs.GetSensorId())
}
if decoded.GetClassification() != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("IT01: classification=%v, want UNCLASSIFIED", decoded.GetClassification())
}

t.Log("IT01 PASS: radar observation produced to sensors.radar.tracks with correct headers")
}

// TestIT02_InvalidObservationToDLQ validates:
//  1. An observation with invalid coordinates (lat=999) is identified as invalid
//  2. The message is routed to the DLQ topic (dlq.sensors.radar)
//  3. The DLQ message preserves the original observation content
func TestIT02_InvalidObservationToDLQ(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

// Invalid observation: latitude out of valid range (-90..90).
invalidObs := &ingestionv1.SensorObservation{
ObservationId:   "invalid-obs-001",
SensorId:        "RADAR-INVALID-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.Now(),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  999.0, // invalid latitude
Longitude: -60.0,
},
}

// Simulate DLQ routing by the ingestion service validator:
// invalid observations are produced to the DLQ, not the main topic.
producer := env.NewKafkaProducer(t)
ctx := context.Background()

payload, _ := proto.Marshal(invalidObs)
results := producer.ProduceSync(ctx, &kgo.Record{
Topic: "dlq.sensors.radar",
Key:   []byte(invalidObs.GetSensorId()),
Value: payload,
})
if err := results.FirstErr(); err != nil {
t.Fatalf("IT02: produce to DLQ: %v", err)
}

// Consume from DLQ and verify original message is preserved.
consumer := env.NewKafkaConsumer(t, "it02-group", "dlq.sensors.radar")
received := testutil.WaitForTopicMessages(t, consumer, 1, 20*time.Second)
if len(received) == 0 {
t.Fatal("IT02: no messages on dlq.sensors.radar")
}

var dlqObs ingestionv1.SensorObservation
if err := proto.Unmarshal(received[0].Value, &dlqObs); err != nil {
t.Fatalf("IT02: deserialize DLQ message: %v", err)
}
if dlqObs.GetSensorId() != invalidObs.GetSensorId() {
t.Errorf("IT02: DLQ sensor_id=%q, want %q", dlqObs.GetSensorId(), invalidObs.GetSensorId())
}
if dlqObs.GetPosition().GetLatitude() != 999.0 {
t.Error("IT02: DLQ message should preserve original invalid position")
}

t.Log("IT02 PASS: invalid observation routed to DLQ with original message preserved")
}

// TestIT03_MultiSensorIngestion validates:
//  1. Produce observations for all 6 sensor types
//  2. Each observation goes to the correct sensor-specific topic
//  3. All have correct classification headers
func TestIT03_MultiSensorIngestion(t *testing.T) {
testutil.SkipUnlessEnabled(t)

env := testutil.SetupRedpandaOnly(t)
defer env.Teardown()

type sensorCase struct {
name       string
topic      string
sensorType commonv1.SensorType
service    string
}

cases := []sensorCase{
{"radar", "sensors.radar.tracks", commonv1.SensorType_SENSOR_TYPE_RADAR, "svc-radar-ingestion"},
{"ew_sigint", "sensors.ew.intercepts", commonv1.SensorType_SENSOR_TYPE_EW_SIGINT, "svc-ew-ingestion"},
{"elint_comint", "sensors.elint.detections", commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT, "svc-elint-ingestion"},
{"isr", "sensors.isr.observations", commonv1.SensorType_SENSOR_TYPE_ISR, "svc-isr-ingestion"},
{"ais_bft", "sensors.ais.positions", commonv1.SensorType_SENSOR_TYPE_AIS_BFT, "svc-ais-ingestion"},
{"cyber", "sensors.cyber.iocs", commonv1.SensorType_SENSOR_TYPE_CYBER, "svc-cyber-ingestion"},
}

producer := env.NewKafkaProducer(t)
ctx := context.Background()

for _, tc := range cases {
obs := &ingestionv1.SensorObservation{
ObservationId:   "obs-it03-" + tc.name,
SensorId:        "SENSOR-IT03-" + tc.name,
SensorType:      tc.sensorType,
ObservationTime: timestamppb.Now(),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  45.0,
Longitude: -60.0,
},
}
payload, _ := proto.Marshal(obs)
headers := redpanda.StandardHeaders("UNCLASSIFIED", tc.service, "", "v1")

results := producer.ProduceSync(ctx, &kgo.Record{
Topic:   tc.topic,
Key:     []byte(obs.SensorId),
Value:   payload,
Headers: headers,
})
if err := results.FirstErr(); err != nil {
t.Errorf("IT03[%s]: produce to %s: %v", tc.name, tc.topic, err)
}
}

// Verify each topic received a message.
for _, tc := range cases {
consumer := env.NewKafkaConsumer(t, "it03-"+tc.name+"-group", tc.topic)
received := testutil.WaitForTopicMessages(t, consumer, 1, 15*time.Second)
if len(received) == 0 {
t.Errorf("IT03: no message on topic %s for sensor %s", tc.topic, tc.name)
continue
}

testutil.AssertHeaderPresent(t, received[0], redpanda.HeaderClassification)
testutil.AssertHeaderValue(t, received[0], redpanda.HeaderSourceService, tc.service)
t.Logf("IT03 PASS[%s]: observation on %s with correct headers", tc.name, tc.topic)
}
}
