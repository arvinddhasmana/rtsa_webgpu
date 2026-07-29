// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/pkg/ingestion"
)

var validSensorNames = map[string]bool{
"EO":  true,
"IR":  true,
"SAR": true,
"MTI": true,
}

// Validator validates ISR observations.
type Validator struct{}

// NewValidator creates a new ISR Validator.
func NewValidator() *Validator { return &Validator{} }

// Validate checks ISR-specific rules.
func (v *Validator) Validate(obs *ingestionv1.SensorObservation) ingestion.ValidationResult {
var errs []ingestion.ValidationError

if obs.GetSensorId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "required", Message: "sensor_id is required"})
} else if len(obs.GetSensorId()) > 128 {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "max_length", Message: "sensor_id exceeds 128 chars"})
}

if obs.GetSensorType() != commonv1.SensorType_SENSOR_TYPE_ISR {
errs = append(errs, ingestion.ValidationError{Field: "sensor_type", Rule: "enum", Message: fmt.Sprintf("sensor_type must be SENSOR_TYPE_ISR, got %s", obs.GetSensorType())})
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

isr := obs.GetIsr()
if isr == nil {
errs = append(errs, ingestion.ValidationError{Field: "isr", Rule: "required", Message: "isr payload is required"})
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

if isr.GetPlatformId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "isr.platform_id", Rule: "required", Message: "isr.platform_id is required"})
}

if !validSensorNames[isr.GetSensorName()] {
errs = append(errs, ingestion.ValidationError{Field: "isr.sensor_name", Rule: "enum", Message: fmt.Sprintf("isr.sensor_name must be one of EO, IR, SAR, MTI; got %s", isr.GetSensorName())})
}

if isr.GetImageId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "isr.image_id", Rule: "required", Message: "isr.image_id is required"})
}

// Coverage polygon: at least 3 vertices
poly := isr.GetCoveragePolygon()
if len(poly) < 3 {
errs = append(errs, ingestion.ValidationError{Field: "isr.coverage_polygon", Rule: "min_vertices", Message: fmt.Sprintf("isr.coverage_polygon must have at least 3 vertices, got %d", len(poly))})
} else {
for i, p := range poly {
if p.GetLatitude() < -90 || p.GetLatitude() > 90 {
errs = append(errs, ingestion.ValidationError{Field: fmt.Sprintf("isr.coverage_polygon[%d].latitude", i), Rule: "range", Message: fmt.Sprintf("polygon vertex %d latitude out of range: %.4f", i, p.GetLatitude())})
}
if p.GetLongitude() < -180 || p.GetLongitude() > 180 {
errs = append(errs, ingestion.ValidationError{Field: fmt.Sprintf("isr.coverage_polygon[%d].longitude", i), Rule: "range", Message: fmt.Sprintf("polygon vertex %d longitude out of range: %.4f", i, p.GetLongitude())})
}
}
}

	// Detection confidences
	for i, det := range isr.GetDetections() {
		if det.GetConfidence() < 0.0 || det.GetConfidence() > 1.0 {
			errs = append(errs, ingestion.ValidationError{Field: fmt.Sprintf("isr.detections[%d].confidence", i), Rule: "range", Message: fmt.Sprintf("detection %d confidence out of range [0.0, 1.0]: %.2f", i, det.GetConfidence())})
		}
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
