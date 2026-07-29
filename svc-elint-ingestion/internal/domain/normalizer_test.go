// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"

ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-elint-ingestion/internal/domain"
)

func TestELINTNormalizer_TrimsWhitespace(t *testing.T) {
n := domain.NewNormalizer()
obs := &ingestionv1.SensorObservation{
SensorId: "  ELINT-01  ",
SensorData: &ingestionv1.SensorObservation_ElintComint{
ElintComint: &ingestionv1.ELINTDetection{
EmitterId: "  EMITTER-01  ",
ScanType:  "  CIRCULAR  ",
},
},
}
n.Normalize(obs)
if obs.GetSensorId() != "ELINT-01" {
t.Errorf("expected 'ELINT-01', got %q", obs.GetSensorId())
}
if obs.GetElintComint().GetScanType() != "circular" {
t.Errorf("expected 'circular', got %q", obs.GetElintComint().GetScanType())
}
}
