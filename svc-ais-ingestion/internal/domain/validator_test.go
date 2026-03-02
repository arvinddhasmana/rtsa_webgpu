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

func TestAISValidator_EdgeCases(t *testing.T) {
	v := domain.NewValidator()
	baseObs := func() *ingestionv1.SensorObservation {
		return &ingestionv1.SensorObservation{
			SensorId:        "AIS-001",
			SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
			ObservationTime: timestamppb.New(time.Now()),
			Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
			Position:        &commonv1.Position{Latitude: 45.0, Longitude: -60.0},
			SensorData: &ingestionv1.SensorObservation_AisBft{
				AisBft: &ingestionv1.AISPosition{Mmsi: "123456789", VesselTypeCode: 30, AisMessageType: 1},
			},
		}
	}

	t.Run("long sensor id", func(t *testing.T) {
		obs := baseObs()
		obs.SensorId = string(make([]byte, 130))
		res := v.Validate(obs)
		if res.Valid { t.Error("expected invalid for long sensor id") }
	})

	t.Run("future time", func(t *testing.T) {
		obs := baseObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(10 * time.Minute))
		res := v.Validate(obs)
		if res.Valid { t.Error("expected invalid for future time") }
	})

	t.Run("old time", func(t *testing.T) {
		obs := baseObs()
		obs.ObservationTime = timestamppb.New(time.Now().Add(-48 * time.Hour))
		res := v.Validate(obs)
		if res.Valid { t.Error("expected invalid for old time") }
	})

	t.Run("invalid lat lon", func(t *testing.T) {
		obs := baseObs()
		obs.Position.Latitude = 100
		res := v.Validate(obs)
		if res.Valid { t.Error("expected invalid for lat 100") }
		obs = baseObs()
		obs.Position.Longitude = 200
		res = v.Validate(obs)
		if res.Valid { t.Error("expected invalid for lon 200") }
	})

	t.Run("bft classification too low", func(t *testing.T) {
		obs := baseObs()
		obs.GetAisBft().IsBft = true
		obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
		res := v.Validate(obs)
		if res.Valid { t.Error("expected invalid for BFT with UNCLASSIFIED") }
	})

	t.Run("spoofing jump", func(t *testing.T) {
		v2 := domain.NewValidator()
		obs1 := baseObs()
		v2.Validate(obs1)
		obs2 := baseObs()
		obs2.Position.Latitude = 48.0 // ~180 NM jump
		res := v2.Validate(obs2)
		if !res.Suspect { t.Error("expected suspect for big jump") }
	})
}
