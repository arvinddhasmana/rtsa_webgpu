// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-radar-ingestion/internal/domain"
)

func TestNormalizer_TrimsWhitespace(t *testing.T) {
n := domain.NewRadarNormalizer()
obs := validObs()
obs.SensorId = "  RADAR-001  "
obs.GetRadar().TrackNumber = " TRK-001 "

result := n.Normalize(obs)
if result.SensorId != "RADAR-001" {
t.Errorf("expected trimmed SensorId, got: %q", result.SensorId)
}
if result.GetRadar().TrackNumber != "TRK-001" {
t.Errorf("expected trimmed TrackNumber, got: %q", result.GetRadar().TrackNumber)
}
}

func TestNormalizer_NormalizesHeading(t *testing.T) {
n := domain.NewRadarNormalizer()
obs := validObs()
heading := float64(370.0) // should be normalized to 10.0
obs.Position.HeadingDegrees = &heading

result := n.Normalize(obs)
got := result.Position.GetHeadingDegrees()
if got < 0 || got >= 360 {
t.Errorf("expected heading in [0,360), got: %.2f", got)
}
}

func TestNormalizer_PrecisionRounding(t *testing.T) {
n := domain.NewRadarNormalizer()
obs := validObs()
obs.Position.Latitude = 45.1234567890
obs.Position.Longitude = -60.9876543210

result := n.Normalize(obs)
// Should be rounded to 6 decimal places
if result.Position.Latitude != 45.123457 {
t.Errorf("expected lat rounded to 6dp, got: %.10f", result.Position.Latitude)
}
}

func TestNormalizer_DoesNotMutateOriginal(t *testing.T) {
n := domain.NewRadarNormalizer()
obs := validObs()
obs.SensorId = "  RADAR-001  "
originalID := obs.SensorId

_ = n.Normalize(obs)
if obs.SensorId != originalID {
t.Error("Normalize must not mutate the original observation")
}
}

func TestNormalizer_NilPosition(t *testing.T) {
n := domain.NewRadarNormalizer()
obs := validObs()
obs.Position = nil

result := n.Normalize(obs)
if result == nil {
t.Fatal("expected non-nil result")
}
if result.Position != nil {
t.Error("expected nil position to remain nil")
}
}

func TestNormalizer_NilRadar(t *testing.T) {
n := domain.NewRadarNormalizer()
obs := &ingestionv1.SensorObservation{
SensorId: "  RADAR-002  ",
}
result := n.Normalize(obs)
if result.SensorId != "RADAR-002" {
t.Errorf("expected trimmed sensor_id, got: %q", result.SensorId)
}
}
