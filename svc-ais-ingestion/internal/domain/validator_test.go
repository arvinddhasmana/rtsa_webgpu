// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ais-ingestion/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

func validAISObs() *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        "AIS-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
Position: &commonv1.Position{
Latitude:  44.65,
Longitude: -63.57,
},
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:           "123456789",
VesselTypeCode: 30,
AisMessageType: 1,
},
},
}
}

func TestAISValidator_T01_ValidPositionReport(t *testing.T) {
v := domain.NewValidator()
result := v.Validate(validAISObs())
if !result.Valid {
t.Errorf("expected valid, got errors: %+v", result.Errors)
}
}

func TestAISValidator_T02_MMSINotNineDigits(t *testing.T) {
obs := validAISObs()
obs.GetAisBft().Mmsi = "12345"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestAISValidator_T03_VesselTypeCodeAbove99(t *testing.T) {
obs := validAISObs()
obs.GetAisBft().VesselTypeCode = 100
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestAISValidator_T04_InvalidAISMessageType(t *testing.T) {
obs := validAISObs()
obs.GetAisBft().AisMessageType = 99
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestAISValidator_T05_BFTWithUnclassified(t *testing.T) {
obs := validAISObs()
obs.GetAisBft().IsBft = true
obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid: BFT must be >= PROTECTED_B")
}
}

func TestAISValidator_T06_MissingPosition(t *testing.T) {
obs := validAISObs()
obs.Position = nil
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid: position required")
}
}

func TestAISValidator_T07_SpeedJump(t *testing.T) {
v := domain.NewValidator()
// First report at Halifax
obs1 := validAISObs()
obs1.GetAisBft().Mmsi = "999888777"
obs1.Position = &commonv1.Position{Latitude: 44.0, Longitude: -63.0}
result1 := v.Validate(obs1)
if !result1.Valid {
t.Fatalf("first report should be valid, got: %+v", result1.Errors)
}

// Second report with huge position jump (~660 NM) 2 minutes later
obs2 := validAISObs()
obs2.GetAisBft().Mmsi = "999888777"
obs2.Position = &commonv1.Position{Latitude: 55.0, Longitude: -63.0}
obs2.ObservationTime = timestamppb.New(time.Now().Add(2 * time.Minute))
result2 := v.Validate(obs2)
if !result2.Valid {
t.Fatalf("second report should be valid but suspect, got: %+v", result2.Errors)
}
if !result2.Suspect {
t.Error("expected suspect flag for position jump")
}
}

func TestAISValidator_MissingAISPayload(t *testing.T) {
obs := validAISObs()
obs.SensorData = nil
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid when ais_bft payload is nil")
}
}
