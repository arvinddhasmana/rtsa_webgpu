// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"strings"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

// Normalizer normalizes Cyber IOC observations.
type Normalizer struct{}

// NewNormalizer creates a new Cyber Normalizer.
func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize trims whitespace and normalizes Cyber fields.
func (n *Normalizer) Normalize(obs *ingestionv1.SensorObservation) {
obs.SensorId = ingestion.TrimString(obs.GetSensorId())
if c := obs.GetCyber(); c != nil {
c.IocType = strings.ToLower(ingestion.TrimString(c.GetIocType()))
c.IocValue = ingestion.TrimString(c.GetIocValue())
c.SourceFeed = ingestion.TrimString(c.GetSourceFeed())
c.DedupHash = strings.ToLower(ingestion.TrimString(c.GetDedupHash()))
}
}
