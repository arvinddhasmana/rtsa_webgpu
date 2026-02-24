// CLASSIFICATION: UNCLASSIFIED
// Package domain — unit tests for the FilterEngine.
//
// Test coverage for T05–T08, T14 per Module 10 specification.
package domain

import (
	"testing"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
)

func makeFilterTrack(id string, et commonv1.EntityType, hc commonv1.HostileClassification, cls commonv1.ClassificationLevel, confidence float64, lat, lon float64) *entityv1.FusedTrack {
	return &entityv1.FusedTrack{
		TrackId:         id,
		EntityType:      et,
		HostileClass:    hc,
		Classification:  cls,
		ConfidenceScore: confidence,
		Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		EstimatedPosition: &commonv1.Position{
			Latitude:  lat,
			Longitude: lon,
		},
	}
}

func TestFilterEngine_Apply_NoFilter(t *testing.T) {
	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("t1", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9, 45.0, -75.0),
		makeFilterTrack("t2", commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8, 46.0, -74.0),
	}
	result := fe.Apply(tracks, nil)
	if len(result) != 2 {
		t.Errorf("nil filter should return all tracks, got %d", len(result))
	}
}

// T05: Filter by entity type — only matching types returned.
func TestFilterEngine_Apply_EntityType(t *testing.T) {
	tests := []struct {
		name          string
		filter        []commonv1.EntityType
		inputCount    int
		expectedCount int
	}{
		{
			name:          "filter to AIR only",
			filter:        []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_AIR},
			inputCount:    3,
			expectedCount: 1,
		},
		{
			name:          "filter to AIR and SURFACE",
			filter:        []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.EntityType_ENTITY_TYPE_SURFACE},
			inputCount:    3,
			expectedCount: 2,
		},
		{
			name:          "empty filter = all types",
			filter:        nil,
			inputCount:    3,
			expectedCount: 3,
		},
	}

	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("e1", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9, 45.0, -75.0),
		makeFilterTrack("e2", commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8, 46.0, -74.0),
		makeFilterTrack("e3", commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.7, 47.0, -73.0),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TrackFilter{
				EntityTypes:    tt.filter,
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			}
			result := fe.Apply(tracks, filter)
			if len(result) != tt.expectedCount {
				t.Errorf("expected %d tracks, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

// T06 / T14: Filter by classification clearance — higher classified tracks excluded.
func TestFilterEngine_Apply_ClassificationClearance(t *testing.T) {
	tests := []struct {
		name           string
		clearance      commonv1.ClassificationLevel
		expectedIDs    []string
		excludedIDs    []string
	}{
		{
			name:        "UNCLASSIFIED clearance sees only UNCLASSIFIED",
			clearance:   commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			expectedIDs: []string{"cls-uncl"},
			excludedIDs: []string{"cls-protb", "cls-secret"},
		},
		{
			name:        "PROTECTED_B clearance sees UNCLASSIFIED and PROTECTED_B",
			clearance:   commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
			expectedIDs: []string{"cls-uncl", "cls-protb"},
			excludedIDs: []string{"cls-secret"},
		},
		{
			name:        "SECRET clearance sees all",
			clearance:   commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			expectedIDs: []string{"cls-uncl", "cls-protb", "cls-secret"},
			excludedIDs: nil,
		},
	}

	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("cls-uncl", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9, 45.0, -75.0),
		makeFilterTrack("cls-protb", commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B, 0.8, 46.0, -74.0),
		makeFilterTrack("cls-secret", commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 0.7, 47.0, -73.0),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TrackFilter{ClearanceLevel: tt.clearance}
			result := fe.Apply(tracks, filter)
			if len(result) != len(tt.expectedIDs) {
				t.Errorf("expected %d tracks, got %d", len(tt.expectedIDs), len(result))
			}
			resultIDs := make(map[string]bool)
			for _, r := range result {
				resultIDs[r.TrackId] = true
			}
			for _, id := range tt.expectedIDs {
				if !resultIDs[id] {
					t.Errorf("expected track %q to be in results", id)
				}
			}
			for _, id := range tt.excludedIDs {
				if resultIDs[id] {
					t.Errorf("track %q should be excluded (clearance violation)", id)
				}
			}
		})
	}
}

// T14: SECRET track with PROTECTED_B clearance — track excluded.
func TestFilterEngine_Apply_SecretTrackProtectedBClearance_Excluded(t *testing.T) {
	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("secret-trk", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 0.95, 45.0, -75.0),
	}
	filter := &TrackFilter{
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	}
	result := fe.Apply(tracks, filter)
	if len(result) != 0 {
		t.Errorf("SECRET track must be excluded for PROTECTED_B clearance, got %d tracks", len(result))
	}
}

// T07: Filter by bounding box — only within bbox returned.
func TestFilterEngine_Apply_BoundingBox(t *testing.T) {
	tests := []struct {
		name          string
		bbox          *commonv1.BoundingBox
		expectedCount int
	}{
		{
			name: "bbox includes only Ottawa area",
			bbox: &commonv1.BoundingBox{
				MinLatitude:  45.0, MaxLatitude: 46.0,
				MinLongitude: -76.0, MaxLongitude: -75.0,
			},
			expectedCount: 1,
		},
		{
			name: "bbox includes all tracks",
			bbox: &commonv1.BoundingBox{
				MinLatitude:  40.0, MaxLatitude: 50.0,
				MinLongitude: -80.0, MaxLongitude: -70.0,
			},
			expectedCount: 3,
		},
		{
			name:          "nil bbox = no spatial filter",
			bbox:          nil,
			expectedCount: 3,
		},
	}

	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("b1", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9, 45.4, -75.6),  // Ottawa
		makeFilterTrack("b2", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8, 43.6, -79.3),  // Toronto
		makeFilterTrack("b3", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.7, 45.5, -73.5),  // Montreal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TrackFilter{
				BoundingBox:    tt.bbox,
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			}
			result := fe.Apply(tracks, filter)
			if len(result) != tt.expectedCount {
				t.Errorf("expected %d tracks, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

// TestFilterEngine_Apply_BboxNoPosition: Track without position excluded when bbox filter set.
func TestFilterEngine_Apply_BboxNoPosition(t *testing.T) {
	fe := &FilterEngine{}
	trackNoPos := &entityv1.FusedTrack{
		TrackId:         "no-pos",
		EntityType:      commonv1.EntityType_ENTITY_TYPE_CYBER,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ConfidenceScore: 0.9,
		Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		// EstimatedPosition is nil intentionally.
	}
	filter := &TrackFilter{
		BoundingBox: &commonv1.BoundingBox{
			MinLatitude: 40.0, MaxLatitude: 50.0,
			MinLongitude: -80.0, MaxLongitude: -70.0,
		},
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}
	result := fe.Apply([]*entityv1.FusedTrack{trackNoPos}, filter)
	if len(result) != 0 {
		t.Errorf("track without position should be excluded by bbox filter, got %d", len(result))
	}
}

// T08: Filter by min confidence — low confidence excluded.
func TestFilterEngine_Apply_MinConfidence(t *testing.T) {
	tests := []struct {
		name          string
		minConf       float64
		expectedCount int
	}{
		{name: "min 0.8 — only high confidence", minConf: 0.8, expectedCount: 2},
		{name: "min 0.5 — medium+high", minConf: 0.5, expectedCount: 3},
		{name: "min 0.0 — all", minConf: 0.0, expectedCount: 4},
		{name: "min 1.0 — none", minConf: 1.0, expectedCount: 0},
	}

	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("conf-1", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.95, 45.0, -75.0),
		makeFilterTrack("conf-2", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.85, 45.0, -75.0),
		makeFilterTrack("conf-3", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.60, 45.0, -75.0),
		makeFilterTrack("conf-4", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.30, 45.0, -75.0),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &TrackFilter{
				MinConfidence:  tt.minConf,
				ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			}
			result := fe.Apply(tracks, filter)
			if len(result) != tt.expectedCount {
				t.Errorf("expected %d tracks with min_confidence=%.2f, got %d", tt.expectedCount, tt.minConf, len(result))
			}
		})
	}
}

// TestInBoundingBox: boundary conditions.
func TestInBoundingBox(t *testing.T) {
	bbox := &commonv1.BoundingBox{
		MinLatitude: 44.0, MaxLatitude: 47.0,
		MinLongitude: -77.0, MaxLongitude: -73.0,
	}
	tests := []struct {
		name string
		lat  float64
		lon  float64
		want bool
	}{
		{name: "inside", lat: 45.5, lon: -75.5, want: true},
		{name: "on min lat boundary", lat: 44.0, lon: -75.5, want: true},
		{name: "on max lat boundary", lat: 47.0, lon: -75.5, want: true},
		{name: "below min lat", lat: 43.9, lon: -75.5, want: false},
		{name: "above max lat", lat: 47.1, lon: -75.5, want: false},
		{name: "left of min lon", lat: 45.5, lon: -77.1, want: false},
		{name: "right of max lon", lat: 45.5, lon: -72.9, want: false},
		{name: "nil bbox", lat: 0, lon: 0, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bboxArg := bbox
			if tt.name == "nil bbox" {
				bboxArg = nil
			}
			got := InBoundingBox(tt.lat, tt.lon, bboxArg)
			if got != tt.want {
				t.Errorf("InBoundingBox(%v, %v) = %v, want %v", tt.lat, tt.lon, got, tt.want)
			}
		})
	}
}

// TestFilterEngine_Apply_HostileClass: hostile class filter.
func TestFilterEngine_Apply_HostileClass(t *testing.T) {
	fe := &FilterEngine{}
	tracks := []*entityv1.FusedTrack{
		makeFilterTrack("hc1", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9, 45.0, -75.0),
		makeFilterTrack("hc2", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8, 45.0, -75.0),
		makeFilterTrack("hc3", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.7, 45.0, -75.0),
	}
	filter := &TrackFilter{
		HostileClasses: []commonv1.HostileClassification{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE},
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}
	result := fe.Apply(tracks, filter)
	if len(result) != 1 {
		t.Errorf("expected 1 hostile track, got %d", len(result))
	}
	if result[0].TrackId != "hc1" {
		t.Errorf("expected track hc1, got %q", result[0].TrackId)
	}
}

// TestFilterEngine_Matches_Direct: tests Matches exported method.
func TestFilterEngine_Matches_Direct(t *testing.T) {
fe := &FilterEngine{}
track := makeFilterTrack("direct-01", commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9, 45.0, -75.0)

if !fe.Matches(track, nil) {
t.Error("nil filter should always pass")
}
if fe.Matches(nil, nil) {
t.Error("nil track should not pass")
}
filter := &TrackFilter{ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED}
if !fe.Matches(track, filter) {
t.Error("matching track should pass filter")
}
}
