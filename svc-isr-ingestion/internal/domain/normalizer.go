// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"strings"

ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/ingestion"
)

// Normalizer normalizes ISR observations.
type Normalizer struct{}

// NewNormalizer creates a new ISR Normalizer.
func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize trims whitespace and normalizes ISR fields.
func (n *Normalizer) Normalize(obs *ingestionv1.SensorObservation) {
obs.SensorId = ingestion.TrimString(obs.GetSensorId())
if isr := obs.GetIsr(); isr != nil {
isr.PlatformId = ingestion.TrimString(isr.GetPlatformId())
isr.SensorName = strings.ToUpper(ingestion.TrimString(isr.GetSensorName()))
isr.ImageId = ingestion.TrimString(isr.GetImageId())
}
}
