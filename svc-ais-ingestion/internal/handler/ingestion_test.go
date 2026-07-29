// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/audit"
	"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/handler"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/mapper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func buildHandler() *handler.IngestionHandler {
	logger := zap.NewNop()
	validator := domain.NewValidator()
	normalizer := domain.NewNormalizer()
	guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
	enricher := mapper.NewEnricher("svc-ais-ingestion", guard)
	// audit emitter with nil producer doesn't crash on Emit
	emitter := audit.NewEmitter(nil, "svc-ais-ingestion", logger)
	return handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, emitter, logger, nil)
}

func validObservation(sensorID string) *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		SensorId:        sensorID,
		SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		Position:        &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
		SensorData: &ingestionv1.SensorObservation_AisBft{
			AisBft: &ingestionv1.AISPosition{
				Mmsi:           "123456789",
				VesselTypeCode: 30,
				AisMessageType: 1,
			},
		},
	}
}

func TestHandler_ValidObservationAccepted(t *testing.T) {
	h := buildHandler()
	ack, err := h.IngestSingleObservation(context.Background(), validObservation("AIS-001"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ack.GetAccepted() {
		t.Errorf("expected Accepted=true, got false: %s", ack.GetRejectionReason())
	}
}

func TestHandler_InvalidObservationRejected(t *testing.T) {
	h := buildHandler()
	obs := validObservation("AIS-001")
	obs.SensorId = ""

	ack, err := h.IngestSingleObservation(context.Background(), obs)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}
	if ack.GetAccepted() {
		t.Error("expected Accepted=false for invalid observation")
	}
}

func TestHandler_ClassificationViolation(t *testing.T) {
	logger := zap.NewNop()
	validator := domain.NewValidator()
	normalizer := domain.NewNormalizer()
	guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
	enricher := mapper.NewEnricher("svc-ais-ingestion", guard)
	h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)

	obs := validObservation("AIS-001")
	obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET

	_, err := h.IngestSingleObservation(context.Background(), obs)
	if err == nil {
		t.Fatal("expected PermissionDenied error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PERMISSION_DENIED, got: %v", err)
	}
}

func TestHandler_GetSensorStatus(t *testing.T) {
	h := buildHandler()
	h.IngestSingleObservation(context.Background(), validObservation("AIS-001"))
	h.IngestSingleObservation(context.Background(), validObservation("AIS-001"))

	obs := validObservation("AIS-001")
	obs.SensorId = ""
	h.IngestSingleObservation(context.Background(), obs)

	resp, err := h.GetSensorStatus(context.Background(),
		&ingestionv1.GetSensorStatusRequest{SensorId: "AIS-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalReceived != 3 {
		t.Errorf("expected 3 received, got %d", resp.TotalReceived)
	}
}

func TestHandler_SensorTypeIsAIS(t *testing.T) {
	h := buildHandler()
	resp, err := h.GetSensorStatus(context.Background(),
		&ingestionv1.GetSensorStatusRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SensorType != commonv1.SensorType_SENSOR_TYPE_AIS_BFT {
		t.Errorf("expected AIS_BFT sensor type, got %v", resp.SensorType)
	}
}

func TestHandler_GetSensorStatus_InitialZero(t *testing.T) {
	h := buildHandler()
	resp, err := h.GetSensorStatus(context.Background(),
		&ingestionv1.GetSensorStatusRequest{SensorId: "any"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalReceived != 0 {
		t.Errorf("expected 0 initially, got %d", resp.TotalReceived)
	}
}

type mockStream struct {
	observations []*ingestionv1.SensorObservation
	idx          int
	summary      *ingestionv1.IngestSummary
	ctx          context.Context
	recvErr      error
}

func (m *mockStream) Recv() (*ingestionv1.SensorObservation, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	if m.idx >= len(m.observations) {
		return nil, io.EOF
	}
	obs := m.observations[m.idx]
	m.idx++
	return obs, nil
}

func (m *mockStream) SendAndClose(summary *ingestionv1.IngestSummary) error {
	m.summary = summary
	return nil
}

func (m *mockStream) SetHeader(md metadata.MD) error  { return nil }
func (m *mockStream) SendHeader(md metadata.MD) error  { return nil }
func (m *mockStream) SetTrailer(md metadata.MD)        {}
func (m *mockStream) Context() context.Context         { return m.ctx }
func (m *mockStream) SendMsg(v interface{}) error      { return nil }
func (m *mockStream) RecvMsg(v interface{}) error      { return nil }

func TestHandler_ClientStreaming_MixedObservations(t *testing.T) {
	h := buildHandler()

	validObs := validObservation("AIS-001")
	invalidObs := validObservation("AIS-002")
	invalidObs.SensorId = ""

	stream := &mockStream{
		observations: []*ingestionv1.SensorObservation{validObs, invalidObs, validObs},
		ctx:          context.Background(),
	}

	err := h.IngestSensorData(stream)
	if err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}
	if stream.summary.Accepted != 2 {
		t.Errorf("expected 2 accepted, got %d", stream.summary.Accepted)
	}
}

func TestHandler_IngestSensorData_RecvError(t *testing.T) {
	h := buildHandler()
	stream := &mockStream{
		recvErr: errors.New("recv fail"),
		ctx:     context.Background(),
	}
	err := h.IngestSensorData(stream)
	if err == nil {
		t.Fatal("expected error on Recv failure")
	}
}

func TestHandler_ListSensorStatuses(t *testing.T) {
	h := buildHandler()
	h.IngestSingleObservation(context.Background(), validObservation("AIS-001"))

	t.Run("list all", func(t *testing.T) {
		resp, err := h.ListSensorStatuses(context.Background(), &ingestionv1.ListSensorStatusesRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Sensors) != 1 {
			t.Errorf("expected 1 sensor, got %d", len(resp.Sensors))
		}
	})

	t.Run("active within seconds", func(t *testing.T) {
		resp, err := h.ListSensorStatuses(context.Background(), &ingestionv1.ListSensorStatusesRequest{
			ActiveWithinSeconds: 10,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Sensors) != 1 {
			t.Errorf("expected 1 sensor, got %d", len(resp.Sensors))
		}

		resp, err = h.ListSensorStatuses(context.Background(), &ingestionv1.ListSensorStatusesRequest{
			ActiveWithinSeconds: 1,
		})
		// sleep to trigger inactivity
		time.Sleep(1100 * time.Millisecond)
		resp, err = h.ListSensorStatuses(context.Background(), &ingestionv1.ListSensorStatusesRequest{
			ActiveWithinSeconds: 1,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Sensors) != 0 {
			t.Errorf("expected 0 sensors (inactive), got %d", len(resp.Sensors))
		}
	})
}

// TestGetSensorDiagnostic_AfterIngestCycle verifies GetSensorDiagnostic returns
// correct counters and DLQ breakdown after a mix of valid and invalid observations.
func TestGetSensorDiagnostic_AfterIngestCycle(t *testing.T) {
	h := buildHandler()

	const sensorID = "DIAG-SENSOR-01"

	// Ingest 10 valid observations.
	for i := 0; i < 10; i++ {
		obs := validObservation(sensorID)
		_, err := h.IngestSingleObservation(context.Background(), obs)
		if err != nil {
			t.Fatalf("valid obs failed: %v", err)
		}
	}

	// Ingest 5 invalid observations (missing sensor_id → rejected to DLQ).
	for i := 0; i < 5; i++ {
		obs := validObservation(sensorID)
		obs.Position.Latitude = 999.0 // out-of-range latitude causes validation failure
		_, err := h.IngestSingleObservation(context.Background(), obs)
		if err != nil {
			t.Fatalf("ingest of invalid obs returned gRPC error (expected soft reject): %v", err)
		}
	}

	req := &ingestionv1.GetSensorDiagnosticRequest{
		SensorId:          sensorID,
		HistorySamples:    20,
		RecentEventsLimit: 20,
	}
	resp, err := h.GetSensorDiagnostic(context.Background(), req)
	if err != nil {
		t.Fatalf("GetSensorDiagnostic failed: %v", err)
	}

	if resp.TotalReceived != 15 {
		t.Errorf("expected TotalReceived=15, got %d", resp.TotalReceived)
	}
	if resp.TotalAccepted != 10 {
		t.Errorf("expected TotalAccepted=10, got %d", resp.TotalAccepted)
	}
	if resp.TotalRejected != 5 {
		t.Errorf("expected TotalRejected=5, got %d", resp.TotalRejected)
	}
	if len(resp.DlqBreakdown) == 0 {
		t.Error("expected at least one DLQ reason entry")
	}
	var totalDLQ int64
	for _, entry := range resp.DlqBreakdown {
		totalDLQ += entry.Count
	}
	if totalDLQ != 5 {
		t.Errorf("expected total DLQ count=5, got %d", totalDLQ)
	}
}
