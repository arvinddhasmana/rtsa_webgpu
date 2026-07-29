// CLASSIFICATION: UNCLASSIFIED
// Package domain — filter engine for track queries.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-003, CR-SEC-001 (classification enforcement)
package domain

import (
	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
)

// TrackFilter specifies filtering criteria for track queries.
// All non-zero/non-nil criteria are AND-combined.
type TrackFilter struct {
	// EntityTypes restricts results to these entity types (empty = all types).
	EntityTypes []commonv1.EntityType
	// HostileClasses restricts results to these hostile classifications (empty = all).
	HostileClasses []commonv1.HostileClassification
	// BoundingBox restricts results to tracks within this geographic area (nil = global).
	BoundingBox *commonv1.BoundingBox
	// MinConfidence requires track.confidence_score ≥ this value.
	MinConfidence float64
	// ClearanceLevel is the caller's clearance; tracks with higher classification are excluded.
	// CLASSIFICATION_LEVEL_UNSPECIFIED (0) is treated as unclassified.
	ClearanceLevel commonv1.ClassificationLevel
}

// FilterEngine applies a TrackFilter to a slice of FusedTracks.
// FilterEngine is stateless and safe for concurrent use.
type FilterEngine struct{}

// Apply returns the subset of tracks that match all criteria in filter.
// Criteria are AND-combined:
//  1. Classification: track.classification ≤ caller's clearanceLevel
//  2. EntityType:     track.entity_type IN filter.EntityTypes  (skipped if empty)
//  3. HostileClass:   track.hostile_class IN filter.HostileClasses  (skipped if empty)
//  4. BoundingBox:    track position within bbox  (skipped if nil)
//  5. MinConfidence:  track.confidence_score ≥ filter.MinConfidence
func (f *FilterEngine) Apply(tracks []*entityv1.FusedTrack, filter *TrackFilter) []*entityv1.FusedTrack {
	if filter == nil {
		return tracks
	}
	result := make([]*entityv1.FusedTrack, 0, len(tracks))
	for _, t := range tracks {
		if f.matches(t, filter) {
			result = append(result, t)
		}
	}
	return result
}

// matches reports whether a single track passes all filter criteria.
func (f *FilterEngine) matches(t *entityv1.FusedTrack, filter *TrackFilter) bool {
	// 1. Classification guard — MANDATORY.
	// A clearance level of 0 (UNSPECIFIED) is treated as UNCLASSIFIED (level 1).
	clearance := filter.ClearanceLevel
	if clearance == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
		clearance = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
	}
	if t.Classification > clearance {
		return false
	}

	// 2. Entity type filter.
	if len(filter.EntityTypes) > 0 && !containsEntityType(filter.EntityTypes, t.EntityType) {
		return false
	}

	// 3. Hostile classification filter.
	if len(filter.HostileClasses) > 0 && !containsHostileClass(filter.HostileClasses, t.HostileClass) {
		return false
	}

	// 4. Bounding box filter.
	if filter.BoundingBox != nil {
		pos := t.EstimatedPosition
		if pos == nil {
			// A track without a position cannot satisfy a spatial filter.
			return false
		}
		if !InBoundingBox(pos.Latitude, pos.Longitude, filter.BoundingBox) {
			return false
		}
	}

	// 5. Minimum confidence filter.
	if t.ConfidenceScore < filter.MinConfidence {
		return false
	}

	return true
}

// InBoundingBox returns true if the given lat/lon falls within bbox (inclusive boundaries).
func InBoundingBox(lat, lon float64, bbox *commonv1.BoundingBox) bool {
	if bbox == nil {
		return true
	}
	return lat >= bbox.MinLatitude &&
		lat <= bbox.MaxLatitude &&
		lon >= bbox.MinLongitude &&
		lon <= bbox.MaxLongitude
}

// Matches reports whether a single track passes all filter criteria.
// Exported so handlers can apply the filter to individual streamed updates.
func (f *FilterEngine) Matches(t *entityv1.FusedTrack, filter *TrackFilter) bool {
	if t == nil {
		return false
	}
	if filter == nil {
		return true
	}
	return f.matches(t, filter)
}

func containsEntityType(allowed []commonv1.EntityType, target commonv1.EntityType) bool {
	for _, et := range allowed {
		if et == target {
			return true
		}
	}
	return false
}

// containsHostileClass reports whether the target classification is present in the allowed list.
func containsHostileClass(allowed []commonv1.HostileClassification, target commonv1.HostileClassification) bool {
	for _, hc := range allowed {
		if hc == target {
			return true
		}
	}
	return false
}
