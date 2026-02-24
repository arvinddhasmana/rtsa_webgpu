// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"strings"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

// Normalizer normalizes ELINT/COMINT observations.
type Normalizer struct{}

// NewNormalizer creates a new ELINT Normalizer.
func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize trims whitespace and normalizes ELINT fields.
func (n *Normalizer) Normalize(obs *ingestionv1.SensorObservation) {
obs.SensorId = ingestion.TrimString(obs.GetSensorId())
if e := obs.GetElintComint(); e != nil {
e.EmitterId = ingestion.TrimString(e.GetEmitterId())
e.RadarType = ingestion.TrimString(e.GetRadarType())
e.ScanType = strings.ToLower(ingestion.TrimString(e.GetScanType()))
}
}
