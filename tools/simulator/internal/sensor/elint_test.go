// CLASSIFICATION: UNCLASSIFIED
package sensor_test

import (
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/generator"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/sensor"
)

var validScanTypes = map[string]bool{
"circular":         true,
"sector":           true,
"track-while-scan": true,
}

func TestGenerateELINTObservation_ValidFields(t *testing.T) {
e := &generator.SimEntity{
ID:         "SURF-ELINT-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Position: generator.Position{
Lat:     44.0,
Lon:     -62.0,
SpeedKn: 15.0,
Heading: 270.0,
},
}

obs := sensor.GenerateELINTObservation(e, "ELINT-SIM-001")

if obs == nil {
t.Fatal("ELINT observation should not be nil")
}
if obs.SensorType != commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT {
t.Errorf("expected SENSOR_TYPE_ELINT_COMINT, got %v", obs.SensorType)
}
if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("must be UNCLASSIFIED, got %v", obs.Classification)
}

elintPl, ok := obs.SensorData.(*ingestionv1.SensorObservation_ElintComint)
if !ok {
t.Fatal("sensor_data should be ELINTDetection")
}
el := elintPl.ElintComint

if el.EmitterId == "" {
t.Error("emitter_id must be non-empty")
}
if el.RadarType == "" {
t.Error("radar_type must be non-empty")
}
if el.FrequencyMhz < 0.5 || el.FrequencyMhz > 40000 {
t.Errorf("frequency_mhz must be 0.5-40000, got %f", el.FrequencyMhz)
}
if !validScanTypes[el.ScanType] {
t.Errorf("scan_type must be circular|sector|track-while-scan, got %q", el.ScanType)
}
if el.CepMeters <= 0 {
t.Errorf("cep_meters must be >0, got %f", el.CepMeters)
}
if el.Confidence < 0 || el.Confidence > 1 {
t.Errorf("confidence must be 0-1, got %f", el.Confidence)
}
if el.ContentClassification == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
t.Error("content_classification must not be UNSPECIFIED")
}
}
