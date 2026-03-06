// CLASSIFICATION: UNCLASSIFIED
// pkg/flatbuf/serializer.go — Protobuf FusedTrack → 128-byte binary record serializer
//
// Converts a Protobuf TrackUpdate (from the track-fused Redpanda topic) into the
// fixed 128-byte GPU-aligned binary record consumed by the browser Wasm decoder and
// written directly into the SharedArrayBuffer.
//
// Design: The record format is a raw binary encoding at documented byte offsets —
// not a FlatBuffer table with vtable overhead — ensuring exact GPU alignment.
// The .fbs schema in proto/rtsa/flatbuf/v1/track_update.fbs documents the layout.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/flatbuffers_guidelines.md §3

package flatbuf

import (
"encoding/binary"
"math"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
)

const (
// fnvPrime and fnvOffset are FNV-1a 32-bit constants.
fnvOffset uint32 = 2166136261
fnvPrime  uint32 = 16777619

// degreesToRadians converts geographic degrees to radians.
degreesToRadians = math.Pi / 180.0
)

// Record is a fixed 128-byte GPU-aligned binary record.
type Record [RecordSize]byte

// Serializer converts Protobuf TrackUpdate messages to 128-byte binary records.
// A single Serializer instance is safe for sequential use. Reuse it to avoid
// per-record allocations.
type Serializer struct {
// trailHistory stores the last NumTrailSlots positions per track to build
// the trail ring buffer. Keyed by track_id_hash.
trailHistory map[uint32][NumTrailSlots][2]float32
}

// NewSerializer creates a new Serializer.
func NewSerializer() *Serializer {
return &Serializer{
trailHistory: make(map[uint32][NumTrailSlots][2]float32),
}
}

// Serialize converts a TrackUpdate proto message to a 128-byte binary record.
// Returns the record and true on success; returns an empty record and false if
// the input is nil or has no position data.
func (s *Serializer) Serialize(update *entityv1.TrackUpdate) (Record, bool) {
if update == nil || update.Track == nil {
return Record{}, false
}
track := update.Track

pos := track.GetEstimatedPosition()
if pos == nil {
return Record{}, false
}

var rec Record

// ── Position fields ──────────────────────────────────────────────────
lonRad := float32(pos.GetLongitude() * degreesToRadians)
latRad := float32(pos.GetLatitude() * degreesToRadians)
courseRad := float32(0)
speedMS := float32(0)

// Derive course and speed from velocity if present
if vel := track.GetVelocity(); vel != nil {
north := vel.GetNorthMps()
east := vel.GetEastMps()
speedMS = float32(math.Sqrt(north*north + east*east))
if speedMS > 0 {
courseRad = float32(math.Atan2(east, north))
if courseRad < 0 {
courseRad += float32(2 * math.Pi)
}
}
}

altM := float32(0)
if a := pos.AltitudeMeters; a != nil {
altM = float32(*a)
}

putF32(&rec, OffLongitude, lonRad)
putF32(&rec, OffLatitude, latRad)
putF32(&rec, OffCourse, courseRad)
putF32(&rec, OffSpeed, speedMS)
putF32(&rec, OffAltitude, altM)

// ── Identity fields ──────────────────────────────────────────────────
idHash := fnv1a(track.GetTrackId())
sourceBitmap := buildSourceBitmap(track.GetSources())
classLevel := uint32(track.GetClassification())
threatLevel := hostileClassToThreatLevel(track.GetHostileClass())

putU32(&rec, OffTrackIDHash, idHash)
putU32(&rec, OffSourceBitmap, sourceBitmap)
putU32(&rec, OffClassificationLevel, classLevel)
putU32(&rec, OffThreatLevel, threatLevel)

// ── Display fields ───────────────────────────────────────────────────
iconIndex := buildIconIndex(track.GetEntityType(), track.GetHostileClass())
alertFlags := uint32(0) // Phase 3 will wire alert state; keep zero for now

var epochMs uint32
if ts := track.GetUpdatedAt(); ts != nil {
t := ts.AsTime()
epochMs = uint32(t.UnixMilli() & 0xFFFF_FFFF)
} else {
epochMs = uint32(time.Now().UnixMilli() & 0xFFFF_FFFF)
}

putU32(&rec, OffIconIndex, iconIndex)
putU32(&rec, OffAlertFlags, alertFlags)
putU32(&rec, OffUpdateEpochMs, epochMs)

// ── Trail ring buffer ────────────────────────────────────────────────
// Shift the history: push current position, drop oldest.
s.updateTrailHistory(idHash, lonRad, latRad)
s.writeTrail(&rec, idHash, lonRad, latRad)

return rec, true
}

// SerializeBatch converts up to MaxBatchSize TrackUpdate messages into a
// concatenated byte slice for transmission as a single WebTransport datagram.
// The caller must ensure len(updates) <= MaxBatchSize.
func (s *Serializer) SerializeBatch(updates []*entityv1.TrackUpdate) []byte {
if len(updates) == 0 {
return nil
}
n := len(updates)
if n > MaxBatchSize {
n = MaxBatchSize
}
out := make([]byte, 0, n*RecordSize)
for i := 0; i < n; i++ {
rec, ok := s.Serialize(updates[i])
if !ok {
continue
}
out = append(out, rec[:]...)
}
return out
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// fnv1a computes the FNV-1a 32-bit hash of a string.
func fnv1a(s string) uint32 {
h := fnvOffset
for i := 0; i < len(s); i++ {
h ^= uint32(s[i])
h *= fnvPrime
}
return h
}

// buildSourceBitmap sets one bit per contributing sensor type.
// Bit positions match the SensorType enum values (0–7).
func buildSourceBitmap(sources []*entityv1.SourceAttribution) uint32 {
var bitmap uint32
for _, src := range sources {
st := src.GetSensorType()
if st > 0 && int(st) < 32 {
bitmap |= 1 << uint(st)
}
}
return bitmap
}

// hostileClassToThreatLevel maps HostileClassification to a ThreatLevel uint32
// matching the ThreatLevel enum in the .fbs schema.
func hostileClassToThreatLevel(hc commonv1.HostileClassification) uint32 {
switch hc {
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE:
return 5 // Hostile
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_SUSPECT:
return 4 // Suspect
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL:
return 3 // Neutral
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY:
return 2 // Friendly
case commonv1.HostileClassification_HOSTILE_CLASSIFICATION_PENDING:
return 1 // Pending
default:
return 0 // Unknown
}
}

// buildIconIndex maps entity type and hostile classification to a GPU atlas index.
// Icon grid: entity_type (0–4) × 6 + threat_level (0–5) = 0–29.
func buildIconIndex(et commonv1.EntityType, hc commonv1.HostileClassification) uint32 {
entityBase := uint32(et) // 0=Unspecified, 1=Surface, 2=Air, 3=Subsurface, 4=Land, 5=Cyber
if entityBase > 5 {
entityBase = 0
}
return entityBase*6 + hostileClassToThreatLevel(hc)
}

// updateTrailHistory pushes the current position into the trail ring buffer
// for the given track, shifting older positions back.
func (s *Serializer) updateTrailHistory(idHash uint32, lon, lat float32) {
history := s.trailHistory[idHash]
// Shift: [1]→[0], [2]→[1], [3]→[2], [4]→[3], current→[4]
copy(history[0:], history[1:])
history[NumTrailSlots-1] = [2]float32{lon, lat}
s.trailHistory[idHash] = history
}

// writeTrail encodes the trail ring buffer into the record.
// Each trail slot holds one segment: (lon_a, lat_a, lon_b, lat_b).
// Trail 0 is the most recent; trail 4 is the oldest.
func (s *Serializer) writeTrail(rec *Record, idHash uint32, curLon, curLat float32) {
history := s.trailHistory[idHash]
offsets := [NumTrailSlots]int{OffTrail0, OffTrail1, OffTrail2, OffTrail3, OffTrail4}

// Trail 0: most recent segment (history[NumTrailSlots-2] → current)
// Trail n: nth-oldest segment
for i := 0; i < NumTrailSlots; i++ {
off := offsets[i]
slotIdx := NumTrailSlots - 1 - i // trail_0 = most recent = history index 4→current
var lonA, latA, lonB, latB float32

if slotIdx == NumTrailSlots-1 {
// Most recent segment: previous position → current
prev := NumTrailSlots - 2
if prev >= 0 {
lonA = history[prev][0]
latA = history[prev][1]
}
lonB = curLon
latB = curLat
} else {
// Older segments: consecutive history entries
next := slotIdx + 1
lonA = history[slotIdx][0]
latA = history[slotIdx][1]
lonB = history[next][0]
latB = history[next][1]
}

putF32(rec, off+0, lonA)
putF32(rec, off+4, latA)
putF32(rec, off+8, lonB)
putF32(rec, off+12, latB)
}
}

// putF32 writes a float32 in little-endian byte order at offset off in rec.
func putF32(rec *Record, off int, v float32) {
binary.LittleEndian.PutUint32(rec[off:off+4], math.Float32bits(v))
}

// putU32 writes a uint32 in little-endian byte order at offset off in rec.
func putU32(rec *Record, off int, v uint32) {
binary.LittleEndian.PutUint32(rec[off:off+4], v)
}
