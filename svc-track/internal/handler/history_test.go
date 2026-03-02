// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"testing"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// T13: GetTrackHistory — returns history points for a known track.
func TestHistoryHandler_GetTrackHistory_ReturnsPoints(t *testing.T) {
	cache := domain.NewTrackCache(50)
	// Insert 5 updates for the same track.
	for i := 0; i < 5; i++ {
		cache.Put(makeTrack("hist-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, float64(i)*0.1+0.5))
	}

	h := NewHistoryHandler(cache)
	resp, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{
		TrackId:        "hist-001",
		MaxPoints:      10,
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TrackId != "hist-001" {
		t.Errorf("expected track_id=hist-001, got %q", resp.TrackId)
	}
	if len(resp.Points) != 5 {
		t.Errorf("expected 5 history points, got %d", len(resp.Points))
	}
	for _, pt := range resp.Points {
		if pt.Position == nil {
			t.Error("history point has nil position")
		}
	}
}

// TestHistoryHandler_NotFound: non-existent track returns NOT_FOUND.
func TestHistoryHandler_NotFound(t *testing.T) {
	h := NewHistoryHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{
		TrackId:        "ghost",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v", err)
	}
}

// TestHistoryHandler_ClassificationBlocked: SECRET track with PROTECTED_B clearance.
func TestHistoryHandler_ClassificationBlocked(t *testing.T) {
	cache := domain.NewTrackCache(10)
	cache.Put(makeTrack("hist-secret", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 0.9))

	h := NewHistoryHandler(cache)
	_, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{
		TrackId:        "hist-secret",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected NOT_FOUND for classification violation, got %v", err)
	}
}

// TestHistoryHandler_EmptyTrackID: returns InvalidArgument.
func TestHistoryHandler_EmptyTrackID(t *testing.T) {
	h := NewHistoryHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackHistory(context.Background(), &entityv1.GetTrackHistoryRequest{TrackId: ""})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}
