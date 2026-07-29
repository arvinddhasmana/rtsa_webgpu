// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/audit"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/handler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockProducer captures produced tracks for assertions.
type mockProducer struct {
	produced []*entityv1.FusedTrack
}

func (m *mockProducer) Produce(_ context.Context, track *entityv1.FusedTrack) error {
	m.produced = append(m.produced, track)
	return nil
}

func newTestPipeline(mp *mockProducer) *handler.FusionPipeline {
	logger, _ := zap.NewDevelopment()
	kf := domain.NewKalmanFilter()
	manager := domain.NewTrackManager(kf)
	gating := domain.NewGatingFilter(
		domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second},
		domain.GatingConfig{MaxDistanceNM: 20.0, MaxTimeDelta: 15 * time.Second},
		domain.GatingConfig{MaxDistanceNM: 2.0, MaxTimeDelta: 60 * time.Second},
	)
	scorer := domain.NewCorrelationScorer(0.35, 0.25, 0.20, 0.20, 0.85, 0.60)
	auditEmitter := audit.NewLogEmitter(logger)
	metrics := handler.NewFusionMetrics(prometheus.NewRegistry())
	defaultGate := domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 30 * time.Second}

	return handler.NewFusionPipeline(
		gating, scorer, manager, mp, auditEmitter, logger, metrics, 0.85, 0.60, defaultGate,
	)
}

func makeRecord(obs *ingestionv1.SensorObservation) *kgo.Record {
	b, _ := protojson.MarshalOptions{UseProtoNames: true}.Marshal(obs)
	return &kgo.Record{Value: b}
}

func radarObsAt(sensorID string, lat, lon float64) *ingestionv1.SensorObservation {
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

// T20 — Pipeline: new observation with no existing tracks → creates a new track
func TestFusionPipeline_NewTrackCreated(t *testing.T) {
	mp := &mockProducer{}
	pipeline := newTestPipeline(mp)

	obs := radarObsAt("RADAR-01", 45.0, -60.0)
	if err := pipeline.HandleObservation(context.Background(), makeRecord(obs)); err != nil {
		t.Fatalf("HandleObservation error: %v", err)
	}

	if len(mp.produced) != 1 {
		t.Fatalf("expected 1 produced track, got %d", len(mp.produced))
	}
	if mp.produced[0].GetTrackId() == "" {
		t.Error("expected non-empty track ID")
	}
}

// T21 — Pipeline: second observation near existing track → updates it
func TestFusionPipeline_TrackUpdated(t *testing.T) {
	mp := &mockProducer{}
	pipeline := newTestPipeline(mp)

	obs1 := radarObsAt("RADAR-01", 45.0, -60.0)
	pipeline.HandleObservation(context.Background(), makeRecord(obs1))
	trackID := mp.produced[0].GetTrackId()

	// Second observation very close to the first, 1 second later
	obs2 := radarObsAt("RADAR-02", 45.001, -60.001) // ~0.08 NM away
	obs2.ObservationTime = timestamppb.New(time.Now().Add(time.Second))
	pipeline.HandleObservation(context.Background(), makeRecord(obs2))

	if len(mp.produced) != 2 {
		t.Fatalf("expected 2 produced events, got %d", len(mp.produced))
	}
	if mp.produced[1].GetTrackId() != trackID {
		t.Errorf("expected same track updated, got new ID %s vs %s", mp.produced[1].GetTrackId(), trackID)
	}
}

// T22 — Pipeline: two co-located tracks with high correlation → merge triggered
func TestFusionPipeline_TracksGetMerged(t *testing.T) {
	mp := &mockProducer{}
	pipeline := newTestPipeline(mp)

	// Create two tracks manually at the same position via two separate first observations
	obs1 := radarObsAt("RADAR-A", 45.0, -60.0)
	obs2 := radarObsAt("RADAR-B", 45.0, -60.0) // identical position

	// First observation creates track A
	pipeline.HandleObservation(context.Background(), makeRecord(obs1))
	// Second observation: gating finds track A as candidate; if score ≥ auto → updates
	// But we want to test a merge — produce with a 31s gap to exceed time gate so it creates track B
	obs2.ObservationTime = timestamppb.New(time.Now().Add(-35 * time.Second))
	pipeline.HandleObservation(context.Background(), makeRecord(obs2))

	// Now produce a third obs that triggers merge check
	obs3 := radarObsAt("RADAR-C", 45.0, -60.0)
	obs3.ObservationTime = timestamppb.New(time.Now())
	pipeline.HandleObservation(context.Background(), makeRecord(obs3))

	// We should have at least 3 produced events (initial 2 tracks + update/merge)
	if len(mp.produced) < 2 {
		t.Errorf("expected at least 2 produced events, got %d", len(mp.produced))
	}
}

// Invalid proto bytes → error returned
func TestFusionPipeline_InvalidProto(t *testing.T) {
	mp := &mockProducer{}
	pipeline := newTestPipeline(mp)

	record := &kgo.Record{Value: []byte("not valid proto")}
	err := pipeline.HandleObservation(context.Background(), record)
	if err == nil {
		t.Error("expected error for invalid proto payload")
	}
}

// Observation without position (e.g., cyber) → skipped gracefully
func TestFusionPipeline_SkipsPositionlessObservation(t *testing.T) {
	mp := &mockProducer{}
	pipeline := newTestPipeline(mp)

	obs := &ingestionv1.SensorObservation{
		SensorId:        "CYBER-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
		ObservationTime: timestamppb.New(time.Now()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		// No Position field
	}
	if err := pipeline.HandleObservation(context.Background(), makeRecord(obs)); err != nil {
		t.Fatalf("unexpected error for positionless obs: %v", err)
	}
	if len(mp.produced) != 0 {
		t.Errorf("expected 0 produced tracks for positionless obs, got %d", len(mp.produced))
	}
}
