// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration_test validates the full radar ingestion handler pipeline:
// validation → normalisation → enrichment → produce/DLQ routing.
// These tests exercise two or more components interacting, without a live Redpanda
// or ClickHouse container (producer/DLQ are stubbed via nil).
package integration_test

import (
	"context"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/handler"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/mapper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Mid-Atlantic test area (non-sensitive synthetic data per §5.1 of testing_strategy.md).
const (
	testLatMidAtlantic = 44.65
	testLonMidAtlantic = -63.57
)

// newTestHandler assembles a radar IngestionHandler with the given classification ceiling.
// Producer and DLQ producer are nil (no Redpanda required); audit emitter is nil.
func newTestHandler(t *testing.T, ceiling commonv1.ClassificationLevel) *handler.IngestionHandler {
	t.Helper()
	logger := zap.NewNop()
	validator := domain.NewRadarValidator(logger)
	normalizer := domain.NewRadarNormalizer()
	guard := classification.NewGuard(ceiling)
	enricher := mapper.NewEnricher("svc-radar-ingestion", guard)
	return handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)
}

// TestRadarIngestionPipeline_ValidObservation_ReturnsAcceptedAck validates that a
// well-formed radar observation is accepted by the full validation→normalisation→enrichment chain.
func TestRadarIngestionPipeline_ValidObservation_ReturnsAcceptedAck(t *testing.T) {
	h := newTestHandler(t, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)

	obs := &ingestionv1.SensorObservation{
		SensorId:        "RADAR-INT-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Position: &commonv1.Position{
			Latitude:  testLatMidAtlantic,
			Longitude: testLonMidAtlantic,
		},
		SensorData: &ingestionv1.SensorObservation_Radar{
			Radar: &ingestionv1.RadarTrack{
				TrackNumber:    "rdr-001",
				RangeNm:        5.4,
				BearingDegrees: 45.0,
			},
		},
	}

	ack, err := h.IngestSingleObservation(context.Background(), obs)
	if err != nil {
		t.Fatalf("TestRadarIngestionPipeline_ValidObservation_ReturnsAcceptedAck: unexpected gRPC error: %v", err)
	}
	if !ack.GetAccepted() {
		t.Errorf("TestRadarIngestionPipeline_ValidObservation_ReturnsAcceptedAck: expected accepted, got rejection: %s",
			ack.GetRejectionReason())
	}
	if ack.GetObservationId() == "" {
		t.Error("TestRadarIngestionPipeline_ValidObservation_ReturnsAcceptedAck: observation_id must be populated after enrichment")
	}
}

// TestRadarIngestionPipeline_InvalidLatitude_ReturnsRejectedAck validates that an
// observation with an out-of-range latitude (>90) is rejected without a gRPC error.
func TestRadarIngestionPipeline_InvalidLatitude_ReturnsRejectedAck(t *testing.T) {
	h := newTestHandler(t, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)

	obs := &ingestionv1.SensorObservation{
		SensorId:        "RADAR-INT-02",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Position: &commonv1.Position{
			Latitude:  999.0, // invalid — latitude must be in -90..90
			Longitude: testLonMidAtlantic,
		},
	}

	ack, err := h.IngestSingleObservation(context.Background(), obs)
	if err != nil {
		t.Fatalf("TestRadarIngestionPipeline_InvalidLatitude_ReturnsRejectedAck: unexpected gRPC error: %v", err)
	}
	if ack.GetAccepted() {
		t.Error("TestRadarIngestionPipeline_InvalidLatitude_ReturnsRejectedAck: expected rejected for invalid latitude 999.0")
	}
	if ack.GetRejectionReason() == "" {
		t.Error("TestRadarIngestionPipeline_InvalidLatitude_ReturnsRejectedAck: rejection_reason must be non-empty")
	}
}

// TestRadarIngestionPipeline_ClassificationCeilingViolation_ReturnsPermissionDenied validates
// that an observation whose classification exceeds the service ceiling is rejected with
// gRPC PermissionDenied (not an ACK).
func TestRadarIngestionPipeline_ClassificationCeilingViolation_ReturnsPermissionDenied(t *testing.T) {
	// Handler ceiling is UNCLASSIFIED; SECRET observation should be denied.
	h := newTestHandler(t, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)

	obs := &ingestionv1.SensorObservation{
		SensorId:        "RADAR-INT-03",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Position: &commonv1.Position{
			Latitude:  testLatMidAtlantic,
			Longitude: testLonMidAtlantic,
		},
		SensorData: &ingestionv1.SensorObservation_Radar{
			Radar: &ingestionv1.RadarTrack{
				TrackNumber:    "rdr-003",
				RangeNm:        1.0,
				BearingDegrees: 90.0,
			},
		},
	}

	_, err := h.IngestSingleObservation(context.Background(), obs)
	if err == nil {
		t.Fatal("TestRadarIngestionPipeline_ClassificationCeilingViolation_ReturnsPermissionDenied: expected gRPC error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("TestRadarIngestionPipeline_ClassificationCeilingViolation_ReturnsPermissionDenied: error is not a gRPC status: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("TestRadarIngestionPipeline_ClassificationCeilingViolation_ReturnsPermissionDenied: code=%v, want PermissionDenied", st.Code())
	}
}

// TestRadarIngestionPipeline_Statistics_CountsAcceptedAndRejected validates that
// GetSensorStatus correctly reports counters after accept/reject cycles.
func TestRadarIngestionPipeline_Statistics_CountsAcceptedAndRejected(t *testing.T) {
	h := newTestHandler(t, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
	ctx := context.Background()

	// Two valid observations.
	for i := 0; i < 2; i++ {
		_, _ = h.IngestSingleObservation(ctx, &ingestionv1.SensorObservation{
			SensorId:        "RADAR-INT-STATS",
			SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			ObservationTime: timestamppb.New(time.Now().UTC()),
			Position:        &commonv1.Position{Latitude: testLatMidAtlantic, Longitude: testLonMidAtlantic},
			SensorData: &ingestionv1.SensorObservation_Radar{
				Radar: &ingestionv1.RadarTrack{TrackNumber: "rdr-stats", RangeNm: 2.0, BearingDegrees: 0.0},
			},
		})
	}

	// One invalid observation.
	_, _ = h.IngestSingleObservation(ctx, &ingestionv1.SensorObservation{
		SensorId:        "RADAR-INT-STATS",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Position:        &commonv1.Position{Latitude: 999.0, Longitude: testLonMidAtlantic}, // invalid
	})

	resp, err := h.GetSensorStatus(ctx, &ingestionv1.GetSensorStatusRequest{SensorId: "RADAR-INT-STATS"})
	if err != nil {
		t.Fatalf("TestRadarIngestionPipeline_Statistics_CountsAcceptedAndRejected: GetSensorStatus error: %v", err)
	}
	if resp.GetTotalReceived() != 3 {
		t.Errorf("TestRadarIngestionPipeline_Statistics_CountsAcceptedAndRejected: total_received=%d, want 3", resp.GetTotalReceived())
	}
	if resp.GetTotalAccepted() != 2 {
		t.Errorf("TestRadarIngestionPipeline_Statistics_CountsAcceptedAndRejected: total_accepted=%d, want 2", resp.GetTotalAccepted())
	}
	if resp.GetTotalRejected() != 1 {
		t.Errorf("TestRadarIngestionPipeline_Statistics_CountsAcceptedAndRejected: total_rejected=%d, want 1", resp.GetTotalRejected())
	}
}
