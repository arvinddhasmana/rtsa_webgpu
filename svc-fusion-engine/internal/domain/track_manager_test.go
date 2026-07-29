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

func newTM() *domain.TrackManager {
	return domain.NewTrackManager(domain.NewKalmanFilter())
}

func radarObs(sensorID string, lat, lon float64) *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        sensorID,
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(time.Now()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		Position: &commonv1.Position{
			Latitude:  lat,
			Longitude: lon,
		},
	}
}

// T12 — CreateTrack: sets UUID
func TestTrackManager_CreateTrack_SetsUUID(t *testing.T) {
	tm := newTM()
	obs := radarObs("RADAR-01", 45.0, -60.0)
	track, err := tm.CreateTrack(obs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.TrackID == "" {
		t.Error("expected non-empty track ID")
	}
}

// T13 — CreateTrack: initial Kalman state matches observation
func TestTrackManager_CreateTrack_KalmanState(t *testing.T) {
	tm := newTM()
	obs := radarObs("RADAR-01", 45.123, -60.456)
	track, err := tm.CreateTrack(obs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if track.KalmanState == nil {
		t.Fatal("expected KalmanState to be initialized")
	}
	if track.KalmanState.Latitude != 45.123 {
		t.Errorf("expected lat=45.123, got %.6f", track.KalmanState.Latitude)
	}
	if track.KalmanState.Longitude != -60.456 {
		t.Errorf("expected lon=-60.456, got %.6f", track.KalmanState.Longitude)
	}
}

// CreateTrack with nil position must return error
func TestTrackManager_CreateTrack_NilPosition(t *testing.T) {
	tm := newTM()
	obs := &ingestionv1.SensorObservation{
		SensorId:        "RADAR-01",
		ObservationTime: timestamppb.New(time.Now()),
	}
	_, err := tm.CreateTrack(obs)
	if err == nil {
		t.Error("expected error for nil position")
	}
}

// T14 — UpdateTrack: source attribution
func TestTrackManager_UpdateTrack_SourceAttribution(t *testing.T) {
	tm := newTM()
	obs1 := radarObs("RADAR-01", 45.0, -60.0)
	track, _ := tm.CreateTrack(obs1)

	obs2 := radarObs("RADAR-02", 45.001, -60.001)
	obs2.ObservationTime = timestamppb.New(time.Now().Add(time.Second))
	_, err := tm.UpdateTrack(track.TrackID, obs2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(track.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(track.Sources))
	}
}

// Updating same sensor increments observation count
func TestTrackManager_UpdateTrack_ObservationCountIncrements(t *testing.T) {
	tm := newTM()
	obs := radarObs("RADAR-01", 45.0, -60.0)
	track, _ := tm.CreateTrack(obs)

	obs2 := radarObs("RADAR-01", 45.001, -60.001)
	obs2.ObservationTime = timestamppb.New(time.Now().Add(time.Second))
	tm.UpdateTrack(track.TrackID, obs2)

	src := track.Sources["RADAR-01"]
	if src == nil || src.ObservationCount != 2 {
		t.Errorf("expected observation count 2, got %v", src)
	}
}

// T15 — UpdateTrack: classification propagation (MAX applied)
func TestTrackManager_UpdateTrack_ClassificationPropagation(t *testing.T) {
	tm := newTM()
	obs1 := radarObs("RADAR-01", 45.0, -60.0)
	obs1.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
	track, _ := tm.CreateTrack(obs1)

	obs2 := radarObs("RADAR-02", 45.001, -60.001)
	obs2.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
	obs2.ObservationTime = timestamppb.New(time.Now().Add(time.Second))
	tm.UpdateTrack(track.TrackID, obs2)

	if track.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET {
		t.Errorf("expected SECRET classification, got %v", track.Classification)
	}
}

// T16 — MergeTracks: sources combined
func TestTrackManager_MergeTracks_SourcesCombined(t *testing.T) {
	tm := newTM()
	trackA, _ := tm.CreateTrack(radarObs("RADAR-A", 45.0, -60.0))
	trackB, _ := tm.CreateTrack(radarObs("RADAR-B", 45.001, -60.001))

	merged, err := tm.MergeTracks(trackA.TrackID, trackB.TrackID)
	if err != nil {
		t.Fatalf("MergeTracks error: %v", err)
	}
	if len(merged.Sources) < 2 {
		t.Errorf("expected ≥2 sources after merge, got %d", len(merged.Sources))
	}
}

// T17 — MergeTracks: source track marked MERGED
func TestTrackManager_MergeTracks_SourceMarkedMerged(t *testing.T) {
	tm := newTM()
	trackA, _ := tm.CreateTrack(radarObs("RADAR-A", 45.0, -60.0))
	trackB, _ := tm.CreateTrack(radarObs("RADAR-B", 45.001, -60.001))

	tm.MergeTracks(trackA.TrackID, trackB.TrackID)

	if trackB.Status != commonv1.TrackStatus_TRACK_STATUS_MERGED {
		t.Errorf("expected trackB status=MERGED, got %v", trackB.Status)
	}
}

// MergeTracks with unknown track returns error
func TestTrackManager_MergeTracks_UnknownTrack(t *testing.T) {
	tm := newTM()
	trackA, _ := tm.CreateTrack(radarObs("RADAR-A", 45.0, -60.0))
	_, err := tm.MergeTracks(trackA.TrackID, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown track B")
	}
}

// GetActiveTracks excludes DROPPED/MERGED
func TestTrackManager_GetActiveTracks_ExcludesTerminated(t *testing.T) {
	tm := newTM()
	active, _ := tm.CreateTrack(radarObs("R1", 45.0, -60.0))
	dropped, _ := tm.CreateTrack(radarObs("R2", 45.1, -60.1))
	merged, _ := tm.CreateTrack(radarObs("R3", 45.2, -60.2))

	tm.MarkDropped(dropped.TrackID)
	tm.MarkDropped(merged.TrackID)

	tracks := tm.GetActiveTracks()
	for _, t2 := range tracks {
		if t2.TrackID == dropped.TrackID || t2.TrackID == merged.TrackID {
			t.Errorf("dropped/merged track %s should not be in active tracks", t2.TrackID)
		}
	}
	_ = active
}

func TestTrackManager_TrackCount(t *testing.T) {
	tm := newTM()
	tm.CreateTrack(radarObs("R1", 45.0, -60.0))
	tm.CreateTrack(radarObs("R2", 45.1, -60.1))
	if tm.TrackCount() != 2 {
		t.Errorf("expected 2, got %d", tm.TrackCount())
	}
}

func TestTrackState_ToFusedTrack(t *testing.T) {
	tm := newTM()
	obs := radarObs("RADAR-01", 45.0, -60.0)
	track, _ := tm.CreateTrack(obs)

	ft := track.ToFusedTrack()
	if ft.TrackId != track.TrackID {
		t.Errorf("expected TrackId=%s, got %s", track.TrackID, ft.TrackId)
	}
	if ft.EstimatedPosition == nil {
		t.Error("expected EstimatedPosition to be set")
	}
}

func TestTrackManager_MarkStale(t *testing.T) {
	tm := newTM()
	track, _ := tm.CreateTrack(radarObs("R1", 45.0, -60.0))
	tm.MarkStale(track.TrackID)
	got, _ := tm.GetTrack(track.TrackID)
	if got.Status != commonv1.TrackStatus_TRACK_STATUS_STALE {
		t.Errorf("expected STALE, got %v", got.Status)
	}
}

func TestTrackManager_MarkDropped(t *testing.T) {
	tm := newTM()
	track, _ := tm.CreateTrack(radarObs("R1", 45.0, -60.0))
	tm.MarkDropped(track.TrackID)
	got, _ := tm.GetTrack(track.TrackID)
	if got.Status != commonv1.TrackStatus_TRACK_STATUS_DROPPED {
		t.Errorf("expected DROPPED, got %v", got.Status)
	}
}

func TestTrackManager_CreateTrack_MetadataParsing(t *testing.T) {
	tm := newTM()
	obs := radarObs("RADAR-01", 45.0, -60.0)
	obs.Metadata = map[string]string{
		"sim_hostile_class": "HOSTILE",
		"sim_entity_type":   "AIR",
	}
	track, err := tm.CreateTrack(obs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if track.HostileClass != commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE {
		t.Errorf("expected HOSTILE, got %v", track.HostileClass)
	}
	if track.EntityType != commonv1.EntityType_ENTITY_TYPE_AIR {
		t.Errorf("expected AIR, got %v", track.EntityType)
	}
}

func TestTrackManager_CreateTrack_SpeedHeading(t *testing.T) {
	tm := newTM()
	obs := radarObs("RADAR-01", 45.0, -60.0)
	speed := float64(10.0)   // 10 knots
	heading := float64(90.0) // East
	obs.Position.SpeedKnots = &speed
	obs.Position.HeadingDegrees = &heading

	track, err := tm.CreateTrack(obs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 10 knots = 5.144 m/s. Heading 90 means East. So vE ~ 5.144, vN ~ 0
	if track.KalmanState.VelocityE <= 5.0 {
		t.Errorf("expected VelocityE > 5.0, got %v", track.KalmanState.VelocityE)
	}
	if track.KalmanState.VelocityN > 0.001 || track.KalmanState.VelocityN < -0.001 {
		t.Errorf("expected VelocityN ~ 0, got %v", track.KalmanState.VelocityN)
	}
}

func TestTrackManager_ComputeConfidence(t *testing.T) {
	tm := newTM()
	obs1 := radarObs("R1", 45.0, -60.0)
	obs2 := radarObs("R2", 45.0, -60.0)
	obs3 := radarObs("R3", 45.0, -60.0)

	track, _ := tm.CreateTrack(obs1)
	ft1 := track.ToFusedTrack()
	if ft1.ConfidenceScore != 0.5 {
		t.Errorf("expected 0.5 for 1 source, got %v", ft1.ConfidenceScore)
	}

	tm.UpdateTrack(track.TrackID, obs2)
	ft2 := track.ToFusedTrack()
	if ft2.ConfidenceScore != 0.75 {
		t.Errorf("expected 0.75 for 2 sources, got %v", ft2.ConfidenceScore)
	}

	tm.UpdateTrack(track.TrackID, obs3)
	ft3 := track.ToFusedTrack()
	if ft3.ConfidenceScore != 0.9 {
		t.Errorf("expected 0.9 for 3 sources, got %v", ft3.ConfidenceScore)
	}
}
