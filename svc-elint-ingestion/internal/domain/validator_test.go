// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-elint-ingestion/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

func validELINTObs() *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        "ELINT-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_ElintComint{
ElintComint: &ingestionv1.ELINTDetection{
EmitterId:             "EMITTER-01",
RadarType:             "SA-11 FIRE CONTROL",
FrequencyMhz:          9000,
CepMeters:             100,
Confidence:            0.8,
ScanType:              "circular",
ContentClassification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
},
},
}
}

func TestELINTValidator_T01_ValidDetection(t *testing.T) {
v := domain.NewValidator()
result := v.Validate(validELINTObs())
if !result.Valid {
t.Errorf("expected valid, got errors: %+v", result.Errors)
}
}

func TestELINTValidator_T02_CEPZero(t *testing.T) {
obs := validELINTObs()
obs.GetElintComint().CepMeters = 0
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestELINTValidator_T03_InvalidScanType(t *testing.T) {
obs := validELINTObs()
obs.GetElintComint().ScanType = "invalid"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestELINTValidator_T04_ContentClassificationExceeds(t *testing.T) {
obs := validELINTObs()
obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED
obs.GetElintComint().ContentClassification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestELINTValidator_T05_MissingRadarType(t *testing.T) {
obs := validELINTObs()
obs.GetElintComint().RadarType = ""
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
}

func TestELINTValidator_MissingElintPayload(t *testing.T) {
obs := validELINTObs()
obs.SensorData = nil
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid when elint_comint payload is nil")
}
}
