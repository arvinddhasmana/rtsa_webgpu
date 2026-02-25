// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
"context"
"io"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/mapper"
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
enricher := mapper.NewEnricher("svc-elint-ingestion", guard)
return handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger)
}

func validObservation(sensorID string) *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
SensorData: &ingestionv1.SensorObservation_ElintComint{
ElintComint: &ingestionv1.ELINTDetection{
EmitterId:              "EMIT-001",
RadarType:              "FIRE_CONTROL",
FrequencyMhz:           5000.0,
CepMeters:              100.0,
Confidence:             0.9,
ContentClassification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
},
},
}
}

func TestHandler_ValidObservationAccepted(t *testing.T) {
h := buildHandler()
ack, err := h.IngestSingleObservation(context.Background(), validObservation("ELINT-001"))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !ack.GetAccepted() {
t.Errorf("expected Accepted=true, got false: %s", ack.GetRejectionReason())
}
}

func TestHandler_InvalidObservationRejected(t *testing.T) {
h := buildHandler()
obs := validObservation("ELINT-001")
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
enricher := mapper.NewEnricher("svc-elint-ingestion", guard)
h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger)

obs := validObservation("ELINT-001")
obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
// Must also set content_classification to match to avoid content_classification validation error
obs.GetElintComint().ContentClassification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET

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
h.IngestSingleObservation(context.Background(), validObservation("ELINT-001"))
h.IngestSingleObservation(context.Background(), validObservation("ELINT-001"))

obs := validObservation("ELINT-001")
obs.SensorId = ""
h.IngestSingleObservation(context.Background(), obs)

resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "ELINT-001"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.TotalReceived != 3 {
t.Errorf("expected 3 received, got %d", resp.TotalReceived)
}
}

func TestHandler_SensorTypeIsELINT(t *testing.T) {
h := buildHandler()
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.SensorType != commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT {
t.Errorf("expected ELINT_COMINT sensor type, got %v", resp.SensorType)
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
if resp.LastObservationTime != nil {
t.Error("expected nil LastObservationTime initially")
}
}

type mockStream struct {
observations []*ingestionv1.SensorObservation
idx          int
summary      *ingestionv1.IngestSummary
ctx          context.Context
}

func (m *mockStream) Recv() (*ingestionv1.SensorObservation, error) {
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

validObs := validObservation("ELINT-001")
invalidObs := validObservation("ELINT-002")
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
