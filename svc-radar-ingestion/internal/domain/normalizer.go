// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"math"
"strings"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"google.golang.org/protobuf/proto"
)

// RadarNormalizer ensures consistent field formatting.
type RadarNormalizer struct{}

// NewRadarNormalizer creates a normalizer.
func NewRadarNormalizer() *RadarNormalizer {
return &RadarNormalizer{}
}

// Normalize standardizes the observation:
// - Trims whitespace from string fields
// - Normalizes heading to 0-360 range
// - Ensures position coordinates use correct precision (6 decimal places)
// - Clones the input to avoid mutating the original
func (n *RadarNormalizer) Normalize(obs *ingestionv1.SensorObservation) *ingestionv1.SensorObservation {
// Deep clone to avoid mutating the original
normalized := proto.Clone(obs).(*ingestionv1.SensorObservation)

// Trim string fields
normalized.SensorId = strings.TrimSpace(normalized.SensorId)
normalized.ObservationId = strings.TrimSpace(normalized.ObservationId)

// Normalize metadata string values
for k, v := range normalized.Metadata {
normalized.Metadata[k] = strings.TrimSpace(v)
}

// Normalize position
if pos := normalized.Position; pos != nil {
// Round lat/lon to 6 decimal places
pos.Latitude = roundTo(pos.Latitude, 6)
pos.Longitude = roundTo(pos.Longitude, 6)

// Normalize heading to [0, 360)
if pos.HeadingDegrees != nil {
h := normalizeHeading(pos.GetHeadingDegrees())
pos.HeadingDegrees = &h
}
}

// Normalize radar fields
if radar := normalized.GetRadar(); radar != nil {
radar.TrackNumber = strings.TrimSpace(radar.TrackNumber)
// Normalize bearing to [0, 360)
if radar.BearingDegrees < 0 || radar.BearingDegrees >= 360 {
radar.BearingDegrees = normalizeBearing(radar.BearingDegrees)
}
}

return normalized
}

// normalizeHeading normalizes a heading value to the range [0, 360).
func normalizeHeading(h float64) float64 {
h = math.Mod(h, 360.0)
if h < 0 {
h += 360.0
}
return h
}

// normalizeBearing normalizes a bearing to [0, 360).
func normalizeBearing(b float64) float64 {
b = math.Mod(b, 360.0)
if b < 0 {
b += 360.0
}
return b
}

// roundTo rounds a float to n decimal places.
func roundTo(v float64, decimals int) float64 {
factor := math.Pow(10, float64(decimals))
return math.Round(v*factor) / factor
}
