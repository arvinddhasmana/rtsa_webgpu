// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// T01 — Haversine: known distance NYC → London ≈ 2998–3010 NM
func TestHaversine_NYCToLondon(t *testing.T) {
	// NYC: 40.7128°N, 74.0060°W  /  London: 51.5074°N, 0.1278°W
	dist := domain.HaversineDistanceNM(40.7128, -74.0060, 51.5074, -0.1278)
	if dist < 2990 || dist > 3020 {
		t.Errorf("expected NYC→London ≈ 3000 NM, got %.2f NM", dist)
	}
}

// T02 — Haversine: same point = 0
func TestHaversine_SamePoint(t *testing.T) {
	dist := domain.HaversineDistanceNM(45.0, -60.0, 45.0, -60.0)
	if dist != 0.0 {
		t.Errorf("expected 0 NM for same point, got %.6f", dist)
	}
}

func newGatingFilter() *domain.GatingFilter {
	return domain.NewGatingFilter(
		domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second},
		domain.GatingConfig{MaxDistanceNM: 20.0, MaxTimeDelta: 15 * time.Second},
		domain.GatingConfig{MaxDistanceNM: 2.0, MaxTimeDelta: 60 * time.Second},
	)
}

func newRadarObs(lat, lon float64, t time.Time) *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        "RADAR-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(t),
		Position: &commonv1.Position{
			Latitude:  lat,
			Longitude: lon,
		},
	}
}

func newSurfaceTrack(lat, lon float64, updatedAt time.Time) *domain.TrackState {
	return &domain.TrackState{
		TrackID:    "track-1",
		EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
		Status:     commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		UpdatedAt:  updatedAt,
		KalmanState: &domain.KalmanState{
			Latitude:   lat,
			Longitude:  lon,
			LastUpdate: updatedAt,
		},
	}
}

// T03 — Gating: observation within gate → in candidates
func TestGatingFilter_WithinGate(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := newRadarObs(45.0, -60.0, now)
	track := newSurfaceTrack(45.01, -60.01, now.Add(-5*time.Second)) // ~0.8 NM away, 5s delta

	candidates := gf.FindCandidates(obs, []*domain.TrackState{track})
	if len(candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(candidates))
	}
}

// T04 — Gating: observation outside distance gate → not in candidates
func TestGatingFilter_OutsideDistance(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := newRadarObs(45.0, -60.0, now)
	track := newSurfaceTrack(46.0, -60.0, now.Add(-5*time.Second)) // >5 NM away

	candidates := gf.FindCandidates(obs, []*domain.TrackState{track})
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(candidates))
	}
}

// T05 — Gating: observation outside time gate → not in candidates
func TestGatingFilter_OutsideTime(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := newRadarObs(45.0, -60.0, now)
	track := newSurfaceTrack(45.001, -60.001, now.Add(-60*time.Second)) // within distance, stale time

	candidates := gf.FindCandidates(obs, []*domain.TrackState{track})
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates outside time gate, got %d", len(candidates))
	}
}

// T06 — Gating: wrong entity type → not in candidates
func TestGatingFilter_WrongEntityType(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := newRadarObs(45.0, -60.0, now)
	track := newSurfaceTrack(45.001, -60.001, now.Add(-5*time.Second))
	track.EntityType = commonv1.EntityType_ENTITY_TYPE_CYBER // incompatible

	candidates := gf.FindCandidates(obs, []*domain.TrackState{track})
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates for wrong entity type, got %d", len(candidates))
	}
}

// DROPPED/MERGED tracks must be excluded
func TestGatingFilter_ExcludesDroppedTracks(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := newRadarObs(45.0, -60.0, now)

	dropped := newSurfaceTrack(45.001, -60.001, now.Add(-5*time.Second))
	dropped.Status = commonv1.TrackStatus_TRACK_STATUS_DROPPED
	merged := newSurfaceTrack(45.001, -60.001, now.Add(-5*time.Second))
	merged.Status = commonv1.TrackStatus_TRACK_STATUS_MERGED

	candidates := gf.FindCandidates(obs, []*domain.TrackState{dropped, merged})
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates (dropped/merged excluded), got %d", len(candidates))
	}
}

// Nearest-first ordering
func TestGatingFilter_SortedByDistance(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := newRadarObs(45.0, -60.0, now)

	close := newSurfaceTrack(45.001, -60.001, now.Add(-2*time.Second))
	close.TrackID = "close"
	far := newSurfaceTrack(45.02, -60.02, now.Add(-2*time.Second))
	far.TrackID = "far"

	candidates := gf.FindCandidates(obs, []*domain.TrackState{far, close})
	if len(candidates) < 2 || candidates[0].TrackID != "close" {
		t.Errorf("expected nearest track first, got %v", candidates)
	}
}

// Nil position observation returns empty
func TestGatingFilter_NilPosition(t *testing.T) {
	gf := newGatingFilter()
	now := time.Now()
	obs := &ingestionv1.SensorObservation{
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(now),
	}
	track := newSurfaceTrack(45.0, -60.0, now.Add(-5*time.Second))
	candidates := gf.FindCandidates(obs, []*domain.TrackState{track})
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates for nil position, got %d", len(candidates))
	}
}

// Test GatingConfigFor fallback
func TestGatingFilter_GatingConfigFor_Fallback(t *testing.T) {
	gf := newGatingFilter()
	// Unspecified should use surface fallback
	cfg := gf.GatingConfigFor(commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED)
	if cfg.MaxDistanceNM != 5.0 || cfg.MaxTimeDelta != 30*time.Second {
		t.Errorf("expected surface fallback, got %+v", cfg)
	}
}

// Test sensorTypeToEntityType logic
func TestSensorTypeToEntityType_Coverage(t *testing.T) {
	tm := domain.NewTrackManager(domain.NewKalmanFilter())

	tests := []struct {
		sensor commonv1.SensorType
		entity commonv1.EntityType
	}{
		{commonv1.SensorType_SENSOR_TYPE_EW_SIGINT, commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED},
		{commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT, commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED},
		{commonv1.SensorType_SENSOR_TYPE_ISR, commonv1.EntityType_ENTITY_TYPE_SURFACE},
		{commonv1.SensorType_SENSOR_TYPE_AIS_BFT, commonv1.EntityType_ENTITY_TYPE_SURFACE},
		{commonv1.SensorType_SENSOR_TYPE_CYBER, commonv1.EntityType_ENTITY_TYPE_CYBER},
		{commonv1.SensorType(999), commonv1.EntityType_ENTITY_TYPE_SURFACE}, // default
	}

	for _, tc := range tests {
		obs := &ingestionv1.SensorObservation{
			SensorType: tc.sensor,
			Position:   &commonv1.Position{Latitude: 0, Longitude: 0},
		}
		track, _ := tm.CreateTrack(obs)
		if track.EntityType != tc.entity {
			t.Errorf("sensor %v: expected entity %v, got %v", tc.sensor, tc.entity, track.EntityType)
		}
	}
}

func TestEntityTypesCompatible_Coverage(t *testing.T) {
	gf := newGatingFilter()
	obs := newRadarObs(0, 0, time.Now())
	track := newSurfaceTrack(0.01, 0.01, time.Now())

	// Both specific, mismatch
	track.EntityType = commonv1.EntityType_ENTITY_TYPE_AIR
	cands := gf.FindCandidates(obs, []*domain.TrackState{track})
	if len(cands) != 0 {
		t.Errorf("expected no match, got %d", len(cands))
	}

	// Obs UNSPECIFIED (ELINT)
	obs2 := newRadarObs(0, 0, time.Now())
	obs2.SensorType = commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT
	track2 := newSurfaceTrack(0.01, 0.01, time.Now())
	cands2 := gf.FindCandidates(obs2, []*domain.TrackState{track2})
	if len(cands2) != 1 {
		t.Errorf("expected match for UNSPECIFIED obs, got %d", len(cands2))
	}
}
