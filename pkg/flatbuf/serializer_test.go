// CLASSIFICATION: UNCLASSIFIED
// pkg/flatbuf/serializer_test.go — Unit tests for the FlatBuffer serializer
//
// Tests: round-trip field values, FNV-1a hash, trail ring buffer,
//        classification level passthrough, nil-safety.

package flatbuf_test

import (
"encoding/binary"
"math"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/flatbuf"
"google.golang.org/protobuf/types/known/timestamppb"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func readF32(rec flatbuf.Record, off int) float32 {
bits := binary.LittleEndian.Uint32(rec[off : off+4])
return math.Float32frombits(bits)
}

func readU32(rec flatbuf.Record, off int) uint32 {
return binary.LittleEndian.Uint32(rec[off : off+4])
}

func makeTrackUpdate(
trackID string,
lonDeg, latDeg float64,
altM float64,
northMps, eastMps float64,
hostile commonv1.HostileClassification,
entity commonv1.EntityType,
classLevel commonv1.ClassificationLevel,
sensors []commonv1.SensorType,
) *entityv1.TrackUpdate {
sources := make([]*entityv1.SourceAttribution, len(sensors))
for i, st := range sensors {
sources[i] = &entityv1.SourceAttribution{SensorType: st}
}
return &entityv1.TrackUpdate{
UpdateType: entityv1.TrackUpdate_UPDATE_TYPE_UPDATED,
Track: &entityv1.FusedTrack{
TrackId:      trackID,
HostileClass: hostile,
EntityType:   entity,
Classification: classLevel,
EstimatedPosition: &commonv1.Position{
Longitude:      lonDeg,
Latitude:       latDeg,
AltitudeMeters: &altM,
},
Velocity: &commonv1.Velocity{
NorthMps: northMps,
EastMps:  eastMps,
},
Sources:   sources,
UpdatedAt: timestamppb.Now(),
},
}
}

// ── test cases ───────────────────────────────────────────────────────────────

func TestSerialize_RecordSize(t *testing.T) {
s := flatbuf.NewSerializer()
update := makeTrackUpdate(
"track-001",
45.0, 30.0, 1000.0,
100.0, 50.0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false for valid input")
}
if len(rec) != flatbuf.RecordSize {
t.Errorf("expected record size %d, got %d", flatbuf.RecordSize, len(rec))
}
}

func TestSerialize_PositionFields(t *testing.T) {
const lonDeg = 45.0
const latDeg = -30.0
const altM = 1500.0

s := flatbuf.NewSerializer()
update := makeTrackUpdate(
"pos-test",
lonDeg, latDeg, altM,
0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false")
}

const deg2rad = math.Pi / 180.0
wantLon := float32(lonDeg * deg2rad)
wantLat := float32(latDeg * deg2rad)
wantAlt := float32(altM)

if got := readF32(rec, flatbuf.OffLongitude); got != wantLon {
t.Errorf("longitude: want %v, got %v", wantLon, got)
}
if got := readF32(rec, flatbuf.OffLatitude); got != wantLat {
t.Errorf("latitude: want %v, got %v", wantLat, got)
}
if got := readF32(rec, flatbuf.OffAltitude); got != wantAlt {
t.Errorf("altitude: want %v, got %v", wantAlt, got)
}
}

func TestSerialize_CourseAndSpeed(t *testing.T) {
// Move due East: north=0, east=100 m/s → speed=100, course=PI/2
s := flatbuf.NewSerializer()
update := makeTrackUpdate(
"speed-test",
0, 0, 0,
0, 100, // east-only velocity
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false")
}

gotSpeed := readF32(rec, flatbuf.OffSpeed)
if math.Abs(float64(gotSpeed)-100.0) > 0.001 {
t.Errorf("speed: want 100.0, got %v", gotSpeed)
}
gotCourse := readF32(rec, flatbuf.OffCourse)
wantCourse := float32(math.Pi / 2)
if math.Abs(float64(gotCourse-wantCourse)) > 0.001 {
t.Errorf("course: want %.4f (PI/2), got %.4f", wantCourse, gotCourse)
}
}

func TestSerialize_TrackIDHash(t *testing.T) {
s := flatbuf.NewSerializer()
update := makeTrackUpdate(
"hash-test-id",
0, 0, 0, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false")
}

hash := readU32(rec, flatbuf.OffTrackIDHash)
if hash == 0 {
t.Error("track_id_hash must not be 0 for non-empty track ID")
}

// Same ID → same hash
rec2, _ := s.Serialize(update)
if readU32(rec2, flatbuf.OffTrackIDHash) != hash {
t.Error("same track ID must produce same hash")
}

// Different ID → different hash (with overwhelming probability)
update2 := makeTrackUpdate("different-id", 0, 0, 0, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec3, _ := s.Serialize(update2)
if readU32(rec3, flatbuf.OffTrackIDHash) == hash {
t.Error("different track IDs produced the same hash")
}
}

func TestSerialize_ClassificationLevel(t *testing.T) {
tests := []struct {
level commonv1.ClassificationLevel
want  uint32
}{
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 1},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A, 2},
{commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 5},
}

s := flatbuf.NewSerializer()
for _, tc := range tests {
update := makeTrackUpdate("cls-test", 0, 0, 0, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_AIR,
tc.level, nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatalf("Serialize returned false for level %v", tc.level)
}
if got := readU32(rec, flatbuf.OffClassificationLevel); got != tc.want {
t.Errorf("classification_level %v: want %d, got %d", tc.level, tc.want, got)
}
}
}

func TestSerialize_ThreatLevel(t *testing.T) {
tests := []struct {
hc   commonv1.HostileClassification
want uint32
}{
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, 5},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_SUSPECT, 4},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL, 3},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY, 2},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_PENDING, 1},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, 0},
}

s := flatbuf.NewSerializer()
for _, tc := range tests {
update := makeTrackUpdate("threat-test", 0, 0, 0, 0, 0,
tc.hc, commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatalf("Serialize returned false for hc %v", tc.hc)
}
if got := readU32(rec, flatbuf.OffThreatLevel); got != tc.want {
t.Errorf("threat_level hc=%v: want %d, got %d", tc.hc, tc.want, got)
}
}
}

func TestSerialize_SourceBitmap(t *testing.T) {
s := flatbuf.NewSerializer()
sensors := []commonv1.SensorType{
commonv1.SensorType_SENSOR_TYPE_RADAR,    // bit 1
commonv1.SensorType_SENSOR_TYPE_AIS_BFT,  // bit 5
}
update := makeTrackUpdate("src-test", 0, 0, 0, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_SURFACE,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
sensors,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false")
}

bitmap := readU32(rec, flatbuf.OffSourceBitmap)
wantBit1 := uint32(1 << 1) // RADAR
wantBit5 := uint32(1 << 5) // AIS_BFT
if bitmap&wantBit1 == 0 {
t.Error("expected RADAR bit (1) set in source_bitmap")
}
if bitmap&wantBit5 == 0 {
t.Error("expected AIS_BFT bit (5) set in source_bitmap")
}
}

func TestSerialize_IconIndex(t *testing.T) {
// Air Hostile → entityBase=2, threat=5 → icon=2*6+5=17
s := flatbuf.NewSerializer()
update := makeTrackUpdate("icon-test", 0, 0, 0, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false")
}
want := uint32(2*6 + 5) // ENTITY_TYPE_AIR=2, Hostile=5
if got := readU32(rec, flatbuf.OffIconIndex); got != want {
t.Errorf("icon_index: want %d, got %d", want, got)
}
}

func TestSerialize_NilUpdate(t *testing.T) {
s := flatbuf.NewSerializer()
_, ok := s.Serialize(nil)
if ok {
t.Error("Serialize(nil) must return false")
}
}

func TestSerialize_NilTrack(t *testing.T) {
s := flatbuf.NewSerializer()
_, ok := s.Serialize(&entityv1.TrackUpdate{})
if ok {
t.Error("Serialize(empty TrackUpdate) must return false")
}
}

func TestSerializeBatch_MaxBatch(t *testing.T) {
s := flatbuf.NewSerializer()
updates := make([]*entityv1.TrackUpdate, flatbuf.MaxBatchSize+2)
for i := range updates {
updates[i] = makeTrackUpdate(
"batch-track",
float64(i), float64(i), float64(i*100),
0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
}
out := s.SerializeBatch(updates)
// Must not exceed MaxBatchSize records
maxBytes := flatbuf.MaxBatchSize * flatbuf.RecordSize
if len(out) > maxBytes {
t.Errorf("SerializeBatch exceeded max size: got %d bytes, want <= %d", len(out), maxBytes)
}
}

func TestSerializeBatch_Empty(t *testing.T) {
s := flatbuf.NewSerializer()
out := s.SerializeBatch(nil)
if out != nil {
t.Error("SerializeBatch(nil) should return nil")
}
}

func TestSerialize_UpdateEpochMs(t *testing.T) {
s := flatbuf.NewSerializer()
update := makeTrackUpdate("ts-test", 0, 0, 0, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false")
}
epochMs := readU32(rec, flatbuf.OffUpdateEpochMs)
if epochMs == 0 {
t.Error("update_epoch_ms must not be zero")
}
}

func TestSerialize_TrailHistory(t *testing.T) {
// After two updates with different positions, trail_0 should contain
// (prev_lon, prev_lat, cur_lon, cur_lat).
s := flatbuf.NewSerializer()

pos1 := makeTrackUpdate("trail-test", 10.0, 20.0, 500, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)
pos2 := makeTrackUpdate("trail-test", 11.0, 21.0, 500, 0, 0,
commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY,
commonv1.EntityType_ENTITY_TYPE_AIR,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
nil,
)

_, ok := s.Serialize(pos1)
if !ok {
t.Fatal("first Serialize returned false")
}
rec2, ok := s.Serialize(pos2)
if !ok {
t.Fatal("second Serialize returned false")
}

// trail_0 lon_b and lat_b should match current position
const deg2rad = math.Pi / 180.0
wantLon := float32(11.0 * deg2rad)
wantLat := float32(21.0 * deg2rad)

lonB := readF32(rec2, flatbuf.OffTrail0+8)
latB := readF32(rec2, flatbuf.OffTrail0+12)

if math.Abs(float64(lonB-wantLon)) > 1e-5 {
t.Errorf("trail_0 lon_b: want %v, got %v", wantLon, lonB)
}
if math.Abs(float64(latB-wantLat)) > 1e-5 {
t.Errorf("trail_0 lat_b: want %v, got %v", wantLat, latB)
}
}
