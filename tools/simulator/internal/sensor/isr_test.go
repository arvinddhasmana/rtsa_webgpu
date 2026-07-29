// CLASSIFICATION: UNCLASSIFIED
package sensor_test

import (
"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/generator"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/sensor"
)

var validISRSensorNames = map[string]bool{
"EO": true, "IR": true, "SAR": true, "MTI": true,
}

func TestGenerateISRObservation_ValidFields(t *testing.T) {
e := &generator.SimEntity{
ID:         "AIR-ISR-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
Position: generator.Position{
Lat:     46.0,
Lon:     -58.0,
AltM:    8000,
SpeedKn: 350.0,
Heading: 0.0,
},
}

obs := sensor.GenerateISRObservation(e, "ISR-SIM-001")

if obs == nil {
t.Fatal("ISR observation should not be nil")
}
if obs.SensorType != commonv1.SensorType_SENSOR_TYPE_ISR {
t.Errorf("expected SENSOR_TYPE_ISR, got %v", obs.SensorType)
}
if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("must be UNCLASSIFIED, got %v", obs.Classification)
}

isrPl, ok := obs.SensorData.(*ingestionv1.SensorObservation_Isr)
if !ok {
t.Fatal("sensor_data should be ISRObservation")
}
isr := isrPl.Isr

if isr.PlatformId == "" {
t.Error("platform_id must be non-empty")
}
if !validISRSensorNames[isr.SensorName] {
t.Errorf("sensor_name must be EO|IR|SAR|MTI, got %q", isr.SensorName)
}
if isr.ImageId == "" {
t.Error("image_id must be non-empty")
}
// Coverage polygon must have at least 3 vertices.
if len(isr.CoveragePolygon) < 3 {
t.Errorf("coverage_polygon must have ≥3 vertices, got %d", len(isr.CoveragePolygon))
}
// All polygon vertices must be valid lat/lon.
for i, v := range isr.CoveragePolygon {
if v.Latitude < -90 || v.Latitude > 90 {
t.Errorf("polygon vertex %d lat %f out of range", i, v.Latitude)
}
if v.Longitude < -180 || v.Longitude > 180 {
t.Errorf("polygon vertex %d lon %f out of range", i, v.Longitude)
}
}
// Detections confidence must be 0-1.
for i, d := range isr.Detections {
if d.Confidence < 0 || d.Confidence > 1 {
t.Errorf("detection %d confidence must be 0-1, got %f", i, d.Confidence)
}
}
}

func TestGenerateISRObservation_SequentialImageIDs(t *testing.T) {
e := &generator.SimEntity{
ID:         "AIR-ISR-002",
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
Position:   generator.Position{Lat: 45.0, Lon: -60.0, SpeedKn: 300.0, AltM: 5000},
}

obs1 := sensor.GenerateISRObservation(e, "ISR-SIM-SEQ")
obs2 := sensor.GenerateISRObservation(e, "ISR-SIM-SEQ")

id1 := obs1.SensorData.(*ingestionv1.SensorObservation_Isr).Isr.ImageId
id2 := obs2.SensorData.(*ingestionv1.SensorObservation_Isr).Isr.ImageId

if id1 == id2 {
t.Errorf("sequential ISR observations should have different image IDs: %q", id1)
}
}
