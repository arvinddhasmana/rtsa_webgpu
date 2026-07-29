// CLASSIFICATION: UNCLASSIFIED
// tests/integration/serializer_roundtrip_test.go — Go serializer round-trip test
//
// Validates that the Go FlatBuffer serializer produces 128-byte binary records
// that exactly match the documented byte-offset layout consumed by the Rust Wasm
// decoder and the WGSL TrackRecord struct.
//
// This test does NOT require Docker or external services. It runs without the
// `integration` build tag so it can be executed with:
//
//go test ./tests/integration/... -run TestSerializerRoundtrip -race
//
// Reference: pkg/flatbuf/layout.go, web-cop-gpu/wasm-decoder/src/lib.rs

package integration

import (
	"encoding/binary"
	"math"
	"os"
	"testing"
	"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/flatbuf"
"google.golang.org/protobuf/types/known/timestamppb"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

func readF32RT(rec flatbuf.Record, off int) float32 {
bits := binary.LittleEndian.Uint32(rec[off : off+4])
return math.Float32frombits(bits)
}

func readU32RT(rec flatbuf.Record, off int) uint32 {
return binary.LittleEndian.Uint32(rec[off : off+4])
}

// degreesToRadiansF32 converts geographic degrees to float32 radians.
func degreesToRadiansF32(deg float64) float32 {
return float32(deg * math.Pi / 180.0)
}

// makeRoundtripUpdate constructs a deterministic TrackUpdate for round-trip testing.
// All field values are chosen to be uniquely identifiable in the binary output.
func makeRoundtripUpdate() *entityv1.TrackUpdate {
altM := 8500.0
// Fixed epoch: 2024-01-15T12:00:00Z = 1705320000 seconds
fixedTime := time.Unix(1_705_320_000, 0).UTC()
return &entityv1.TrackUpdate{
UpdateType: entityv1.TrackUpdate_UPDATE_TYPE_UPDATED,
Track: &entityv1.FusedTrack{
TrackId:        "ROUNDTRIP-001",
HostileClass:   commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE,
EntityType:     commonv1.EntityType_ENTITY_TYPE_AIR,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
EstimatedPosition: &commonv1.Position{
Longitude:      45.678,
Latitude:       -12.345,
AltitudeMeters: &altM,
},
Velocity: &commonv1.Velocity{
NorthMps: 100.0,
EastMps:  200.0,
},
Sources: []*entityv1.SourceAttribution{
{SensorType: commonv1.SensorType_SENSOR_TYPE_RADAR},
},
UpdatedAt: timestamppb.New(fixedTime),
},
}
}

// ── TestSerializerRoundtrip ───────────────────────────────────────────────────

// TestSerializerRoundtrip serializes a deterministic TrackUpdate and verifies
// that every documented field offset in the 128-byte record contains the
// expected value. This ensures the Go serializer and the Rust decoder (which
// reads the same offsets) are mutually compatible.
func TestSerializerRoundtrip(t *testing.T) {
t.Parallel()

s := flatbuf.NewSerializer()
update := makeRoundtripUpdate()

rec, ok := s.Serialize(update)
if !ok {
t.Fatal("Serialize returned false — track update should produce a valid record")
}

if len(rec) != flatbuf.RecordSize {
t.Fatalf("record size: got %d, want %d", len(rec), flatbuf.RecordSize)
}

pos := update.Track.GetEstimatedPosition()

// ── Position fields (offsets 0x00–0x13) ─────────────────────────────────
wantLon := degreesToRadiansF32(pos.GetLongitude())
wantLat := degreesToRadiansF32(pos.GetLatitude())
gotLon := readF32RT(rec, flatbuf.OffLongitude)
gotLat := readF32RT(rec, flatbuf.OffLatitude)

if math.Abs(float64(gotLon-wantLon)) > 1e-6 {
t.Errorf("longitude: got %.6f rad, want %.6f rad", gotLon, wantLon)
}
if math.Abs(float64(gotLat-wantLat)) > 1e-6 {
t.Errorf("latitude: got %.6f rad, want %.6f rad", gotLat, wantLat)
}

// Verify speed is non-zero (derived from northMps=100, eastMps=200 → ≈223.6 m/s)
gotSpeed := readF32RT(rec, flatbuf.OffSpeed)
wantSpeed := float32(math.Sqrt(100*100 + 200*200))
if math.Abs(float64(gotSpeed-wantSpeed)) > 0.1 {
t.Errorf("speed: got %.2f m/s, want %.2f m/s", gotSpeed, wantSpeed)
}

// Altitude: *altM = 8500.0
gotAlt := readF32RT(rec, flatbuf.OffAltitude)
if math.Abs(float64(gotAlt)-8500.0) > 0.01 {
t.Errorf("altitude: got %.2f m, want 8500.0 m", gotAlt)
}

// ── Identity fields (offsets 0x14–0x23) ─────────────────────────────────
gotHash := readU32RT(rec, flatbuf.OffTrackIDHash)
if gotHash == 0 {
t.Error("track_id_hash must be non-zero for a non-empty track ID")
}

// Classification: UNCLASSIFIED = 1
gotClass := readU32RT(rec, flatbuf.OffClassificationLevel)
wantClass := uint32(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
if gotClass != wantClass {
t.Errorf("classification_level: got %d, want %d", gotClass, wantClass)
}

// Threat level: HOSTILE → 5
gotThreat := readU32RT(rec, flatbuf.OffThreatLevel)
if gotThreat != 5 {
t.Errorf("threat_level: got %d, want 5 (Hostile)", gotThreat)
}

// ── Display fields (offsets 0x24–0x2F) ───────────────────────────────────
// Alert flags: threat_level ≥ 4 → alertFlags = 1
gotAlertFlags := readU32RT(rec, flatbuf.OffAlertFlags)
if gotAlertFlags != 1 {
t.Errorf("alert_flags: got %d, want 1 (hostile track triggers alert)", gotAlertFlags)
}

// Source bitmap: RADAR sensor → bit set at SensorType_SENSOR_TYPE_RADAR value
gotSourceBitmap := readU32RT(rec, flatbuf.OffSourceBitmap)
if gotSourceBitmap == 0 {
t.Error("source_bitmap must be non-zero when a sensor source is present")
}

// Update epoch: non-zero (derived from 2024-01-15T12:00:00Z)
gotEpoch := readU32RT(rec, flatbuf.OffUpdateEpochMs)
if gotEpoch == 0 {
t.Error("update_epoch_ms must be non-zero for a track with a timestamp")
}

// ── Record is exactly 128 bytes ───────────────────────────────────────────
t.Logf("Record is %d bytes (correct)", len(rec))
t.Logf("track_id_hash=0x%08x threat=%d class=%d speed=%.1f alt=%.1f epoch=%d",
gotHash, gotThreat, gotClass, gotSpeed, gotAlt, gotEpoch)

	// ── Write a temp record for optional manual Rust decoder validation ───────
	// Keep the test hermetic: never write artifacts into the repository tree.
	tmpPath := t.TempDir() + "/roundtrip_track.bin"
	if err := os.WriteFile(tmpPath, rec[:], 0o600); err == nil {
		t.Logf("roundtrip binary fixture written to %s", tmpPath)
	}
}

// TestSerializerRoundtrip_NilInput ensures the serializer handles nil gracefully.
func TestSerializerRoundtrip_NilInput(t *testing.T) {
t.Parallel()
s := flatbuf.NewSerializer()

_, ok := s.Serialize(nil)
if ok {
t.Error("Serialize(nil) must return false")
}

_, ok = s.Serialize(&entityv1.TrackUpdate{Track: nil})
if ok {
t.Error("Serialize with nil Track must return false")
}
}

// TestSerializerRoundtrip_SerializeBatch ensures SerializeBatch produces the
// correct byte length and caps at MaxBatchSize records.
func TestSerializerRoundtrip_SerializeBatch(t *testing.T) {
t.Parallel()

s := flatbuf.NewSerializer()
update := makeRoundtripUpdate()

// 3 identical updates → expect 3 × 128 bytes
batch := s.SerializeBatch([]*entityv1.TrackUpdate{update, update, update})
if len(batch) != 3*flatbuf.RecordSize {
t.Errorf("batch length: got %d, want %d", len(batch), 3*flatbuf.RecordSize)
}

// MaxBatchSize (9) updates → 11 input is capped to exactly 9 × 128 bytes
updates := make([]*entityv1.TrackUpdate, 11)
for i := range updates {
updates[i] = update
}
bigBatch := s.SerializeBatch(updates)
wantLen := flatbuf.MaxBatchSize * flatbuf.RecordSize
if len(bigBatch) != wantLen {
t.Errorf("batch capping: got %d bytes, want %d bytes (MaxBatchSize=%d)",
len(bigBatch), wantLen, flatbuf.MaxBatchSize)
}
}
