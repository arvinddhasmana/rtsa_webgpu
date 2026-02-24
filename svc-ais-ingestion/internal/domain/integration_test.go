// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package domain_test

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ais-ingestion/internal/domain"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestAISIntegration_FullPipeline tests the full validation -> handler pipeline.
func TestAISIntegration_FullPipeline(t *testing.T) {
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
producer := &ingestion.NoopProducer{}
dlqProducer := &ingestion.NoopProducer{}
logger := zap.NewNop()
cfg := &ingestion.Config{
ServiceName:       "svc-ais-ingestion",
OutputTopic:       "sensors.ais.positions",
DLQTopic:          "dlq.sensors.ais",
MaxClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}

handler := ingestion.NewHandler(validator, normalizer, producer, dlqProducer, logger, cfg)

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

ack, err := handler.IngestSingleObservation(context.Background(), obs)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !ack.GetAccepted() {
t.Errorf("expected accepted, got: %s", ack.GetRejectionReason())
}

// Test 2: Invalid MMSI rejected to DLQ
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

ack2, err2 := handler.IngestSingleObservation(context.Background(), obsInvalid)
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
ack3, err3 := handler.IngestSingleObservation(context.Background(), obsBFT)
if err3 != nil {
t.Fatalf("unexpected error: %v", err3)
}
if ack3.GetAccepted() {
t.Error("expected rejected: BFT must be >= PROTECTED_B")
}
}
