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
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/mapper"
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
enricher := mapper.NewEnricher("svc-ew-ingestion", guard)
return handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger)
}

func validObservation(sensorID string) *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
SensorData: &ingestionv1.SensorObservation_EwSigint{
EwSigint: &ingestionv1.EWIntercept{
EmitterId:      "EMIT-001",
FrequencyMhz:   100.0,
BearingDegrees: 45.0,
Confidence:     0.9,
ModulationType: "AM",
},
},
}
}

func TestHandler_ValidObservationAccepted(t *testing.T) {
h := buildHandler()
ack, err := h.IngestSingleObservation(context.Background(), validObservation("EW-001"))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !ack.GetAccepted() {
t.Errorf("expected Accepted=true, got false: %s", ack.GetRejectionReason())
}
if ack.GetObservationId() == "" {
t.Error("expected observation_id to be set")
}
}

func TestHandler_InvalidObservationRejected(t *testing.T) {
h := buildHandler()
obs := validObservation("EW-001")
obs.SensorId = ""

ack, err := h.IngestSingleObservation(context.Background(), obs)
if err != nil {
t.Fatalf("unexpected gRPC error: %v", err)
}
if ack.GetAccepted() {
t.Error("expected Accepted=false for invalid observation")
}
if ack.GetRejectionReason() == "" {
t.Error("expected rejection reason to be set")
}
}

func TestHandler_ClassificationViolation(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
enricher := mapper.NewEnricher("svc-ew-ingestion", guard)
h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger)

obs := validObservation("EW-001")
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
h.IngestSingleObservation(context.Background(), validObservation("EW-001"))
h.IngestSingleObservation(context.Background(), validObservation("EW-001"))

obs := validObservation("EW-001")
obs.SensorId = ""
h.IngestSingleObservation(context.Background(), obs)

resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "EW-001"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.TotalReceived != 3 {
t.Errorf("expected 3 received, got %d", resp.TotalReceived)
}
if resp.TotalAccepted != 2 {
t.Errorf("expected 2 accepted, got %d", resp.TotalAccepted)
}
if resp.TotalRejected != 1 {
t.Errorf("expected 1 rejected, got %d", resp.TotalRejected)
}
}

func TestHandler_ObservationIDAssigned(t *testing.T) {
h := buildHandler()
ack1, _ := h.IngestSingleObservation(context.Background(), validObservation("EW-001"))
ack2, _ := h.IngestSingleObservation(context.Background(), validObservation("EW-001"))

if ack1.GetObservationId() == ack2.GetObservationId() {
t.Error("each observation should get a unique ID")
}
}

func TestHandler_SensorTypeIsEW(t *testing.T) {
h := buildHandler()
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.SensorType != commonv1.SensorType_SENSOR_TYPE_EW_SIGINT {
t.Errorf("expected EW_SIGINT sensor type, got %v", resp.SensorType)
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

func TestHandler_NilDLQProducer_NocrashOnRejection(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-ew-ingestion", guard)
h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger)

obs := validObservation("EW-001")
obs.SensorId = ""

defer func() {
if r := recover(); r != nil {
t.Errorf("should not panic with nil DLQ producer: %v", r)
}
}()

ack, err := h.IngestSingleObservation(context.Background(), obs)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if ack.GetAccepted() {
t.Error("expected rejection")
}
}

// mockStream implements ingestionv1.IngestionService_IngestSensorDataServer
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

validObs := validObservation("EW-001")
invalidObs := validObservation("EW-002")
invalidObs.SensorId = ""

stream := &mockStream{
observations: []*ingestionv1.SensorObservation{validObs, invalidObs, validObs},
ctx:          context.Background(),
}

err := h.IngestSensorData(stream)
if err != nil {
t.Fatalf("unexpected stream error: %v", err)
}

if stream.summary.TotalReceived != 3 {
t.Errorf("expected 3 received, got %d", stream.summary.TotalReceived)
}
if stream.summary.Accepted != 2 {
t.Errorf("expected 2 accepted, got %d", stream.summary.Accepted)
}
if stream.summary.Rejected != 1 {
t.Errorf("expected 1 rejected, got %d", stream.summary.Rejected)
}
}

func TestHandler_ClientStreaming_AllValid(t *testing.T) {
h := buildHandler()

obs := []*ingestionv1.SensorObservation{
validObservation("EW-001"),
validObservation("EW-002"),
validObservation("EW-003"),
}

stream := &mockStream{observations: obs, ctx: context.Background()}
err := h.IngestSensorData(stream)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if stream.summary.Accepted != 3 {
t.Errorf("expected 3 accepted, got %d", stream.summary.Accepted)
}
}

func TestHandler_ClientStreaming_Empty(t *testing.T) {
h := buildHandler()
stream := &mockStream{observations: nil, ctx: context.Background()}
err := h.IngestSensorData(stream)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if stream.summary.TotalReceived != 0 {
t.Errorf("expected 0, got %d", stream.summary.TotalReceived)
}
}
