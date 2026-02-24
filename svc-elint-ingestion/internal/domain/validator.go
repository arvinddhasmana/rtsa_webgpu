// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

var validScanTypes = map[string]bool{
"circular":         true,
"sector":           true,
"track-while-scan": true,
}

// Validator validates ELINT/COMINT observations.
type Validator struct{}

// NewValidator creates a new ELINT Validator.
func NewValidator() *Validator { return &Validator{} }

// Validate checks ELINT/COMINT-specific rules.
func (v *Validator) Validate(obs *ingestionv1.SensorObservation) ingestion.ValidationResult {
var errs []ingestion.ValidationError

if obs.GetSensorId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "required", Message: "sensor_id is required"})
} else if len(obs.GetSensorId()) > 128 {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "max_length", Message: "sensor_id exceeds 128 chars"})
}

if obs.GetSensorType() != commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT {
errs = append(errs, ingestion.ValidationError{Field: "sensor_type", Rule: "enum", Message: fmt.Sprintf("sensor_type must be SENSOR_TYPE_ELINT_COMINT, got %s", obs.GetSensorType())})
}

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

if obs.GetClassification() == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
errs = append(errs, ingestion.ValidationError{Field: "classification", Rule: "required", Message: "classification must not be UNSPECIFIED"})
}

elint := obs.GetElintComint()
if elint == nil {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint", Rule: "required", Message: "elint_comint payload is required"})
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

if elint.GetEmitterId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.emitter_id", Rule: "required", Message: "elint_comint.emitter_id is required"})
}

if elint.GetRadarType() == "" {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.radar_type", Rule: "required", Message: "elint_comint.radar_type is required"})
}

freq := elint.GetFrequencyMhz()
if freq < 0.5 || freq > 40000 {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.frequency_mhz", Rule: "range", Message: fmt.Sprintf("elint_comint.frequency_mhz out of range [0.5, 40000]: %.2f", freq)})
}

if elint.GetCepMeters() <= 0 {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.cep_meters", Rule: "positive", Message: "elint_comint.cep_meters must be > 0"})
}

conf := elint.GetConfidence()
if conf < 0.0 || conf > 1.0 {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.confidence", Rule: "range", Message: fmt.Sprintf("elint_comint.confidence out of range [0.0, 1.0]: %.2f", conf)})
}

if elint.GetContentClassification() == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.content_classification", Rule: "required", Message: "elint_comint.content_classification must not be UNSPECIFIED"})
}

// content_classification must not exceed metadata classification
if elint.GetContentClassification() > obs.GetClassification() {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.content_classification", Rule: "classification_ceiling", Message: "content_classification exceeds metadata classification"})
}

// scan_type validation
scanType := elint.GetScanType()
if scanType != "" && !validScanTypes[scanType] {
errs = append(errs, ingestion.ValidationError{Field: "elint_comint.scan_type", Rule: "enum", Message: fmt.Sprintf("invalid scan_type: %s", scanType)})
}

// position (optional)
if pos := obs.GetPosition(); pos != nil {
if pos.GetLatitude() < -90 || pos.GetLatitude() > 90 {
errs = append(errs, ingestion.ValidationError{Field: "position.latitude", Rule: "range", Message: fmt.Sprintf("position.latitude out of range: %.4f", pos.GetLatitude())})
}
if pos.GetLongitude() < -180 || pos.GetLongitude() > 180 {
errs = append(errs, ingestion.ValidationError{Field: "position.longitude", Rule: "range", Message: fmt.Sprintf("position.longitude out of range: %.4f", pos.GetLongitude())})
}
}

if len(errs) > 0 {
return ingestion.ValidationResult{Valid: false, Errors: errs}
}
return ingestion.ValidationResult{Valid: true}
}
