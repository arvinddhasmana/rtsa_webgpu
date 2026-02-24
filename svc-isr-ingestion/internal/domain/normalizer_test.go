// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-isr-ingestion/internal/domain"
)

func TestISRNormalizer_TrimsWhitespace(t *testing.T) {
n := domain.NewNormalizer()
obs := &ingestionv1.SensorObservation{
SensorId: "  ISR-01  ",
SensorData: &ingestionv1.SensorObservation_Isr{
Isr: &ingestionv1.ISRObservation{
PlatformId: "  PLATFORM-01  ",
SensorName: " eo ",
},
},
}
n.Normalize(obs)
if obs.GetSensorId() != "ISR-01" {
t.Errorf("expected 'ISR-01', got %q", obs.GetSensorId())
}
if obs.GetIsr().GetSensorName() != "EO" {
t.Errorf("expected 'EO', got %q", obs.GetIsr().GetSensorName())
}
}
