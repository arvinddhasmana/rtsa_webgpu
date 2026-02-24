// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-isr-ingestion/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

func validISRObs() *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        "ISR-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_ISR,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Isr{
Isr: &ingestionv1.ISRObservation{
PlatformId: "PLATFORM-01",
SensorName: "EO",
ImageId:    "IMG-001",
CoveragePolygon: []*commonv1.Position{
{Latitude: 44.0, Longitude: -63.0},
{Latitude: 45.0, Longitude: -63.0},
{Latitude: 45.0, Longitude: -64.0},
},
Detections: []*ingestionv1.ISRDetection{
{Confidence: 0.9},
{Confidence: 0.7},
},
},
},
}
}

func TestISRValidator_T01_ValidObservation(t *testing.T) {
v := domain.NewValidator()
result := v.Validate(validISRObs())
if !result.Valid {
t.Errorf("expected valid, got errors: %+v", result.Errors)
}
}

func TestISRValidator_T02_PolygonLessThan3Vertices(t *testing.T) {
obs := validISRObs()
obs.GetIsr().CoveragePolygon = []*commonv1.Position{
{Latitude: 44.0, Longitude: -63.0},
}
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestISRValidator_T03_InvalidSensorName(t *testing.T) {
obs := validISRObs()
obs.GetIsr().SensorName = "LIDAR"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestISRValidator_T04_DetectionConfidenceAbove1(t *testing.T) {
obs := validISRObs()
obs.GetIsr().Detections[0].Confidence = 1.5
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestISRValidator_T05_MissingPlatformID(t *testing.T) {
obs := validISRObs()
obs.GetIsr().PlatformId = ""
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestISRValidator_MissingISRPayload(t *testing.T) {
obs := validISRObs()
obs.SensorData = nil
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid when isr payload is nil")
}
}
