// CLASSIFICATION: UNCLASSIFIED
package sensor_test

import (
"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/generator"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/sensor"
)

func TestGenerateEWObservation_ValidFields(t *testing.T) {
e := &generator.SimEntity{
ID:         "AIR-EW-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
Position: generator.Position{
Lat:     45.5,
Lon:     -61.0,
AltM:    5000,
SpeedKn: 300.0,
Heading: 90.0,
},
}

obs := sensor.GenerateEWObservation(e, "EW-SIM-001")

if obs == nil {
t.Fatal("EW observation should not be nil")
}
if obs.SensorType != commonv1.SensorType_SENSOR_TYPE_EW_SIGINT {
t.Errorf("expected SENSOR_TYPE_EW_SIGINT, got %v", obs.SensorType)
}
if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("must be UNCLASSIFIED, got %v", obs.Classification)
}

ewPl, ok := obs.SensorData.(*ingestionv1.SensorObservation_EwSigint)
if !ok {
t.Fatal("sensor_data should be EWIntercept")
}
ew := ewPl.EwSigint

if ew.EmitterId == "" {
t.Error("emitter_id must be non-empty")
}
if ew.FrequencyMhz < 0.5 || ew.FrequencyMhz > 40000 {
t.Errorf("frequency_mhz must be 0.5-40000, got %f", ew.FrequencyMhz)
}
if ew.BearingDegrees < 0 || ew.BearingDegrees > 360 {
t.Errorf("bearing_degrees must be 0-360, got %f", ew.BearingDegrees)
}
if ew.Confidence < 0.0 || ew.Confidence > 1.0 {
t.Errorf("confidence must be 0-1, got %f", ew.Confidence)
}
if ew.ModulationType == "" {
t.Error("modulation_type must be non-empty")
}
}

func TestGenerateEWObservation_SensorIDSet(t *testing.T) {
e := &generator.SimEntity{
ID:         "AIR-EW-002",
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
Position:   generator.Position{Lat: 45.0, Lon: -60.0, SpeedKn: 200.0, Heading: 0},
}
obs := sensor.GenerateEWObservation(e, "EW-SIM-002")
if obs.SensorId != "EW-SIM-002" {
t.Errorf("expected sensor_id EW-SIM-002, got %q", obs.SensorId)
}
}
