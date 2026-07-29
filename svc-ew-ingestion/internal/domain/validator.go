// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/ingestion"
)

// Validator validates EW/SIGINT observations.
type Validator struct{}

// NewValidator creates a new EW Validator.
func NewValidator() *Validator { return &Validator{} }

// Validate checks EW/SIGINT-specific rules.
func (v *Validator) Validate(obs *ingestionv1.SensorObservation) ingestion.ValidationResult {
var errs []ingestion.ValidationError

// sensor_id
if obs.GetSensorId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "required", Message: "sensor_id is required"})
} else if len(obs.GetSensorId()) > 128 {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "max_length", Message: "sensor_id exceeds 128 chars"})
}

// sensor_type
if obs.GetSensorType() != commonv1.SensorType_SENSOR_TYPE_EW_SIGINT {
errs = append(errs, ingestion.ValidationError{Field: "sensor_type", Rule: "enum", Message: fmt.Sprintf("sensor_type must be SENSOR_TYPE_EW_SIGINT, got %s", obs.GetSensorType())})
}

// observation_time
if obs.GetObservationTime() == nil {
errs = append(errs, ingestion.ValidationError{Field: "observation_time", Rule: "required", Message: "observation_time is required"})
} else {
t := obs.GetObservationTime().AsTime()
now := time.Now()
if t.After(now.Add(5 * time.Minute)) {
errs = append(errs, ingestion.ValidationError{Field: "observation_time", Rule: "future", Message: "observation_time is more than 5 minutes in the future"})
}
if t.Before(now.Add(-24 * time.Hour)) {
errs = append(errs, ingestion.ValidationError{Field: "observation_time", Rule: "past", Message: "observation_time is more than 24 hours in the past"})
}
}

// classification
if obs.GetClassification() == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
errs = append(errs, ingestion.ValidationError{Field: "classification", Rule: "required", Message: "classification must not be UNSPECIFIED"})
}

// EW-specific fields
ew := obs.GetEwSigint()
if ew == nil {
errs = append(errs, ingestion.ValidationError{Field: "ew_sigint", Rule: "required", Message: "ew_sigint payload is required"})
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

if ew.GetEmitterId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "ew_sigint.emitter_id", Rule: "required", Message: "ew_sigint.emitter_id is required"})
}

freq := ew.GetFrequencyMhz()
if freq < 0.5 || freq > 40000 {
errs = append(errs, ingestion.ValidationError{Field: "ew_sigint.frequency_mhz", Rule: "range", Message: fmt.Sprintf("ew_sigint.frequency_mhz out of range [0.5, 40000]: %.2f", freq)})
}

bearing := ew.GetBearingDegrees()
if bearing < 0.0 || bearing > 360.0 {
errs = append(errs, ingestion.ValidationError{Field: "ew_sigint.bearing_degrees", Rule: "range", Message: fmt.Sprintf("ew_sigint.bearing_degrees out of range [0.0, 360.0]: %.2f", bearing)})
}

conf := ew.GetConfidence()
if conf < 0.0 || conf > 1.0 {
errs = append(errs, ingestion.ValidationError{Field: "ew_sigint.confidence", Rule: "range", Message: fmt.Sprintf("ew_sigint.confidence out of range [0.0, 1.0]: %.2f", conf)})
}

if ew.GetModulationType() == "" {
errs = append(errs, ingestion.ValidationError{Field: "ew_sigint.modulation_type", Rule: "required", Message: "ew_sigint.modulation_type is required"})
}

// position (optional)
if pos := obs.GetPosition(); pos != nil {
if pos.GetLatitude() < -90 || pos.GetLatitude() > 90 {
errs = append(errs, ingestion.ValidationError{Field: "position.latitude", Rule: "range", Message: fmt.Sprintf("position.latitude out of range [-90, 90]: %.4f", pos.GetLatitude())})
}
if pos.GetLongitude() < -180 || pos.GetLongitude() > 180 {
errs = append(errs, ingestion.ValidationError{Field: "position.longitude", Rule: "range", Message: fmt.Sprintf("position.longitude out of range [-180, 180]: %.4f", pos.GetLongitude())})
}
}

if len(errs) > 0 {
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

// Suspect flags (do not reject)
suspect := false
if ew.PowerDbm != nil {
p := ew.GetPowerDbm()
if p < -200 || p > 100 {
suspect = true
}
}
if ew.PriMicroseconds != nil {
pri := ew.GetPriMicroseconds()
if pri < 0.1 || pri > 100000 {
suspect = true
}
}

return ingestion.ValidationResult{Valid: true, Suspect: suspect}
}
