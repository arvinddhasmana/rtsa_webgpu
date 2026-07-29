// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"

ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-ais-ingestion/internal/domain"
)

func TestAISNormalizer_TrimsAndUppercases(t *testing.T) {
n := domain.NewNormalizer()
obs := &ingestionv1.SensorObservation{
SensorId: "  AIS-01  ",
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:       " 123456789 ",
VesselName: " my vessel ",
},
},
}
n.Normalize(obs)
if obs.GetSensorId() != "AIS-01" {
t.Errorf("expected 'AIS-01', got %q", obs.GetSensorId())
}
if obs.GetAisBft().GetMmsi() != "123456789" {
t.Errorf("expected '123456789', got %q", obs.GetAisBft().GetMmsi())
}
if obs.GetAisBft().GetVesselName() != "MY VESSEL" {
t.Errorf("expected 'MY VESSEL', got %q", obs.GetAisBft().GetVesselName())
}
}
