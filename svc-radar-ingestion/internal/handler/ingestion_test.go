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
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/handler"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/mapper"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func buildHandler() *handler.IngestionHandler {
logger := zap.NewNop()
validator := domain.NewRadarValidator(logger)
normalizer := domain.NewRadarNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-radar-ingestion", guard)
	// nil producers — no actual Redpanda needed for unit tests
	// nil coverage
	return handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)
}

func validObservation(sensorID string) *ingestionv1.SensorObservation {
speed := float64(15.0)
heading := float64(180.0)
return &ingestionv1.SensorObservation{
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:       45.0,
Longitude:      -60.0,
SpeedKnots:     &speed,
HeadingDegrees: &heading,
},
SensorData: &ingestionv1.SensorObservation_Radar{
Radar: &ingestionv1.RadarTrack{
TrackNumber:    "TRK-001",
RangeNm:        25.0,
BearingDegrees: 90.0,
TrackQuality:   0.85,
},
},
}
}

// T01: Valid radar observation accepted
func TestHandler_T01_ValidObservationAccepted(t *testing.T) {
h := buildHandler()
ack, err := h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))
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

// T02: Invalid observation rejected
func TestHandler_T02_InvalidObservationRejected(t *testing.T) {
h := buildHandler()
obs := validObservation("RADAR-001")
obs.SensorId = "" // invalid

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

// T04: Classification violation
func TestHandler_T04_ClassificationViolation(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewRadarValidator(logger)
normalizer := domain.NewRadarNormalizer()
// Set ceiling to UNCLASSIFIED only
	guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
	enricher := mapper.NewEnricher("svc-radar-ingestion", guard)
	h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)

obs := validObservation("RADAR-001")
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

// T06: GetSensorStatus returns stats
func TestHandler_T06_GetSensorStatus(t *testing.T) {
h := buildHandler()
// Process a few observations
h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))
h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))

obs := validObservation("RADAR-001")
obs.SensorId = "" // reject
h.IngestSingleObservation(context.Background(), obs)

resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "RADAR-001"})
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

// T01 observation_id uniqueness
func TestHandler_ObservationIDAssigned(t *testing.T) {
h := buildHandler()
ack1, _ := h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))
ack2, _ := h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))

if ack1.GetObservationId() == ack2.GetObservationId() {
t.Error("each observation should get a unique ID")
}
}

// TestDLQProducer with nil doesn't panic
func TestHandler_NilDLQProducer_NocrashOnRejection(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewRadarValidator(logger)
normalizer := domain.NewRadarNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-radar-ingestion", guard)

// nil DLQ producer
	h := handler.NewIngestionHandler(validator, normalizer, enricher,
		nil, // main producer nil
		nil, // dlq nil
		nil, logger, nil)

obs := validObservation("RADAR-001")
obs.SensorId = "" // force rejection

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

// TestProducerNilNoError — with nil producer, accepted obs returns accepted=true
// (the nil producer is checked before produce)
func TestHandler_NilMainProducer_ValidObs(t *testing.T) {
h := buildHandler()
ack, err := h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !ack.GetAccepted() {
t.Errorf("expected accepted=true, got: %s", ack.GetRejectionReason())
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

func TestHandler_LastObservationTimeSetAfterAccept(t *testing.T) {
h := buildHandler()
h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))

resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.LastObservationTime == nil {
t.Error("expected LastObservationTime to be set after accepted observation")
}
}

func TestHandler_SensorTypeIsRadar(t *testing.T) {
h := buildHandler()
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.SensorType != commonv1.SensorType_SENSOR_TYPE_RADAR {
t.Errorf("expected RADAR sensor type, got %v", resp.SensorType)
}
}

func TestHandler_MultipleValidObservations(t *testing.T) {
h := buildHandler()
for i := 0; i < 5; i++ {
ack, err := h.IngestSingleObservation(context.Background(), validObservation("RADAR-001"))
if err != nil {
t.Fatalf("unexpected error on iteration %d: %v", i, err)
}
if !ack.Accepted {
t.Errorf("expected accepted on iteration %d", i)
}
}
resp, _ := h.GetSensorStatus(context.Background(), &ingestionv1.GetSensorStatusRequest{})
if resp.TotalAccepted != 5 {
t.Errorf("expected 5 accepted, got %d", resp.TotalAccepted)
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

// T03: Client streaming with mix of valid/invalid
func TestHandler_T03_ClientStreaming_MixedObservations(t *testing.T) {
h := buildHandler()

validObs := validObservation("RADAR-001")
invalidObs := validObservation("RADAR-002")
invalidObs.SensorId = "" // make it invalid

stream := &mockStream{
observations: []*ingestionv1.SensorObservation{validObs, invalidObs, validObs},
ctx:          context.Background(),
}

err := h.IngestSensorData(stream)
if err != nil {
t.Fatalf("unexpected stream error: %v", err)
}

if stream.summary == nil {
t.Fatal("expected non-nil summary")
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

func TestHandler_T03_ClientStreaming_AllValid(t *testing.T) {
h := buildHandler()

obs := []*ingestionv1.SensorObservation{
validObservation("RADAR-001"),
validObservation("RADAR-002"),
validObservation("RADAR-003"),
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

func TestHandler_T03_EmptyStream(t *testing.T) {
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
