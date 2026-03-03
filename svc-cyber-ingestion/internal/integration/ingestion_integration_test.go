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
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/mapper"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestCyberIntegration_FullPipeline tests the full validation -> handler pipeline.
func TestCyberIntegration_FullPipeline(t *testing.T) {
logger := zap.NewNop()
validator := domain.NewValidator()
normalizer := domain.NewNormalizer()
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
enricher := mapper.NewEnricher("svc-cyber-ingestion", guard)

h := handler.NewIngestionHandler(validator, normalizer, enricher, nil, nil, nil, logger, nil)

// Test 1: Valid Cyber IOC observation accepted
obs := &ingestionv1.SensorObservation{
SensorId:        "CYBER-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Cyber{
Cyber: &ingestionv1.CyberIOC{
StixId:     "indicator--12345678-1234-1234-1234-123456789012",
IocType:    "ipv4-addr",
IocValue:   "10.0.0.1",
Confidence: 0.9,
ValidFrom:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
SourceFeed: "integration-test-feed",
DedupHash:  "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
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

// Test 2: Duplicate IOC rejected
obsDup := &ingestionv1.SensorObservation{
SensorId:        "CYBER-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Cyber{
Cyber: &ingestionv1.CyberIOC{
StixId:     "indicator--12345678-1234-1234-1234-123456789012",
IocType:    "ipv4-addr",
IocValue:   "10.0.0.1",
Confidence: 0.9,
ValidFrom:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
SourceFeed: "integration-test-feed",
DedupHash:  "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", // same hash
},
},
}

ack2, err2 := h.IngestSingleObservation(context.Background(), obsDup)
if err2 != nil {
t.Fatalf("unexpected error: %v", err2)
}
if ack2.GetAccepted() {
t.Error("expected rejected for duplicate IOC")
}

// Test 3: Invalid STIX ID rejected
obsInvalidStix := &ingestionv1.SensorObservation{
SensorId:        "CYBER-INTEGRATION-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Cyber{
Cyber: &ingestionv1.CyberIOC{
StixId:     "invalid-stix-id",
IocType:    "ipv4-addr",
IocValue:   "10.0.0.2",
Confidence: 0.9,
ValidFrom:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
SourceFeed: "integration-test-feed",
DedupHash:  "b1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
},
},
}

ack3, err3 := h.IngestSingleObservation(context.Background(), obsInvalidStix)
if err3 != nil {
t.Fatalf("unexpected error: %v", err3)
}
if ack3.GetAccepted() {
t.Error("expected rejected for invalid STIX ID")
}

// Test 4: Statistics
resp, err := h.GetSensorStatus(context.Background(),
&ingestionv1.GetSensorStatusRequest{SensorId: "CYBER-INTEGRATION-01"})
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
