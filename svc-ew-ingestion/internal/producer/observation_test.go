// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/producer"
)

func TestObservationProducer_Topic(t *testing.T) {
p := producer.NewObservationProducer(nil, "sensors.ew.intercepts")
if p.Topic() != "sensors.ew.intercepts" {
t.Errorf("expected topic sensors.ew.intercepts, got %s", p.Topic())
}
}

func TestObservationProducer_NilProducer(t *testing.T) {
p := producer.NewObservationProducer(nil, "test-topic")
if p == nil {
t.Fatal("expected non-nil ObservationProducer")
}
}

func TestObservation_RequiredFields(t *testing.T) {
obs := &ingestionv1.SensorObservation{
SensorId:       "EW-001",
ObservationId:  "obs-001",
SensorType:     commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if obs.GetSensorId() == "" {
t.Error("sensor_id should not be empty")
}
if obs.GetObservationId() == "" {
t.Error("observation_id should not be empty")
}
if obs.GetClassification() == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
t.Error("classification should not be unspecified")
}
}
