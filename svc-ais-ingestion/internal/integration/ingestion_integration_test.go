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
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ais-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ais-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ais-ingestion/internal/mapper"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestAISIntegration_FullPipeline tests the full validation -> handler pipeline.
func TestAISIntegration_FullPipeline(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-ais-ingestion", guard)

h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger)

// Test 1: Valid AIS observation accepted
obs := &ingestionv1.SensorObservation{
SensorId:        "AIS-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
Position:        &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:           "123456789",
VesselTypeCode: 30,
AisMessageType: 1,
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

// Test 2: Invalid MMSI rejected
obsInvalid := &ingestionv1.SensorObservation{
SensorId:        "AIS-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
Position:        &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:           "INVALID",
VesselTypeCode: 30,
AisMessageType: 1,
},
},
}

ack2, err2 := h.IngestSingleObservation(context.Background(), obsInvalid)
if err2 != nil {
t.Fatalf("unexpected error: %v", err2)
}
if ack2.GetAccepted() {
t.Error("expected rejected for invalid MMSI")
}

// Test 3: BFT with UNCLASSIFIED rejected
obsBFT := &ingestionv1.SensorObservation{
SensorId:        "AIS-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
Position:        &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:           "123456789",
VesselTypeCode: 30,
AisMessageType: 1,
IsBft:          true,
},
},
}
ack3, err3 := h.IngestSingleObservation(context.Background(), obsBFT)
if err3 != nil {
t.Fatalf("unexpected error: %v", err3)
}
if ack3.GetAccepted() {
t.Error("expected rejected: BFT must be >= PROTECTED_B")
}

// Test 4: Statistics
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "AIS-INTEGRATION-01"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.TotalReceived != 3 {
t.Errorf("expected 3 received, got %d", resp.TotalReceived)
}
if resp.TotalAccepted != 1 {
t.Errorf("expected 1 accepted, got %d", resp.TotalAccepted)
}
}
