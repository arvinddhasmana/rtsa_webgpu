// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-ew-ingestion/internal/domain"
)

func NewNormalizer() *domain.Normalizer { return domain.NewNormalizer() }

func TestEWNormalizer_TrimsWhitespace(t *testing.T) {
n := domain.NewNormalizer()
obs := &ingestionv1.SensorObservation{
SensorId: "  EW-01  ",
SensorData: &ingestionv1.SensorObservation_EwSigint{
EwSigint: &ingestionv1.EWIntercept{
EmitterId:      "  EMITTER-01  ",
ModulationType: " am ",
},
},
}
n.Normalize(obs)
if obs.GetSensorId() != "EW-01" {
t.Errorf("expected 'EW-01', got %q", obs.GetSensorId())
}
if obs.GetEwSigint().GetEmitterId() != "EMITTER-01" {
t.Errorf("expected 'EMITTER-01', got %q", obs.GetEwSigint().GetEmitterId())
}
if obs.GetEwSigint().GetModulationType() != "AM" {
t.Errorf("expected 'AM', got %q", obs.GetEwSigint().GetModulationType())
}
}
