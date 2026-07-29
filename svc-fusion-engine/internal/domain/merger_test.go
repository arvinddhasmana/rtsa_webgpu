// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/domain"
)

func defaultScorerForMerge() *domain.CorrelationScorer {
	return domain.NewCorrelationScorer(0.35, 0.25, 0.20, 0.20, 0.85, 0.60)
}

func surfaceTrack(id string, lat, lon float64) *domain.TrackState {
	now := time.Now()
	return &domain.TrackState{
		TrackID:    id,
		EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
		Status:     commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		UpdatedAt:  now,
		KalmanState: &domain.KalmanState{
			Latitude:   lat,
			Longitude:  lon,
			LastUpdate: now,
		},
	}
}

// T16/T22 — Two identical co-located tracks should be merge candidates
func TestFindMergeCandidates_CoLocated(t *testing.T) {
	scorer := defaultScorerForMerge()
	gate := domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second}

	a := surfaceTrack("A", 45.0, -60.0)
	b := surfaceTrack("B", 45.0, -60.0) // identical position

	pairs := domain.FindMergeCandidates([]*domain.TrackState{a, b}, scorer, gate, 0.85)
	if len(pairs) == 0 {
		t.Error("expected co-located tracks to be merge candidates")
	}
}

// Distant tracks must not be merge candidates
func TestFindMergeCandidates_Distant(t *testing.T) {
	scorer := defaultScorerForMerge()
	gate := domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second}

	a := surfaceTrack("A", 45.0, -60.0)
	b := surfaceTrack("B", 50.0, -60.0) // ~300 NM away

	pairs := domain.FindMergeCandidates([]*domain.TrackState{a, b}, scorer, gate, 0.85)
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs for distant tracks, got %d", len(pairs))
	}
}

// Nil KalmanState tracks are skipped
func TestFindMergeCandidates_NilKalmanState(t *testing.T) {
	scorer := defaultScorerForMerge()
	gate := domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second}

	a := surfaceTrack("A", 45.0, -60.0)
	b := &domain.TrackState{TrackID: "B", EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE}

	pairs := domain.FindMergeCandidates([]*domain.TrackState{a, b}, scorer, gate, 0.85)
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs when KalmanState is nil, got %d", len(pairs))
	}
}

// Empty track list returns empty pairs
func TestFindMergeCandidates_Empty(t *testing.T) {
	scorer := defaultScorerForMerge()
	gate := domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second}

	pairs := domain.FindMergeCandidates(nil, scorer, gate, 0.85)
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs for empty track list, got %d", len(pairs))
	}
}
