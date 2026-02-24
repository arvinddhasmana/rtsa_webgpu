// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

func validEWObs() *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        "EW-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_EwSigint{
EwSigint: &ingestionv1.EWIntercept{
EmitterId:      "EMITTER-01",
FrequencyMhz:   100.0,
BearingDegrees: 45.0,
Confidence:     0.9,
ModulationType: "AM",
},
},
}
}

func ptr[T any](v T) *T { return &v }

func TestEWValidator_T01_ValidIntercept(t *testing.T) {
v := domain.NewValidator()
result := v.Validate(validEWObs())
if !result.Valid {
t.Errorf("expected valid, got errors: %+v", result.Errors)
}
}

func TestEWValidator_T02_FrequencyBelowMin(t *testing.T) {
obs := validEWObs()
obs.GetEwSigint().FrequencyMhz = 0.1
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestEWValidator_T03_FrequencyAboveMax(t *testing.T) {
obs := validEWObs()
obs.GetEwSigint().FrequencyMhz = 50000
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestEWValidator_T04_MissingEmitterID(t *testing.T) {
obs := validEWObs()
obs.GetEwSigint().EmitterId = ""
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestEWValidator_T05_BearingOutOfRange(t *testing.T) {
obs := validEWObs()
obs.GetEwSigint().BearingDegrees = 361.0
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestEWValidator_T06_ConfidenceOutOfRange(t *testing.T) {
obs := validEWObs()
obs.GetEwSigint().Confidence = 1.5
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestEWValidator_WrongSensorType(t *testing.T) {
obs := validEWObs()
obs.SensorType = commonv1.SensorType_SENSOR_TYPE_RADAR
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestEWValidator_SuspectPowerDbm(t *testing.T) {
obs := validEWObs()
obs.GetEwSigint().PowerDbm = ptr(-250.0)
v := domain.NewValidator()
result := v.Validate(obs)
if !result.Valid {
t.Error("expected valid (suspect, not rejected)")
}
if !result.Suspect {
t.Error("expected suspect flag")
}
}

func TestEWValidator_MissingEWSigint(t *testing.T) {
obs := validEWObs()
obs.SensorData = nil
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid when ew_sigint payload is nil")
}
}
