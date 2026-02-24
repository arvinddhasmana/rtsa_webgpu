// CLASSIFICATION: UNCLASSIFIED
package ingestion

import (
"context"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
)

// ValidationResult holds validation outcome.
type ValidationResult struct {
Valid   bool
Errors  []ValidationError
Suspect bool // true if flagged suspect but not rejected
}

// ValidationError holds a single validation failure.
type ValidationError struct {
Field   string
Rule    string
Message string
}

// Validator validates a SensorObservation.
type Validator interface {
Validate(obs *ingestionv1.SensorObservation) ValidationResult
}

// Normalizer normalizes a raw SensorObservation in-place.
type Normalizer interface {
Normalize(obs *ingestionv1.SensorObservation)
}

// Producer produces a SensorObservation to the output topic.
type Producer interface {
Produce(ctx context.Context, obs *ingestionv1.SensorObservation) error
Close() error
}
