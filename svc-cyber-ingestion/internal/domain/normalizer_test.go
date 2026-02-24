// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/domain"
)

func TestCyberNormalizer_NormalizesFields(t *testing.T) {
n := domain.NewNormalizer()
obs := &ingestionv1.SensorObservation{
SensorId: "  CYBER-01  ",
SensorData: &ingestionv1.SensorObservation_Cyber{
Cyber: &ingestionv1.CyberIOC{
IocType:    " IPV4-ADDR ",
IocValue:   " 192.168.1.1 ",
SourceFeed: "  CTI  ",
},
},
}
n.Normalize(obs)
if obs.GetSensorId() != "CYBER-01" {
t.Errorf("expected 'CYBER-01', got %q", obs.GetSensorId())
}
if obs.GetCyber().GetIocType() != "ipv4-addr" {
t.Errorf("expected 'ipv4-addr', got %q", obs.GetCyber().GetIocType())
}
if obs.GetCyber().GetIocValue() != "192.168.1.1" {
t.Errorf("expected '192.168.1.1', got %q", obs.GetCyber().GetIocValue())
}
}
