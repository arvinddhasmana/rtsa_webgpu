// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-radar-ingestion/internal/domain"
"go.uber.org/zap"
"google.golang.org/protobuf/types/known/timestamppb"
)

func newValidator() *domain.RadarValidator {
return domain.NewRadarValidator(zap.NewNop())
}

func validObs() *ingestionv1.SensorObservation {
speed := float64(15.0)
heading := float64(180.0)
return &ingestionv1.SensorObservation{
SensorId:        "RADAR-001",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:       45.0,
Longitude:      -60.0,
SpeedKnots:     &speed,
HeadingDegrees: &heading,
},
SensorData: &ingestionv1.SensorObservation_Radar{
Radar: &ingestionv1.RadarTrack{
TrackNumber:    "TRK-001",
RangeNm:        25.0,
BearingDegrees: 90.0,
TrackQuality:   0.85,
},
},
}
}

// T01: Valid radar observation
func TestValidator_T01_ValidObservation(t *testing.T) {
v := newValidator()
result := v.Validate(validObs())
if !result.Valid {
t.Errorf("expected valid, got errors: %v", result.Errors)
}
if len(result.Errors) != 0 {
t.Errorf("expected no errors, got: %v", result.Errors)
}
}

// T02: Missing sensor_id
func TestValidator_T02_MissingSensorID(t *testing.T) {
v := newValidator()
obs := validObs()
obs.SensorId = ""
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "sensor_id" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on sensor_id field")
}
}

// T03: Wrong sensor_type
func TestValidator_T03_WrongSensorType(t *testing.T) {
v := newValidator()
obs := validObs()
obs.SensorType = commonv1.SensorType_SENSOR_TYPE_EW_SIGINT
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "sensor_type" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on sensor_type")
}
}

// T04: Latitude out of range
func TestValidator_T04_LatitudeOutOfRange(t *testing.T) {
v := newValidator()
obs := validObs()
obs.Position.Latitude = 91.0
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "position.latitude" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on position.latitude")
}
}

// T05: Longitude out of range
func TestValidator_T05_LongitudeOutOfRange(t *testing.T) {
v := newValidator()
obs := validObs()
obs.Position.Longitude = -181.0
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "position.longitude" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on position.longitude")
}
}

// T06: Heading out of range
func TestValidator_T06_HeadingOutOfRange(t *testing.T) {
v := newValidator()
obs := validObs()
heading := float64(361.0)
obs.Position.HeadingDegrees = &heading
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "position.heading_degrees" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on position.heading_degrees")
}
}

// T07: Future timestamp
func TestValidator_T07_FutureTimestamp(t *testing.T) {
v := newValidator()
obs := validObs()
obs.ObservationTime = timestamppb.New(time.Now().UTC().Add(20 * time.Minute))
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid for future timestamp")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "observation_time" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on observation_time")
}
}

// T08: Past timestamp (> 24h ago)
func TestValidator_T08_PastTimestamp(t *testing.T) {
v := newValidator()
obs := validObs()
obs.ObservationTime = timestamppb.New(time.Now().UTC().Add(-25 * time.Hour))
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid for old timestamp")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "observation_time" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on observation_time")
}
}

// T09: Speed warning (surface track, speed > maxSurface)
func TestValidator_T09_SpeedWarning(t *testing.T) {
v := domain.NewRadarValidator(zap.NewNop(),
domain.WithMaxSurfaceSpeedKnots(100.0))
obs := validObs()
speed := float64(1500.0) // exceeds surface threshold
obs.Position.SpeedKnots = &speed
result := v.Validate(obs)
if !result.Valid {
t.Error("expected valid (warning only, not error)")
}
if len(result.Warnings) == 0 {
t.Error("expected warning for high speed")
}
}

// T10: Missing classification (UNSPECIFIED)
func TestValidator_T10_MissingClassification(t *testing.T) {
v := newValidator()
obs := validObs()
obs.Classification = commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "classification" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on classification")
}
}

// T11: Missing radar track_number
func TestValidator_T11_MissingTrackNumber(t *testing.T) {
v := newValidator()
obs := validObs()
obs.GetRadar().TrackNumber = ""
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "radar.track_number" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on radar.track_number")
}
}

// T12: Track quality out of range
func TestValidator_T12_TrackQualityOutOfRange(t *testing.T) {
v := newValidator()
obs := validObs()
obs.GetRadar().TrackQuality = 1.5
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid")
}
hasErr := false
for _, e := range result.Errors {
if e.Field == "radar.track_quality" {
hasErr = true
break
}
}
if !hasErr {
t.Error("expected error on radar.track_quality")
}
}

// T13: No position (valid for some radars)
func TestValidator_T13_NoPosition(t *testing.T) {
v := newValidator()
obs := validObs()
obs.Position = nil
result := v.Validate(obs)
if !result.Valid {
t.Errorf("expected valid without position, got errors: %v", result.Errors)
}
}

// T14: Multiple errors
func TestValidator_T14_MultipleErrors(t *testing.T) {
v := newValidator()
obs := validObs()
obs.Position.Latitude = 91.0
obs.Position.Longitude = 181.0
speed := float64(-1.0) // negative speed — warning
obs.Position.SpeedKnots = &speed
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid with multiple errors")
}
if len(result.Errors) < 2 {
t.Errorf("expected at least 2 errors, got %d: %v", len(result.Errors), result.Errors)
}
}

func TestValidator_SensorIDTooLong(t *testing.T) {
v := newValidator()
obs := validObs()
// 129 chars
obs.SensorId = "RADAR-" + string(make([]byte, 129))
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid for too-long sensor_id")
}
}

func TestValidator_BearingWarning(t *testing.T) {
v := newValidator()
obs := validObs()
obs.GetRadar().BearingDegrees = -5.0 // out of [0-360] → warning
result := v.Validate(obs)
// Bearing is a soft warning
if len(result.Warnings) == 0 {
t.Error("expected warning for bearing out of range")
}
}

func TestValidator_RangeWarning(t *testing.T) {
v := newValidator()
obs := validObs()
obs.GetRadar().RangeNm = 600.0 // > 500 → warning
result := v.Validate(obs)
if len(result.Warnings) == 0 {
t.Error("expected warning for range_nm > 500")
}
}

func TestValidator_NilObservationTime(t *testing.T) {
v := newValidator()
obs := validObs()
obs.ObservationTime = nil // nil time — no temporal errors expected
result := v.Validate(obs)
hasTimeErr := false
for _, e := range result.Errors {
if e.Field == "observation_time" {
hasTimeErr = true
}
}
if hasTimeErr {
t.Error("nil observation_time should not produce a temporal error")
}
}

func TestValidator_NilRadarData(t *testing.T) {
v := newValidator()
obs := validObs()
obs.SensorData = nil // no radar data
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid when radar data is nil")
}
}

func TestValidator_OptionsCoverage(t *testing.T) {
	v := domain.NewRadarValidator(zap.NewNop(),
		domain.WithMaxAirSpeedKnots(2000.0),
		domain.WithMaxFutureOffset(1*time.Minute),
		domain.WithMaxPastOffset(1*time.Minute),
	)
	if v == nil {
		t.Fatal("expected non-nil validator")
	}
}
