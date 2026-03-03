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
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/mapper"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestELINTIntegration_FullPipeline tests the full validation -> handler pipeline.
func TestELINTIntegration_FullPipeline(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-elint-ingestion", guard)

h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)

// Test 1: Valid ELINT observation accepted
obs := &ingestionv1.SensorObservation{
SensorId:        "ELINT-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_ElintComint{
ElintComint: &ingestionv1.ELINTDetection{
EmitterId:             "EMIT-INTEGRATION-01",
RadarType:             "FIRE_CONTROL",
FrequencyMhz:          5000.0,
CepMeters:             100.0,
Confidence:            0.9,
ContentClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
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

// Test 2: Invalid ELINT observation rejected
obsInvalid := &ingestionv1.SensorObservation{
SensorId:        "ELINT-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_ElintComint{
ElintComint: &ingestionv1.ELINTDetection{
EmitterId:             "",  // missing emitter_id
RadarType:             "FIRE_CONTROL",
FrequencyMhz:          5000.0,
CepMeters:             100.0,
Confidence:            0.9,
ContentClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
},
},
}

ack2, err2 := h.IngestSingleObservation(context.Background(), obsInvalid)
if err2 != nil {
t.Fatalf("unexpected error: %v", err2)
}
if ack2.GetAccepted() {
t.Error("expected rejected for missing emitter_id")
}

// Test 3: Statistics
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "ELINT-INTEGRATION-01"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.TotalReceived != 2 {
t.Errorf("expected 2 received, got %d", resp.TotalReceived)
}
}
