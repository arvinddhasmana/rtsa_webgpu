// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/producer"
)

func TestObservationProducer_Topic(t *testing.T) {
p := producer.NewObservationProducer(nil, "sensors.cyber.iocs")
if p.Topic() != "sensors.cyber.iocs" {
t.Errorf("expected topic sensors.cyber.iocs, got %s", p.Topic())
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
SensorId:       "CYBER-001",
ObservationId:  "obs-001",
SensorType:     commonv1.SensorType_SENSOR_TYPE_CYBER,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if obs.GetSensorId() == "" {
t.Error("sensor_id should not be empty")
}
if obs.GetObservationId() == "" {
t.Error("observation_id should not be empty")
}
}
