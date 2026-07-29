// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-fusion-engine/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createActiveTrack(tm *domain.TrackManager, sensorID string) *domain.TrackState {
	obs := &ingestionv1.SensorObservation{
		SensorId:        sensorID,
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(time.Now()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		Position: &commonv1.Position{
			Latitude:  45.0,
			Longitude: -60.0,
		},
	}
	t, _ := tm.CreateTrack(obs)
	return t
}

// T18 — StaleMonitor marks ACTIVE track as STALE after timeout
func TestStaleMonitor_MarksStale(t *testing.T) {
	tm := domain.NewTrackManager(domain.NewKalmanFilter())
	track := createActiveTrack(tm, "R1")

	// Back-date UpdatedAt so track appears old
	track.UpdatedAt = time.Now().Add(-2 * time.Second)

	var gotOld, gotNew commonv1.TrackStatus
	sm := domain.NewStaleMonitor(tm,
		1*time.Second,  // staleTimeout
		10*time.Second, // dropTimeout
		50*time.Millisecond, // checkInterval
		func(tr *domain.TrackState, old, newS commonv1.TrackStatus) {
			gotOld = old
			gotNew = newS
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go sm.Start(ctx)
	<-ctx.Done()
	time.Sleep(10 * time.Millisecond) // let final check complete

	got, _ := tm.GetTrack(track.TrackID)
	if got.Status != commonv1.TrackStatus_TRACK_STATUS_STALE {
		t.Errorf("expected STALE, got %v (old=%v new=%v)", got.Status, gotOld, gotNew)
	}
}

// T19 — StaleMonitor drops STALE track after drop timeout
func TestStaleMonitor_DropsStaleTrack(t *testing.T) {
	tm := domain.NewTrackManager(domain.NewKalmanFilter())
	track := createActiveTrack(tm, "R1")

	// Pre-mark as stale and back-date
	tm.MarkStale(track.TrackID)
	track.UpdatedAt = time.Now().Add(-5 * time.Second)

	sm := domain.NewStaleMonitor(tm,
		1*time.Second,       // staleTimeout
		2*time.Second,       // dropTimeout
		50*time.Millisecond, // checkInterval
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go sm.Start(ctx)
	<-ctx.Done()
	time.Sleep(10 * time.Millisecond)

	got, _ := tm.GetTrack(track.TrackID)
	if got.Status != commonv1.TrackStatus_TRACK_STATUS_DROPPED {
		t.Errorf("expected DROPPED, got %v", got.Status)
	}
}

// Context cancellation stops the monitor
func TestStaleMonitor_StopsOnContextCancellation(t *testing.T) {
	tm := domain.NewTrackManager(domain.NewKalmanFilter())
	sm := domain.NewStaleMonitor(tm, time.Minute, time.Minute, 10*time.Millisecond, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sm.Start(ctx)
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Error("StaleMonitor did not stop after context cancellation")
	}
}
