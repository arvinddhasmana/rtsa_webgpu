// CLASSIFICATION: UNCLASSIFIED
// Package generator — InvalidObservationInjector
//
// Corrupts a configurable fraction of sensor observations to exercise
// ingestion-layer DLQ logic and populate SensorStateTracker.RecordRejected
// counters, making them visible in GetSensorDiagnostic output.
package generator

import (
"math/rand"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"google.golang.org/protobuf/types/known/timestamppb"
)

// InvalidInjectionConfig controls what fraction of observations are corrupted
// and how corruption is distributed across reason categories.
type InvalidInjectionConfig struct {
// Enabled turns the injector on or off without changing other config.
Enabled bool
// Rate is the probability (0.0–1.0) that any given observation is corrupted.
Rate float64
// Reasons maps reason label → relative weight. Weights are normalised
// internally so they do not need to sum to 1.0.
Reasons map[string]float64
}

// InvalidObservationInjector decides whether to corrupt an observation and,
// if so, applies one of the configured corruption strategies.
type InvalidObservationInjector struct {
cfg      InvalidInjectionConfig
rng      *rand.Rand
buckets  []reasonBucket // pre-computed cumulative distribution
}

type reasonBucket struct {
reason    string
cumWeight float64
}

// NewInvalidObservationInjector constructs a ready-to-use injector.
// If cfg.Enabled is false or cfg.Rate == 0, MaybeCorrupt always returns
// the observation unchanged.
func NewInvalidObservationInjector(cfg InvalidInjectionConfig, rng *rand.Rand) *InvalidObservationInjector {
inj := &InvalidObservationInjector{cfg: cfg, rng: rng}
inj.buildBuckets()
return inj
}

// buildBuckets pre-computes a cumulative distribution from the reason weights.
func (inj *InvalidObservationInjector) buildBuckets() {
var total float64
for _, w := range inj.cfg.Reasons {
total += w
}
if total == 0 {
return
}
cumulative := 0.0
for reason, w := range inj.cfg.Reasons {
cumulative += w / total
inj.buckets = append(inj.buckets, reasonBucket{reason: reason, cumWeight: cumulative})
}
}

// MaybeCorrupt returns either the original observation or a corrupted clone,
// depending on the configured rate. The original is never modified in-place.
func (inj *InvalidObservationInjector) MaybeCorrupt(obs *ingestionv1.SensorObservation) *ingestionv1.SensorObservation {
if !inj.cfg.Enabled || inj.cfg.Rate <= 0 || len(inj.buckets) == 0 {
return obs
}
if inj.rng.Float64() >= inj.cfg.Rate {
return obs
}
reason := inj.pickReason()
return inj.corrupt(obs, reason)
}

// pickReason selects a corruption reason using the pre-computed distribution.
func (inj *InvalidObservationInjector) pickReason() string {
roll := inj.rng.Float64()
for _, b := range inj.buckets {
if roll <= b.cumWeight {
return b.reason
}
}
return inj.buckets[len(inj.buckets)-1].reason
}

// corrupt returns a shallow clone of obs with the chosen field corrupted.
func (inj *InvalidObservationInjector) corrupt(obs *ingestionv1.SensorObservation, reason string) *ingestionv1.SensorObservation {
// Shallow clone — only the fields we corrupt are overridden.
clone := *obs
switch reason {
case "invalid_timestamp":
// Set observation time 10 years in the future to fail timestamp validation.
future := time.Now().Add(10 * 365 * 24 * time.Hour)
clone.ObservationTime = timestamppb.New(future)
case "coordinates_out_of_range":
// Set position to clearly invalid coordinates (valid range: lat -90..90, lon -180..180).
clone.Position = &commonv1.Position{
Latitude:  999.0,
Longitude: 999.0,
}
case "missing_sensor_id":
clone.SensorId = ""
case "schema_mismatch":
// Set an invalid (out-of-enum) classification level to trigger schema checks.
clone.Classification = commonv1.ClassificationLevel(99)
default:
// Unknown reason: blank the sensor_id as a generic corruption.
clone.SensorId = ""
}
return &clone
}
