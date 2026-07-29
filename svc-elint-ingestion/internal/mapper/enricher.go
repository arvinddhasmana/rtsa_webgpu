// CLASSIFICATION: UNCLASSIFIED
package mapper

import (
"context"
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
"github.com/google/uuid"
)

// Enricher adds system-generated fields to observations.
type Enricher struct {
serviceName string
guard       *classification.Guard
}

// NewEnricher creates an enricher.
func NewEnricher(serviceName string, guard *classification.Guard) *Enricher {
return &Enricher{
serviceName: serviceName,
guard:       guard,
}
}

// Enrich adds system-generated metadata to the observation.
// - observation_id: UUID if not already set
// - Classification ceiling check (returns error if above ceiling)
// - metadata["rtsa.source_service"] = serviceName
// - metadata["rtsa.ingestion_time"] = current UTC time
func (e *Enricher) Enrich(ctx context.Context, obs *ingestionv1.SensorObservation) error {
if obs.ObservationId == "" {
obs.ObservationId = uuid.New().String()
}

if e.guard != nil {
if err := e.guard.Check(obs.GetClassification()); err != nil {
return fmt.Errorf("enricher: classification violation: %w", err)
}
}

if obs.Metadata == nil {
obs.Metadata = make(map[string]string)
}
obs.Metadata["rtsa.source_service"] = e.serviceName
obs.Metadata["rtsa.ingestion_time"] = time.Now().UTC().Format(time.RFC3339Nano)

return nil
}

// ClassificationLevel returns the classification level of the observation.
func ClassificationLevel(obs *ingestionv1.SensorObservation) commonv1.ClassificationLevel {
return obs.GetClassification()
}
