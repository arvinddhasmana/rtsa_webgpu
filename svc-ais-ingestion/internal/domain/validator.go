// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"math"
"regexp"
"sync"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

var (
mmsiRe              = regexp.MustCompile(`^\d{9}$`)
validAISMessageTypes = map[int32]bool{1: true, 2: true, 3: true, 5: true, 18: true, 24: true}
)

// lastPosition holds the last known position for AIS spoofing detection.
type lastPosition struct {
lat float64
lon float64
ts  time.Time
}

// Validator validates AIS/BFT observations with spoofing detection.
type Validator struct {
mu            sync.Mutex
lastPositions map[string]*lastPosition // keyed by MMSI
}

// NewValidator creates a new AIS Validator.
func NewValidator() *Validator {
return &Validator{lastPositions: make(map[string]*lastPosition)}
}

// Validate checks AIS/BFT-specific rules.
func (v *Validator) Validate(obs *ingestionv1.SensorObservation) ingestion.ValidationResult {
var errs []ingestion.ValidationError
suspect := false

if obs.GetSensorId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "required", Message: "sensor_id is required"})
} else if len(obs.GetSensorId()) > 128 {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "max_length", Message: "sensor_id exceeds 128 chars"})
}

if obs.GetSensorType() != commonv1.SensorType_SENSOR_TYPE_AIS_BFT {
errs = append(errs, ingestion.ValidationError{Field: "sensor_type", Rule: "enum", Message: fmt.Sprintf("sensor_type must be SENSOR_TYPE_AIS_BFT, got %s", obs.GetSensorType())})
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

ais := obs.GetAisBft()
if ais == nil {
errs = append(errs, ingestion.ValidationError{Field: "ais_bft", Rule: "required", Message: "ais_bft payload is required"})
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

// MMSI: 9 digits
if !mmsiRe.MatchString(ais.GetMmsi()) {
errs = append(errs, ingestion.ValidationError{Field: "ais_bft.mmsi", Rule: "format", Message: fmt.Sprintf("ais_bft.mmsi must be 9 digits, got %q", ais.GetMmsi())})
}

// Vessel type code: 1-99
vtc := ais.GetVesselTypeCode()
if vtc < 1 || vtc > 99 {
errs = append(errs, ingestion.ValidationError{Field: "ais_bft.vessel_type_code", Rule: "range", Message: fmt.Sprintf("ais_bft.vessel_type_code must be 1-99, got %d", vtc)})
}

// AIS message type
if !validAISMessageTypes[ais.GetAisMessageType()] {
errs = append(errs, ingestion.ValidationError{Field: "ais_bft.ais_message_type", Rule: "enum", Message: fmt.Sprintf("ais_bft.ais_message_type must be one of 1,2,3,5,18,24; got %d", ais.GetAisMessageType())})
}

// Position: required for AIS
pos := obs.GetPosition()
if pos == nil {
errs = append(errs, ingestion.ValidationError{Field: "position", Rule: "required", Message: "position is required for AIS"})
} else {
if pos.GetLatitude() < -90 || pos.GetLatitude() > 90 {
errs = append(errs, ingestion.ValidationError{Field: "position.latitude", Rule: "range", Message: fmt.Sprintf("position.latitude out of range: %.4f", pos.GetLatitude())})
}
if pos.GetLongitude() < -180 || pos.GetLongitude() > 180 {
errs = append(errs, ingestion.ValidationError{Field: "position.longitude", Rule: "range", Message: fmt.Sprintf("position.longitude out of range: %.4f", pos.GetLongitude())})
}
}

// BFT classification check
if ais.GetIsBft() && obs.GetClassification() < commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B {
errs = append(errs, ingestion.ValidationError{Field: "classification", Rule: "bft_classification", Message: "BFT position must be classified PROTECTED_B or higher"})
}

if len(errs) > 0 {
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

// Spoofing detection (only if position provided)
if pos != nil && ais.GetMmsi() != "" && obs.GetObservationTime() != nil {
suspect = v.checkSpoofing(ais.GetMmsi(), pos.GetLatitude(), pos.GetLongitude(), obs.GetObservationTime().AsTime())
}

return ingestion.ValidationResult{Valid: true, Suspect: suspect}
}

// checkSpoofing detects position and speed anomalies.
// Returns true if the report looks suspicious (position jump > ~100 NM or speed > 50 kts).
func (v *Validator) checkSpoofing(mmsi string, lat, lon float64, ts time.Time) bool {
v.mu.Lock()
defer v.mu.Unlock()

prev, ok := v.lastPositions[mmsi]
v.lastPositions[mmsi] = &lastPosition{lat: lat, lon: lon, ts: ts}
if !ok {
return false
}

// Time delta in hours
dt := ts.Sub(prev.ts).Hours()
if dt <= 0 {
return false
}

// Approximate distance in nautical miles using equirectangular projection
// 1 degree lat ≈ 60 NM; 1 degree lon ≈ 60 * cos(lat) NM
midLat := (lat + prev.lat) / 2.0
dlat := (lat - prev.lat) * 60.0
dlon := (lon - prev.lon) * 60.0 * math.Cos(midLat*math.Pi/180.0)
distNM := math.Sqrt(dlat*dlat + dlon*dlon)

// Position jump > 100 NM is suspicious
if distNM > 100 {
return true
}

// Speed > 50 knots is suspicious
speedKnots := distNM / dt
if speedKnots > 50 {
return true
}

return false
}
