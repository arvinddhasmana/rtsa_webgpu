// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
	"context"

"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-elint-ingestion/internal/producer"
)

func TestObservationProducer_Topic(t *testing.T) {
p := producer.NewObservationProducer(nil, "sensors.elint.detections")
if p.Topic() != "sensors.elint.detections" {
t.Errorf("expected topic sensors.elint.detections, got %s", p.Topic())
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
SensorId:       "ELINT-001",
ObservationId:  "obs-001",
SensorType:     commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}

if obs.GetSensorId() == "" {
t.Error("sensor_id should not be empty")
}
if obs.GetObservationId() == "" {
t.Error("observation_id should not be empty")
}
}

func TestObservationProducer_Produce_Panic(t *testing.T) {
	p := producer.NewObservationProducer(nil, "topic")
	obs := &ingestionv1.SensorObservation{}
	defer func() {
		_ = recover()
	}()
	_ = p.Produce(context.Background(), obs)
}
