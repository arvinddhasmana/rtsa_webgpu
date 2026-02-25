// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"fmt"
	"math"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/state"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHaversineNM_SamePoint(t *testing.T) {
dist := HaversineNM(44.6476, -63.5728, 44.6476, -63.5728)
if dist > 0.001 {
t.Errorf("same point distance = %v NM, want ~0", dist)
}
}

func TestHaversineNM_KnownDistance(t *testing.T) {
// 1 degree of latitude ≈ 60 NM.
dist := HaversineNM(44.0, -63.0, 45.0, -63.0)
if dist < 58 || dist > 62 {
t.Errorf("1° latitude distance = %v NM, want ~60", dist)
}
}

func TestAngularDifference_Wraparound(t *testing.T) {
tests := []struct {
name    string
a, b    float64
want    float64
}{
{"wraparound_pos", 350, 10, -20},
{"wraparound_neg", 10, 350, 20},
{"simple", 90, 45, 45},
{"opposite_neg", 0, 180, -180},
}
for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := angularDifference(tt.a, tt.b)
if math.Abs(got-tt.want) > 0.001 {
t.Errorf("angularDifference(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
}
})
}
}

func TestComputePatternConfidence_Loitering(t *testing.T) {
fe := &FeatureExtractor{}
entries := buildLoiteringEntries(45, 44.65, -63.57)
pos := &Position{Latitude: 44.65, Longitude: -63.57}

conf := fe.computePatternConfidence(entries, pos)
if conf < 0.75 {
t.Errorf("loitering confidence = %v, want >= 0.75", conf)
}
}

func TestComputePatternConfidence_NormalTransit(t *testing.T) {
fe := &FeatureExtractor{}
entries := buildTransitEntries()
pos := &Position{Latitude: 45.0, Longitude: -64.0}

conf := fe.computePatternConfidence(entries, pos)
if conf >= 0.75 {
t.Errorf("normal transit confidence = %v, want < 0.75", conf)
}
}

func TestComputeExpectedHeading_Empty(t *testing.T) {
fe := &FeatureExtractor{}
got := fe.computeExpectedHeading(nil, 90.0)
if got != 90.0 {
t.Errorf("expected heading with no history = %v, want 90.0", got)
}
}

// ── Test Helpers ──────────────────────────────────────────────────────────────

// buildLoiteringEntries creates entries at nearly the same position over durationMin minutes.
func buildLoiteringEntries(durationMin int, lat, lon float64) []*state.HistoryEntry {
entries := make([]*state.HistoryEntry, durationMin)
for i := 0; i < durationMin; i++ {
entries[i] = &state.HistoryEntry{
Timestamp:  time.Now().Add(-time.Duration(durationMin-i) * time.Minute),
Latitude:   lat,
Longitude:  lon,
SpeedKnots: 0.5, // near stationary
Heading:    90,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
}
}
return entries
}

// buildTransitEntries creates entries spread over a large distance (normal transit).
func buildTransitEntries() []*state.HistoryEntry {
entries := make([]*state.HistoryEntry, 10)
for i := 0; i < 10; i++ {
entries[i] = &state.HistoryEntry{
Timestamp:  time.Now().Add(-time.Duration(10-i) * time.Minute),
Latitude:   44.0 + float64(i)*0.1, // Moving ~6NM/update
Longitude:  -63.0 - float64(i)*0.1,
SpeedKnots: 15.0,
Heading:    45,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
}
}
return entries
}


// buildTestFusedTrack builds a minimal FusedTrack for testing.
func buildTestFusedTrack(trackID string, lat, lon, speedKnots, heading float64) *entityv1.FusedTrack {
	_ = math.Pi // ensure math is used
	_ = fmt.Sprintf // ensure fmt is used
	return &entityv1.FusedTrack{
		TrackId:    trackID,
		EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
		Velocity: &commonv1.Velocity{
			NorthMps: speedKnots / 1.94384 * 0, // simplified; east only
			EastMps:  speedKnots / 1.94384,
		},
		EstimatedPosition: &commonv1.Position{
			Latitude:  lat,
			Longitude: lon,
		},
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		AgeSeconds: 300.0,
	}
}

func TestFeatureExtractor_Extract_BasicTrack(t *testing.T) {
history := state.NewTrackHistory(100, 2*time.Hour)
extractor := NewFeatureExtractor(history, nil)

track := buildTestFusedTrack("extract-track-01", 44.65, -63.57, 10.0, 90.0)

fv, err := extractor.Extract(track)
if err != nil {
t.Fatalf("Extract returned error: %v", err)
}
if fv.TrackID != "extract-track-01" {
t.Errorf("TrackID = %q, want extract-track-01", fv.TrackID)
}
}

func TestFeatureExtractor_Extract_NilTrack(t *testing.T) {
history := state.NewTrackHistory(100, 2*time.Hour)
extractor := NewFeatureExtractor(history, nil)

_, err := extractor.Extract(nil)
if err == nil {
t.Error("Expected error for nil track")
}
}

func TestFeatureExtractor_Extract_WithExclusionZone_InZone(t *testing.T) {
history := state.NewTrackHistory(100, 2*time.Hour)
zones := []ExclusionZone{
{
Name:      "Test Zone",
CenterLat: 44.65,
CenterLon: -63.57,
RadiusNM:  5.0,
},
}
extractor := NewFeatureExtractor(history, zones)

// Track right at the zone center.
track := buildTestFusedTrack("zone-track", 44.65, -63.57, 5.0, 0.0)

fv, err := extractor.Extract(track)
if err != nil {
t.Fatalf("Extract returned error: %v", err)
}
if !fv.InExclusionZone {
t.Error("Track at zone center should be inside exclusion zone")
}
}

func TestFeatureExtractor_Extract_WithSpeedHistory(t *testing.T) {
history := state.NewTrackHistory(100, 2*time.Hour)

// Pre-seed with history.
for i := 0; i < 5; i++ {
history.Append("speed-hist-track", &state.HistoryEntry{
Timestamp:  time.Now().Add(-time.Duration(5-i) * time.Minute),
SpeedKnots: 10.0 + float64(i),
Heading:    90.0,
Latitude:   44.65,
Longitude:  -63.57,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
})
}

extractor := NewFeatureExtractor(history, nil)
track := buildTestFusedTrack("speed-hist-track", 44.65, -63.57, 12.0, 90.0)

fv, err := extractor.Extract(track)
if err != nil {
t.Fatalf("Extract returned error: %v", err)
}
if fv.AvgSpeed30Min <= 0 {
t.Errorf("AvgSpeed30Min should be > 0 after seeding history, got %v", fv.AvgSpeed30Min)
}
}

func TestHaversineNM_ExportedAlias(t *testing.T) {
// Test via exported alias.
d := HaversineNM(44.65, -63.57, 44.65, -63.57)
if d > 0.001 {
t.Errorf("Same point distance = %v, want ~0", d)
}
}

func TestComputeTemporalPValue(t *testing.T) {
fe := &FeatureExtractor{}
tests := []struct {
hour   int
maxP   float64
}{
{2, 0.02},  // Early morning: very low p-value
{5, 0.04},  // Pre-dawn: low p-value
{14, 0.20}, // Afternoon: higher p-value
}
for _, tt := range tests {
t.Run(fmt.Sprintf("hour=%d", tt.hour), func(t *testing.T) {
testTime := time.Date(2026, 1, 1, tt.hour, 0, 0, 0, time.UTC)
p := fe.computeTemporalPValue("", testTime)
if p > tt.maxP {
t.Errorf("Hour %d: p-value %v exceeds expected max %v", tt.hour, p, tt.maxP)
}
})
}
}
