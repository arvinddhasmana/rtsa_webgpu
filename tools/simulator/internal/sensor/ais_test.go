// CLASSIFICATION: UNCLASSIFIED
package sensor_test

import (
"regexp"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/generator"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/sensor"
)

var mmsiRe = regexp.MustCompile(`^[0-9]{9}$`)
var validAISMessageTypes = map[int32]bool{1: true, 2: true, 3: true, 5: true, 18: true, 24: true}

func TestGenerateAISObservation_ValidFields(t *testing.T) {
e := &generator.SimEntity{
ID:         "SURF-AIS-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Position: generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 10.0,
Heading: 180.0,
},
}

obs := sensor.GenerateAISObservation(e, nil)

if obs == nil {
t.Fatal("AIS observation should not be nil")
}
if obs.SensorType != commonv1.SensorType_SENSOR_TYPE_AIS_BFT {
t.Errorf("expected SENSOR_TYPE_AIS_BFT, got %v", obs.SensorType)
}
if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("must be UNCLASSIFIED, got %v", obs.Classification)
}

aisPl, ok := obs.SensorData.(*ingestionv1.SensorObservation_AisBft)
if !ok {
t.Fatal("sensor_data should be AISPosition")
}
ais := aisPl.AisBft

// MMSI must be exactly 9 digits.
if !mmsiRe.MatchString(ais.Mmsi) {
t.Errorf("MMSI must be 9 digits, got %q", ais.Mmsi)
}

// Vessel name must be non-empty.
if ais.VesselName == "" {
t.Error("vessel_name must be non-empty")
}

// Vessel type code must be 1-99.
if ais.VesselTypeCode < 1 || ais.VesselTypeCode > 99 {
t.Errorf("vessel_type_code must be 1-99, got %d", ais.VesselTypeCode)
}

// AIS message type must be in {1,2,3,5,18,24}.
if !validAISMessageTypes[ais.AisMessageType] {
t.Errorf("invalid ais_message_type %d", ais.AisMessageType)
}

// BFT must be false for UNCLASSIFIED.
if ais.IsBft {
t.Error("is_bft must be false for UNCLASSIFIED observations")
}

// Nav status must be non-empty.
if ais.NavStatus == "" {
t.Error("nav_status must be non-empty")
}
}

func TestGenerateAISObservation_ManipulatedPosition(t *testing.T) {
e := &generator.SimEntity{
ID:         "SURF-AIS-SPOOF",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Position: generator.Position{
Lat:     45.0,
Lon:     -60.0,
SpeedKn: 10.0,
},
}
manipulatedPos := &generator.Position{
Lat:     45.02,
Lon:     -59.98,
SpeedKn: 10.0,
}

obs := sensor.GenerateAISObservation(e, manipulatedPos)

if obs.Position.Latitude != manipulatedPos.Lat {
t.Errorf("expected manipulated lat %f, got %f", manipulatedPos.Lat, obs.Position.Latitude)
}
if obs.Position.Longitude != manipulatedPos.Lon {
t.Errorf("expected manipulated lon %f, got %f", manipulatedPos.Lon, obs.Position.Longitude)
}
}

func TestGenerateAISObservation_StableMMSI(t *testing.T) {
e := &generator.SimEntity{
ID:         "SURF-AIS-STABLE",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Position:   generator.Position{Lat: 45.0, Lon: -60.0, SpeedKn: 5.0},
}

obs1 := sensor.GenerateAISObservation(e, nil)
obs2 := sensor.GenerateAISObservation(e, nil)

mmsi1 := obs1.SensorData.(*ingestionv1.SensorObservation_AisBft).AisBft.Mmsi
mmsi2 := obs2.SensorData.(*ingestionv1.SensorObservation_AisBft).AisBft.Mmsi

if mmsi1 != mmsi2 {
t.Errorf("same entity should always get same MMSI: %q vs %q", mmsi1, mmsi2)
}
}

func TestGenerateAISObservation_PositionSet(t *testing.T) {
e := &generator.SimEntity{
ID:         "SURF-AIS-POS",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Position: generator.Position{
Lat:     46.0,
Lon:     -58.0,
SpeedKn: 8.0,
Heading: 270.0,
},
}
obs := sensor.GenerateAISObservation(e, nil)

if obs.Position == nil {
t.Fatal("AIS observation must have position")
}
if obs.Position.Latitude != e.Position.Lat {
t.Errorf("lat mismatch: want %f, got %f", e.Position.Lat, obs.Position.Latitude)
}
}
