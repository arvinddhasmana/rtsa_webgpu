// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration_test

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/mapper"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestEWIntegration_FullPipeline tests the full validation -> handler pipeline.
func TestEWIntegration_FullPipeline(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-ew-ingestion", guard)

h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)

// Test 1: Valid EW observation accepted
obs := &ingestionv1.SensorObservation{
SensorId:        "EW-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_EwSigint{
EwSigint: &ingestionv1.EWIntercept{
EmitterId:      "EMIT-INTEGRATION-01",
FrequencyMhz:   100.0,
BearingDegrees: 45.0,
Confidence:     0.9,
ModulationType: "AM",
},
},
}

ack, err := h.IngestSingleObservation(context.Background(), obs)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !ack.GetAccepted() {
t.Errorf("expected accepted, got: %s", ack.GetRejectionReason())
}
if ack.GetObservationId() == "" {
t.Error("expected observation_id to be assigned")
}

// Test 2: Invalid observation rejected (no sensor_id)
obsInvalid := &ingestionv1.SensorObservation{
SensorId:        "",
SensorType:      commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
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

ack2, err2 := h.IngestSingleObservation(context.Background(), obsInvalid)
if err2 != nil {
t.Fatalf("unexpected error: %v", err2)
}
if ack2.GetAccepted() {
t.Error("expected rejected for missing sensor_id")
}

// Test 3: Statistics are tracked correctly
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "EW-INTEGRATION-01"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.TotalReceived != 2 {
t.Errorf("expected 2 received, got %d", resp.TotalReceived)
}
if resp.TotalAccepted != 1 {
t.Errorf("expected 1 accepted, got %d", resp.TotalAccepted)
}
if resp.TotalRejected != 1 {
t.Errorf("expected 1 rejected, got %d", resp.TotalRejected)
}
}
