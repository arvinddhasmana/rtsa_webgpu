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

// T11: GetTrackDetails — existing track returns full track.
func TestDetailsHandler_GetTrackDetails_Exists(t *testing.T) {
	cache := domain.NewTrackCache(10)
	tr := makeTrack("detail-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9)
	cache.Put(tr)

	h := NewDetailsHandler(cache)
	resp, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId:        "detail-001",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.TrackId != "detail-001" {
		t.Errorf("expected track_id=detail-001, got %q", resp.TrackId)
	}
}

// T12: GetTrackDetails — non-existent track returns NOT_FOUND.
func TestDetailsHandler_GetTrackDetails_NotFound(t *testing.T) {
	cache := domain.NewTrackCache(10)
	h := NewDetailsHandler(cache)

	_, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId:        "nonexistent",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	})
	if err == nil {
		t.Fatal("expected NOT_FOUND error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v", err)
	}
}

// T14 (details): SECRET track with PROTECTED_B clearance returns NOT_FOUND.
func TestDetailsHandler_GetTrackDetails_ClassificationBlocked(t *testing.T) {
	cache := domain.NewTrackCache(10)
	tr := makeTrack("secret-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET, 0.9)
	cache.Put(tr)

	h := NewDetailsHandler(cache)
	_, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId:        "secret-001",
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	})
	if err == nil {
		t.Fatal("expected NOT_FOUND for SECRET track with PROTECTED_B clearance")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Errorf("expected codes.NotFound, got %v", err)
	}
}

// TestDetailsHandler_NilRequest: nil request returns InvalidArgument.
func TestDetailsHandler_NilRequest(t *testing.T) {
	h := NewDetailsHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackDetails(context.Background(), nil)
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for nil request, got %v", err)
	}
}

// TestDetailsHandler_EmptyTrackID: empty track_id returns InvalidArgument.
func TestDetailsHandler_EmptyTrackID(t *testing.T) {
	h := NewDetailsHandler(domain.NewTrackCache(10))
	_, err := h.GetTrackDetails(context.Background(), &entityv1.GetTrackDetailsRequest{
		TrackId: "",
	})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument for empty track_id, got %v", err)
	}
}
