// CLASSIFICATION: UNCLASSIFIED
// Package mapper — unit tests for track mapping helpers.
package mapper

import (
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-track/internal/domain"
)

func TestToTrackFilter_FromStreamRequest(t *testing.T) {
	req := &entityv1.StreamTracksRequest{
		EntityTypes:    []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_AIR},
		HostileClasses: []commonv1.HostileClassification{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE},
		MinConfidence:  0.7,
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
		BoundingBox: &commonv1.BoundingBox{
			MinLatitude: 44.0, MaxLatitude: 47.0,
			MinLongitude: -77.0, MaxLongitude: -73.0,
		},
	}
	f := ToTrackFilter(req)
	if len(f.EntityTypes) != 1 || f.EntityTypes[0] != commonv1.EntityType_ENTITY_TYPE_AIR {
		t.Errorf("EntityTypes mismatch")
	}
	if f.MinConfidence != 0.7 {
		t.Errorf("MinConfidence mismatch: %v", f.MinConfidence)
	}
	if f.ClearanceLevel != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
		t.Errorf("ClearanceLevel mismatch")
	}
	if f.BoundingBox == nil {
		t.Error("expected BoundingBox, got nil")
	}
}

func TestToTrackFilter_NilRequest(t *testing.T) {
	f := ToTrackFilter(nil)
	if f == nil {
		t.Error("ToTrackFilter(nil) must not return nil")
	}
}

func TestToDetailsFilter(t *testing.T) {
	req := &entityv1.GetTrackDetailsRequest{
		TrackId:        "test-trk",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	}
	f := ToDetailsFilter(req)
	if f.ClearanceLevel != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B {
		t.Errorf("ClearanceLevel mismatch")
	}
}

func TestToDetailsFilter_Nil(t *testing.T) {
	f := ToDetailsFilter(nil)
	if f == nil {
		t.Error("expected non-nil TrackFilter")
	}
}

func TestToHistoryFilter(t *testing.T) {
	req := &entityv1.GetTrackHistoryRequest{
		TrackId:        "hist-trk",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}
	f := ToHistoryFilter(req)
	if f.ClearanceLevel != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
		t.Errorf("ClearanceLevel mismatch")
	}
}

func TestToHistoryFilter_Nil(t *testing.T) {
	f := ToHistoryFilter(nil)
	if f == nil {
		t.Error("expected non-nil TrackFilter")
	}
}

func TestToProtoHistoryPoints_Empty(t *testing.T) {
	result := ToProtoHistoryPoints(nil)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestToProtoHistoryPoints_NonEmpty(t *testing.T) {
	pts := []*domain.HistoryPoint{
		{
			Position:   &commonv1.Position{Latitude: 45.0, Longitude: -75.0},
			Timestamp:  time.Now(),
			Confidence: 0.9,
			Status:     commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		},
		{
			Position:   &commonv1.Position{Latitude: 45.1, Longitude: -75.1},
			Timestamp:  time.Now(),
			Confidence: 0.85,
			Status:     commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		},
	}
	result := ToProtoHistoryPoints(pts)
	if len(result) != 2 {
		t.Errorf("expected 2 proto points, got %d", len(result))
	}
	if result[0].Confidence != 0.9 {
		t.Errorf("confidence mismatch: %v", result[0].Confidence)
	}
}

func TestTrackPassesClassification(t *testing.T) {
	tests := []struct {
		name      string
		trackCls  commonv1.ClassificationLevel
		clearance commonv1.ClassificationLevel
		wantPass  bool
	}{
		{
			name:      "unclassified passes unclassified clearance",
			trackCls:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			clearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			wantPass:  true,
		},
		{
			name:      "secret blocked by protected_b clearance",
			trackCls:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			clearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
			wantPass:  false,
		},
		{
			name:      "secret passes secret clearance",
			trackCls:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			clearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			wantPass:  true,
		},
		{
			name:      "nil track returns false",
			trackCls:  0,
			clearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			wantPass:  false,
		},
		{
			name:      "unspecified clearance treated as unclassified",
			trackCls:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			clearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED,
			wantPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var track *entityv1.FusedTrack
			if tt.name != "nil track returns false" {
				track = &entityv1.FusedTrack{
					TrackId:        "test",
					Classification: tt.trackCls,
				}
			}
			got := TrackPassesClassification(track, tt.clearance)
			if got != tt.wantPass {
				t.Errorf("expected %v, got %v", tt.wantPass, got)
			}
		})
	}
}
