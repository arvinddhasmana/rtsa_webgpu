// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"strings"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

// Normalizer normalizes EW/SIGINT observations.
type Normalizer struct{}

// NewNormalizer creates a new EW Normalizer.
func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize trims whitespace and normalizes EW fields.
func (n *Normalizer) Normalize(obs *ingestionv1.SensorObservation) {
obs.SensorId = ingestion.TrimString(obs.GetSensorId())
if ew := obs.GetEwSigint(); ew != nil {
ew.EmitterId = ingestion.TrimString(ew.GetEmitterId())
ew.ModulationType = strings.ToUpper(ingestion.TrimString(ew.GetModulationType()))
}
}
