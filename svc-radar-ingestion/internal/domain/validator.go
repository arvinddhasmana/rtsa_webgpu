// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"go.uber.org/zap"
)

// ValidationResult holds the outcome of validation.
type ValidationResult struct {
Valid    bool
Errors   []ValidationError
Warnings []string
}

// ValidationError describes a single validation failure.
type ValidationError struct {
Field   string
Value   string
Rule    string
Message string
}

// ValidatorOption configures a RadarValidator.
type ValidatorOption func(*RadarValidator)

// WithMaxSurfaceSpeedKnots sets the maximum surface speed threshold.
func WithMaxSurfaceSpeedKnots(v float64) ValidatorOption {
return func(r *RadarValidator) { r.maxSurfaceSpeedKnots = v }
}

// WithMaxAirSpeedKnots sets the maximum air speed threshold.
func WithMaxAirSpeedKnots(v float64) ValidatorOption {
return func(r *RadarValidator) { r.maxAirSpeedKnots = v }
}

// WithMaxFutureOffset sets the maximum future time offset.
func WithMaxFutureOffset(d time.Duration) ValidatorOption {
return func(r *RadarValidator) { r.maxFutureOffset = d }
}

// WithMaxPastOffset sets the maximum past time offset.
func WithMaxPastOffset(d time.Duration) ValidatorOption {
return func(r *RadarValidator) { r.maxPastOffset = d }
}

// RadarValidator validates radar-specific SensorObservation messages.
type RadarValidator struct {
logger               *zap.Logger
maxSurfaceSpeedKnots float64
maxAirSpeedKnots     float64
maxFutureOffset      time.Duration
maxPastOffset        time.Duration
}

// NewRadarValidator creates a validator with configured thresholds.
func NewRadarValidator(logger *zap.Logger, opts ...ValidatorOption) *RadarValidator {
v := &RadarValidator{
logger:               logger,
maxSurfaceSpeedKnots: 999.0,
maxAirSpeedKnots:     2500.0,
maxFutureOffset:      300 * time.Second,
maxPastOffset:        86400 * time.Second,
}
for _, opt := range opts {
opt(v)
}
return v
}

// Validate checks all validation rules for a SensorObservation.
func (v *RadarValidator) Validate(obs *ingestionv1.SensorObservation) *ValidationResult {
result := &ValidationResult{Valid: true}

// ── Hard failures (REJECT TO DLQ) ──

// sensor_id: non-empty, max 128 chars
sensorID := obs.GetSensorId()
if sensorID == "" {
result.addError("sensor_id", "", "non-empty", "sensor_id must not be empty")
} else if len(sensorID) > 128 {
result.addError("sensor_id", sensorID[:10]+"...", "max-128-chars",
"sensor_id exceeds maximum length of 128 characters")
}

// sensor_type: must be SENSOR_TYPE_RADAR
if obs.GetSensorType() != commonv1.SensorType_SENSOR_TYPE_RADAR {
result.addError("sensor_type",
fmt.Sprintf("%v", obs.GetSensorType()),
"must-be-RADAR",
fmt.Sprintf("sensor_type must be SENSOR_TYPE_RADAR, got %v", obs.GetSensorType()))
}

// observation_time: temporal bounds
if obs.GetObservationTime() != nil {
obsTime := obs.GetObservationTime().AsTime()
now := time.Now().UTC()
if obsTime.After(now.Add(v.maxFutureOffset)) {
result.addError("observation_time",
obsTime.Format(time.RFC3339),
fmt.Sprintf("max-future-offset:%s", v.maxFutureOffset),
fmt.Sprintf("observation_time is too far in the future (max offset: %s)", v.maxFutureOffset))
}
if obsTime.Before(now.Add(-v.maxPastOffset)) {
result.addError("observation_time",
obsTime.Format(time.RFC3339),
fmt.Sprintf("max-past-offset:%s", v.maxPastOffset),
fmt.Sprintf("observation_time is too far in the past (max offset: %s)", v.maxPastOffset))
}
}

// position validation (if position provided)
if pos := obs.GetPosition(); pos != nil {
if pos.GetLatitude() < -90.0 || pos.GetLatitude() > 90.0 {
result.addError("position.latitude",
fmt.Sprintf("%.6f", pos.GetLatitude()),
"range:-90.0 to +90.0",
fmt.Sprintf("position.latitude out of range: %.4f (must be -90.0 to +90.0)", pos.GetLatitude()))
}
if pos.GetLongitude() < -180.0 || pos.GetLongitude() > 180.0 {
result.addError("position.longitude",
fmt.Sprintf("%.6f", pos.GetLongitude()),
"range:-180.0 to +180.0",
fmt.Sprintf("position.longitude out of range: %.4f (must be -180.0 to +180.0)", pos.GetLongitude()))
}
if pos.HeadingDegrees != nil {
h := pos.GetHeadingDegrees()
if h < 0.0 || h > 360.0 {
result.addError("position.heading_degrees",
fmt.Sprintf("%.2f", h),
"range:0.0 to 360.0",
fmt.Sprintf("position.heading_degrees out of range: %.2f (must be 0.0 to 360.0)", h))
}
}

// Soft warning: speed
if pos.SpeedKnots != nil {
speed := pos.GetSpeedKnots()
if speed < 0 || speed > v.maxSurfaceSpeedKnots {
result.Warnings = append(result.Warnings,
fmt.Sprintf("position.speed_knots suspect: %.1f (max surface: %.1f)",
speed, v.maxSurfaceSpeedKnots))
}
}
}

// classification: valid enum value (not UNSPECIFIED)
cls := obs.GetClassification()
if cls == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
result.addError("classification",
fmt.Sprintf("%v", cls),
"not-UNSPECIFIED",
"classification must not be CLASSIFICATION_LEVEL_UNSPECIFIED")
}

// radar-specific validation
radar := obs.GetRadar()
if radar == nil {
result.addError("radar", "nil", "required", "radar track data is required for RADAR sensor type")
} else {
if radar.GetTrackNumber() == "" {
result.addError("radar.track_number", "", "non-empty", "radar.track_number must not be empty")
}
if radar.GetTrackQuality() < 0.0 || radar.GetTrackQuality() > 1.0 {
result.addError("radar.track_quality",
fmt.Sprintf("%.2f", radar.GetTrackQuality()),
"range:0.0 to 1.0",
fmt.Sprintf("radar.track_quality out of range: %.2f (must be 0.0 to 1.0)", radar.GetTrackQuality()))
}

// Soft warnings
if radar.GetRangeNm() < 0 || radar.GetRangeNm() > 500 {
result.Warnings = append(result.Warnings,
fmt.Sprintf("radar.range_nm suspect: %.1f (expected 0-500)", radar.GetRangeNm()))
}
if radar.GetBearingDegrees() < 0.0 || radar.GetBearingDegrees() > 360.0 {
result.Warnings = append(result.Warnings,
fmt.Sprintf("radar.bearing_degrees suspect: %.2f (expected 0.0-360.0)", radar.GetBearingDegrees()))
}
}

return result
}

func (r *ValidationResult) addError(field, value, rule, message string) {
r.Valid = false
r.Errors = append(r.Errors, ValidationError{
Field:   field,
Value:   value,
Rule:    rule,
Message: message,
})
}
