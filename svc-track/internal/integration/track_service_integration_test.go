// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration_test provides integration tests for svc-track.
// These tests exercise the consumer → cache pipeline using proto marshaling/unmarshaling
// to simulate the message flow, and the cache+filter engine.
package integration

import (
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-track/internal/domain"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestFusedTrackConsumer_IntegrationFlow tests the full consumer → cache flow:
// proto.Marshal → proto.Unmarshal → cache.Put → cache.Get.
//
// The test deliberately avoids spinning up a live Redpanda container because
// svc-track's go.mod does not include testcontainers-go/modules/redpanda and the
// generic container approach cannot resolve the broker's advertised address from
// the test host. End-to-end Redpanda connectivity is covered by the centralized
// ingestion/fusion integration tests that use tcredpanda.
func TestFusedTrackConsumer_IntegrationFlow(t *testing.T) {
	testTrack := &entityv1.FusedTrack{
		TrackId:         "integ-trk-001",
		EntityType:      commonv1.EntityType_ENTITY_TYPE_AIR,
		HostileClass:    commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
		Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		ConfidenceScore: 0.88,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		EstimatedPosition: &commonv1.Position{
			Latitude:  45.4215,
			Longitude: -75.6972,
		},
		UpdatedAt: timestamppb.New(time.Now()),
	}

	// Simulate consumer Kafka record payload: marshal → unmarshal.
	payload, err := proto.Marshal(testTrack)
	if err != nil {
		t.Fatalf("TestFusedTrackConsumer_IntegrationFlow: proto.Marshal: %v", err)
	}
	var decoded entityv1.FusedTrack
	if err := proto.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("TestFusedTrackConsumer_IntegrationFlow: proto.Unmarshal: %v", err)
	}

	// Act: simulate consumer processRecord → cache.Put.
	cache := domain.NewTrackCache(100)
	cache.Put(&decoded)

	// Assert: track is retrievable with correct fields.
	got := cache.Get("integ-trk-001")
	if got == nil {
		t.Fatal("TestFusedTrackConsumer_IntegrationFlow: track not found in cache after Put")
	}
	if got.TrackId != "integ-trk-001" {
		t.Errorf("TestFusedTrackConsumer_IntegrationFlow: track_id=%q, want %q", got.TrackId, "integ-trk-001")
	}
	if got.ConfidenceScore != 0.88 {
		t.Errorf("TestFusedTrackConsumer_IntegrationFlow: confidence_score=%f, want 0.88", got.ConfidenceScore)
	}
	if got.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
		t.Errorf("TestFusedTrackConsumer_IntegrationFlow: classification=%v, want UNCLASSIFIED", got.Classification)
	}
	t.Logf("TestFusedTrackConsumer_IntegrationFlow PASS: track id=%s, confidence=%.2f",
		got.TrackId, got.ConfidenceScore)
}

// TestTrackCache_FilterIntegration tests cache + filter engine together.
func TestTrackCache_FilterIntegration(t *testing.T) {
	cache := domain.NewTrackCache(50)
	filter := &domain.FilterEngine{}

	// Insert mixed tracks.
	tracks := []*entityv1.FusedTrack{
		{
			TrackId: "fi-air-uncl", EntityType: commonv1.EntityType_ENTITY_TYPE_AIR,
			Status:         commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			ConfidenceScore: 0.9,
			EstimatedPosition: &commonv1.Position{Latitude: 45.4, Longitude: -75.6},
			UpdatedAt:         timestamppb.Now(),
		},
		{
			TrackId: "fi-surface-protb", EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
			Status:         commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
			ConfidenceScore: 0.7,
			EstimatedPosition: &commonv1.Position{Latitude: 46.0, Longitude: -76.0},
			UpdatedAt:         timestamppb.Now(),
		},
		{
			TrackId: "fi-cyber-secret", EntityType: commonv1.EntityType_ENTITY_TYPE_CYBER,
			Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
			ConfidenceScore: 0.95,
			UpdatedAt:       timestamppb.Now(),
		},
	}
	for _, tr := range tracks {
		cache.Put(tr)
	}

	// Filter with PROTECTED_B clearance — should see unclassified and PROTECTED_B.
	f := &domain.TrackFilter{
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
	}
	result := cache.GetFiltered(f)
	if len(result) != 2 {
		t.Errorf("expected 2 tracks with PROTECTED_B clearance, got %d", len(result))
	}

	// Filter by entity type AIR + bbox.
	f2 := &domain.TrackFilter{
		EntityTypes:    []commonv1.EntityType{commonv1.EntityType_ENTITY_TYPE_AIR},
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
		BoundingBox: &commonv1.BoundingBox{
			MinLatitude: 44.0, MaxLatitude: 47.0,
			MinLongitude: -77.0, MaxLongitude: -74.0,
		},
	}
	result2 := filter.Apply(cache.GetAll(), f2)
	if len(result2) != 1 {
		t.Errorf("expected 1 AIR track in bbox, got %d", len(result2))
	}

	t.Log("Filter integration test passed")
}
