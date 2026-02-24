// CLASSIFICATION: UNCLASSIFIED
package ingestion_test

import (
"context"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// mockValidator always accepts or rejects based on configuration.
type mockValidator struct {
valid  bool
errors []ingestion.ValidationError
}

func (m *mockValidator) Validate(_ *ingestionv1.SensorObservation) ingestion.ValidationResult {
return ingestion.ValidationResult{Valid: m.valid, Errors: m.errors}
}

// mockNormalizer is a no-op.
type mockNormalizer struct{}

func (m *mockNormalizer) Normalize(_ *ingestionv1.SensorObservation) {}

// mockProducer records produce calls.
type mockProducer struct {
called int
err    error
}

func (m *mockProducer) Produce(_ context.Context, _ *ingestionv1.SensorObservation) error {
m.called++
return m.err
}

func (m *mockProducer) Close() error { return nil }

func newTestObs() *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        "TEST-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.Now(),
}
}

func TestHandler_IngestSingleObservation_Accepted(t *testing.T) {
p := &mockProducer{}
dlq := &mockProducer{}
h := ingestion.NewHandler(&mockValidator{valid: true}, &mockNormalizer{}, p, dlq,
zap.NewNop(), &ingestion.Config{
ServiceName:       "test",
MaxClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
ack, err := h.IngestSingleObservation(context.Background(), newTestObs())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !ack.GetAccepted() {
t.Error("expected accepted")
}
if p.called != 1 {
t.Errorf("expected 1 produce call, got %d", p.called)
}
}

func TestHandler_IngestSingleObservation_Rejected(t *testing.T) {
p := &mockProducer{}
dlq := &mockProducer{}
errs := []ingestion.ValidationError{{Field: "sensor_id", Rule: "required", Message: "sensor_id is required"}}
h := ingestion.NewHandler(&mockValidator{valid: false, errors: errs}, &mockNormalizer{}, p, dlq,
zap.NewNop(), &ingestion.Config{
ServiceName:       "test",
MaxClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
ack, err := h.IngestSingleObservation(context.Background(), newTestObs())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if ack.GetAccepted() {
t.Error("expected rejected")
}
if dlq.called != 1 {
t.Errorf("expected 1 DLQ produce call, got %d", dlq.called)
}
}

func TestHandler_IngestSingleObservation_ClassificationViolation(t *testing.T) {
h := ingestion.NewHandler(&mockValidator{valid: true}, &mockNormalizer{}, &mockProducer{}, &mockProducer{},
zap.NewNop(), &ingestion.Config{
ServiceName:       "test",
MaxClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
})
obs := newTestObs()
obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
_, err := h.IngestSingleObservation(context.Background(), obs)
if err == nil {
t.Error("expected error for classification violation")
}
}

func TestHandler_IngestSingleObservation_NilObservation(t *testing.T) {
h := ingestion.NewHandler(&mockValidator{valid: true}, &mockNormalizer{}, &mockProducer{}, &mockProducer{},
zap.NewNop(), &ingestion.Config{
ServiceName:       "test",
MaxClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
_, err := h.IngestSingleObservation(context.Background(), nil)
if err == nil {
t.Error("expected error for nil observation")
}
}

func TestHandler_GetSensorStatus_Unknown(t *testing.T) {
h := ingestion.NewHandler(&mockValidator{valid: true}, &mockNormalizer{}, &mockProducer{}, &mockProducer{},
zap.NewNop(), &ingestion.Config{
ServiceName:       "test",
MaxClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
resp, err := h.GetSensorStatus(context.Background(), &ingestionv1.GetSensorStatusRequest{SensorId: "UNKNOWN"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.GetConnected() {
t.Error("expected not connected for unknown sensor")
}
}
