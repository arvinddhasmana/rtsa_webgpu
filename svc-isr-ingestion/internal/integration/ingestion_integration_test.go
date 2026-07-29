// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package integration_test

import (
"context"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
"github.com/arvinddhasmana/rtsa_webgpu/svc-isr-ingestion/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-isr-ingestion/internal/handler"
"github.com/arvinddhasmana/rtsa_webgpu/svc-isr-ingestion/internal/mapper"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestISRIntegration_FullPipeline tests the full validation -> handler pipeline.
func TestISRIntegration_FullPipeline(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-isr-ingestion", guard)

h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)

// Test 1: Valid ISR observation accepted
obs := &ingestionv1.SensorObservation{
SensorId:        "ISR-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_ISR,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Isr{
Isr: &ingestionv1.ISRObservation{
PlatformId: "PLATFORM-INTEGRATION-01",
SensorName: "EO",
ImageId:    "IMG-INTEGRATION-001",
CoveragePolygon: []*commonv1.Position{
{Latitude: 44.0, Longitude: -63.0},
{Latitude: 45.0, Longitude: -63.0},
{Latitude: 45.0, Longitude: -62.0},
},
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

// Test 2: ISR with insufficient polygon vertices rejected
obsInvalid := &ingestionv1.SensorObservation{
SensorId:        "ISR-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_ISR,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Isr{
Isr: &ingestionv1.ISRObservation{
PlatformId: "PLATFORM-INTEGRATION-01",
SensorName: "EO",
ImageId:    "IMG-INTEGRATION-002",
CoveragePolygon: []*commonv1.Position{
{Latitude: 44.0, Longitude: -63.0},
{Latitude: 45.0, Longitude: -63.0},
// only 2 vertices — need at least 3
},
},
},
}

ack2, err2 := h.IngestSingleObservation(context.Background(), obsInvalid)
if err2 != nil {
t.Fatalf("unexpected error: %v", err2)
}
if ack2.GetAccepted() {
t.Error("expected rejected for insufficient polygon vertices")
}

// Test 3: Statistics
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "ISR-INTEGRATION-01"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.TotalReceived != 2 {
t.Errorf("expected 2 received, got %d", resp.TotalReceived)
}
}
