// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"strings"

ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/ingestion"
)

// Normalizer normalizes AIS/BFT observations.
type Normalizer struct{}

// NewNormalizer creates a new AIS Normalizer.
func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize trims whitespace and normalizes AIS fields.
func (n *Normalizer) Normalize(obs *ingestionv1.SensorObservation) {
obs.SensorId = ingestion.TrimString(obs.GetSensorId())
if ais := obs.GetAisBft(); ais != nil {
ais.Mmsi = ingestion.TrimString(ais.GetMmsi())
ais.VesselName = strings.ToUpper(ingestion.TrimString(ais.GetVesselName()))
ais.NavStatus = ingestion.TrimString(ais.GetNavStatus())
}
}
