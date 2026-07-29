// CLASSIFICATION: UNCLASSIFIED
package sensor_test

import (
"math/rand"
"regexp"
"testing"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/generator"
"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/sensor"
)

func init() {
sensor.SetRNG(rand.New(rand.NewSource(42)))
}

func testSurfaceEntityForSensor() *generator.SimEntity {
return &generator.SimEntity{
ID:         "SURF-0001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
HostileClass: commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL,
Position: generator.Position{
Lat:     45.0,
Lon:     -60.0,
AltM:    0,
SpeedKn: 12.0,
Heading: 90.0,
},
}
}

func testAirEntityForSensor() *generator.SimEntity {
return &generator.SimEntity{
ID:         "AIR-0001",
EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
Position: generator.Position{
Lat:     45.5,
Lon:     -61.0,
AltM:    5000,
SpeedKn: 300.0,
Heading: 270.0,
},
}
}

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestGenerateRadarObservation_AllFieldsPopulated(t *testing.T) {
sensor.SetRNG(rand.New(rand.NewSource(42)))
e := testSurfaceEntityForSensor()
obs := sensor.GenerateRadarObservation(e, "RADAR-SIM-001")

if obs == nil {
t.Fatal("observation should not be nil")
}
if obs.SensorId != "RADAR-SIM-001" {
t.Errorf("expected sensor_id RADAR-SIM-001, got %q", obs.SensorId)
}
if obs.SensorType != commonv1.SensorType_SENSOR_TYPE_RADAR {
t.Errorf("expected SENSOR_TYPE_RADAR, got %v", obs.SensorType)
}
if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("classification must be UNCLASSIFIED, got %v", obs.Classification)
}
if obs.ObservationTime == nil {
t.Error("observation_time must be set")
}
if obs.Position == nil {
t.Error("position must be set")
}
if !uuidRe.MatchString(obs.ObservationId) {
t.Errorf("observation_id should be a UUID, got %q", obs.ObservationId)
}

radarPayload, ok := obs.SensorData.(*ingestionv1.SensorObservation_Radar)
if !ok {
t.Fatal("sensor_data should be RadarTrack")
}
rt := radarPayload.Radar
if rt.TrackNumber == "" {
t.Error("track_number must be non-empty")
}
if rt.TrackQuality < 0.6 || rt.TrackQuality > 1.0 {
t.Errorf("track_quality must be in [0.6,1.0], got %f", rt.TrackQuality)
}
if rt.RangeNm < 0 {
t.Errorf("range_nm must be non-negative, got %f", rt.RangeNm)
}
if rt.BearingDegrees < 0 || rt.BearingDegrees > 360 {
t.Errorf("bearing_degrees must be in [0,360], got %f", rt.BearingDegrees)
}
}

func TestGenerateRadarObservation_Position(t *testing.T) {
sensor.SetRNG(rand.New(rand.NewSource(42)))
e := testSurfaceEntityForSensor()
obs := sensor.GenerateRadarObservation(e, "RADAR-SIM-001")

pos := obs.Position
if pos.Latitude != e.Position.Lat {
t.Errorf("latitude mismatch: want %f, got %f", e.Position.Lat, pos.Latitude)
}
if pos.Longitude != e.Position.Lon {
t.Errorf("longitude mismatch: want %f, got %f", e.Position.Lon, pos.Longitude)
}
}

func TestGenerateRadarObservation_Classification(t *testing.T) {
sensor.SetRNG(rand.New(rand.NewSource(42)))
e := testAirEntityForSensor()
obs := sensor.GenerateRadarObservation(e, "RADAR-SIM-002")

if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("all simulator observations must be UNCLASSIFIED, got %v", obs.Classification)
}
}

func TestGenerateRadarObservation_SequentialTrackNumbers(t *testing.T) {
sensor.SetRNG(rand.New(rand.NewSource(42)))
e := testSurfaceEntityForSensor()

obs1 := sensor.GenerateRadarObservation(e, "RADAR-SIM-UNIQUE-001")
obs2 := sensor.GenerateRadarObservation(e, "RADAR-SIM-UNIQUE-001")

rt1 := obs1.SensorData.(*ingestionv1.SensorObservation_Radar).Radar
rt2 := obs2.SensorData.(*ingestionv1.SensorObservation_Radar).Radar

if rt1.TrackNumber == rt2.TrackNumber {
t.Errorf("sequential observations should have different track numbers: %q vs %q",
rt1.TrackNumber, rt2.TrackNumber)
}
}
