// CLASSIFICATION: UNCLASSIFIED
// Package mapper provides conversion helpers between domain types and proto types.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012
package mapper

import (
	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
)

// ToTrackFilter converts a StreamTracksRequest's filter fields into a domain.TrackFilter.
func ToTrackFilter(req *entityv1.StreamTracksRequest) *domain.TrackFilter {
	if req == nil {
		return &domain.TrackFilter{}
	}
	return &domain.TrackFilter{
		EntityTypes:    req.EntityTypes,
		HostileClasses: req.HostileClasses,
		BoundingBox:    req.BoundingBox,
		MinConfidence:  req.MinConfidence,
		ClearanceLevel: req.ClearanceLevel,
	}
}

// ToDetailsFilter converts a GetTrackDetailsRequest into a minimal TrackFilter
// for classification enforcement.
func ToDetailsFilter(req *entityv1.GetTrackDetailsRequest) *domain.TrackFilter {
	if req == nil {
		return &domain.TrackFilter{}
	}
	return &domain.TrackFilter{
		ClearanceLevel: req.ClearanceLevel,
	}
}

// ToHistoryFilter converts a GetTrackHistoryRequest into a minimal TrackFilter
// for classification enforcement.
func ToHistoryFilter(req *entityv1.GetTrackHistoryRequest) *domain.TrackFilter {
	if req == nil {
		return &domain.TrackFilter{}
	}
	return &domain.TrackFilter{
		ClearanceLevel: req.ClearanceLevel,
	}
}

// ToProtoHistoryPoints converts a slice of domain HistoryPoints to protobuf TrackHistoryPoints.
func ToProtoHistoryPoints(pts []*domain.HistoryPoint) []*entityv1.TrackHistoryPoint {
	if len(pts) == 0 {
		return nil
	}
	out := make([]*entityv1.TrackHistoryPoint, len(pts))
	for i, p := range pts {
		out[i] = domain.ToProtoHistoryPoint(p)
	}
	return out
}

// TrackPassesClassification returns true if the track's classification is ≤ the caller's clearance.
func TrackPassesClassification(track *entityv1.FusedTrack, clearance commonv1.ClassificationLevel) bool {
	if track == nil {
		return false
	}
	effective := clearance
	if effective == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
		effective = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
	}
	return track.Classification <= effective
}
