// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"math"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/state"
)

// Position is a geographic coordinate (WGS-84).
type Position struct {
Latitude  float64
Longitude float64
}

// ExclusionZone defines a geographic area to monitor for proximity alerts.
type ExclusionZone struct {
Name      string
CenterLat float64
CenterLon float64
RadiusNM  float64
}

// FeatureVector contains all extracted features for anomaly detection.
type FeatureVector struct {
// ── Speed Features ──
CurrentSpeedKnots float64
AvgSpeed30Min     float64 // 30-minute rolling average
SpeedStdDev       float64 // Standard deviation of speed over history
SpeedDeltaSigma   float64 // (current - avg) / stddev

// ── Route Features ──
CurrentHeading   float64
HeadingChangeRate float64 // degrees/minute over last 5 updates
ExpectedHeading  float64 // From linear regression of last 10 positions
HeadingDeviation float64 // |current - expected|

// ── AIS Features (surface only) ──
AISReportedPosition *Position // From AIS sensor
FusedPosition       *Position // From fusion engine
AISPositionDeltaNM  float64   // Haversine distance between AIS and fused
HasAISSource        bool

// ── Behavioral Features ──
ActivityPattern   []float64 // Encoded activity sequence (last 20 states)
PatternConfidence float64   // How anomalous the pattern is (0-1)

// ── Temporal Features ──
HourOfDay     int
DayOfWeek     int
IsNighttime   bool
TemporalPValue float64 // p-value for time-of-day activity

// ── Proximity Features ──
NearestExclusionZoneDistNM float64 // Distance to nearest exclusion zone
InExclusionZone            bool
NearestCriticalAssetDistNM float64
NearestZoneName            string
NearestZoneRadiusNM        float64

// ── Track Metadata ──
TrackID     string
EntityType  commonv1.EntityType
SourceCount uint32
TrackAgeMin float64
}

// FeatureExtractor builds FeatureVectors from track state and history.
type FeatureExtractor struct {
history        *state.TrackHistory
exclusionZones []ExclusionZone
}

// NewFeatureExtractor creates a new FeatureExtractor.
func NewFeatureExtractor(history *state.TrackHistory, zones []ExclusionZone) *FeatureExtractor {
return &FeatureExtractor{
history:        history,
exclusionZones: zones,
}
}

// Extract builds a FeatureVector for the given fused track update.
func (fe *FeatureExtractor) Extract(track *entityv1.FusedTrack) (*FeatureVector, error) {
if track == nil {
return nil, fmt.Errorf("[domain.FeatureExtractor.Extract]: track is nil")
}

fv := &FeatureVector{
TrackID:     track.GetTrackId(),
EntityType:  track.GetEntityType(),
SourceCount: track.GetSourceCount(),
TrackAgeMin: track.GetAgeSeconds() / 60.0,
}

// Set fused position.
if pos := track.GetEstimatedPosition(); pos != nil {
fv.FusedPosition = &Position{
Latitude:  pos.GetLatitude(),
Longitude: pos.GetLongitude(),
}
}

// Set velocity/speed features from position (preferred) or NED velocity vector.
if pos := track.GetEstimatedPosition(); pos != nil {
fv.CurrentSpeedKnots = pos.GetSpeedKnots()
fv.CurrentHeading = pos.GetHeadingDegrees()
}
// Fall back to computing from NED velocity components if position has no speed.
if fv.CurrentSpeedKnots == 0 {
if vel := track.GetVelocity(); vel != nil {
speedMPS := math.Sqrt(vel.GetNorthMps()*vel.GetNorthMps() + vel.GetEastMps()*vel.GetEastMps())
fv.CurrentSpeedKnots = speedMPS * 1.94384
if fv.CurrentHeading == 0 && (vel.GetNorthMps() != 0 || vel.GetEastMps() != 0) {
hdg := math.Atan2(vel.GetEastMps(), vel.GetNorthMps()) * 180.0 / math.Pi
if hdg < 0 {
hdg += 360
}
fv.CurrentHeading = hdg
}
}
}

// Extract historical speed statistics (30-minute window).
fv.AvgSpeed30Min = fe.history.AvgSpeed(track.GetTrackId(), 30*time.Minute)
fv.SpeedStdDev = fe.history.SpeedStdDev(track.GetTrackId(), 30*time.Minute)
if fv.SpeedStdDev > 0.1 {
fv.SpeedDeltaSigma = math.Abs(fv.CurrentSpeedKnots-fv.AvgSpeed30Min) / fv.SpeedStdDev
}

// Extract heading features from recent history.
headings := fe.history.RecentHeadings(track.GetTrackId(), 10)
fv.ExpectedHeading = fe.computeExpectedHeading(headings, fv.CurrentHeading)
fv.HeadingDeviation = angularDifference(fv.CurrentHeading, fv.ExpectedHeading)
fv.HeadingChangeRate = fe.computeHeadingChangeRate(headings, track.GetTrackId())

// AIS source detection (surface tracks only).
if track.GetEntityType() == commonv1.EntityType_ENTITY_TYPE_SURFACE {
fv.HasAISSource, fv.AISReportedPosition = fe.extractAISPosition(track)
if fv.HasAISSource && fv.AISReportedPosition != nil && fv.FusedPosition != nil {
fv.AISPositionDeltaNM = haversineNM(
fv.AISReportedPosition.Latitude, fv.AISReportedPosition.Longitude,
fv.FusedPosition.Latitude, fv.FusedPosition.Longitude,
)
}
}

// Behavioral features from history.
recentEntries := fe.history.GetHistory(track.GetTrackId(), 30*time.Minute)
fv.ActivityPattern = fe.encodeActivityPattern(recentEntries)
fv.PatternConfidence = fe.computePatternConfidence(recentEntries, fv.FusedPosition)

// Temporal features.
now := time.Now().UTC()
fv.HourOfDay = now.Hour()
fv.DayOfWeek = int(now.Weekday())
fv.IsNighttime = now.Hour() < 6 || now.Hour() >= 22
fv.TemporalPValue = fe.computeTemporalPValue(track.GetTrackId(), now)

// Proximity features.
fe.computeProximityFeatures(fv)

// Append current state to history for future extractions.
fe.history.Append(track.GetTrackId(), &state.HistoryEntry{
Timestamp:  time.Now(),
Latitude:   fv.FusedPosition.latOrZero(),
Longitude:  fv.FusedPosition.lonOrZero(),
SpeedKnots: fv.CurrentSpeedKnots,
Heading:    fv.CurrentHeading,
EntityType: fv.EntityType,
})

return fv, nil
}

// computeExpectedHeading uses a simple weighted average of recent headings.
func (fe *FeatureExtractor) computeExpectedHeading(headings []float64, current float64) float64 {
if len(headings) == 0 {
return current
}
// Simple average with circular mean for headings.
var sinSum, cosSum float64
for _, h := range headings {
rad := h * math.Pi / 180
sinSum += math.Sin(rad)
cosSum += math.Cos(rad)
}
avg := math.Atan2(sinSum, cosSum) * 180 / math.Pi
if avg < 0 {
avg += 360
}
return avg
}

// computeHeadingChangeRate calculates degrees/minute from recent headings.
func (fe *FeatureExtractor) computeHeadingChangeRate(headings []float64, _ string) float64 {
if len(headings) < 2 {
return 0
}
// Sum of heading changes across last entries, normalised to 1 minute.
totalChange := 0.0
for i := 1; i < len(headings); i++ {
totalChange += math.Abs(angularDifference(headings[i], headings[i-1]))
}
// Assume 1 minute intervals between updates as a rough estimate.
return totalChange / float64(len(headings)-1)
}

// extractAISPosition detects whether an AIS sensor is among the track sources.
// For MVP, we detect the presence of AIS but cannot recover the original AIS-reported
// position from the FusedTrack proto (SourceAttribution has no position field).
// The AISPositionDeltaNM is therefore only populated when it can be inferred from
// context (e.g., set externally). AIS manipulation detection checks HasAISSource.
func (fe *FeatureExtractor) extractAISPosition(track *entityv1.FusedTrack) (bool, *Position) {
for _, src := range track.GetSources() {
if src.GetSensorType() == commonv1.SensorType_SENSOR_TYPE_AIS_BFT {
// AIS source present — return true with nil position (delta computed externally).
return true, nil
}
}
return false, nil
}

// encodeActivityPattern creates a simple float64 encoding of recent speed values.
func (fe *FeatureExtractor) encodeActivityPattern(entries []*state.HistoryEntry) []float64 {
pattern := make([]float64, 0, len(entries))
for _, e := range entries {
pattern = append(pattern, e.SpeedKnots)
}
// Pad to 20 with zeros if needed.
for len(pattern) < 20 {
pattern = append(pattern, 0)
}
if len(pattern) > 20 {
pattern = pattern[len(pattern)-20:]
}
return pattern
}

// computePatternConfidence detects behaviorally anomalous patterns.
// MVP: rule-based detection for loitering, zigzag, speed pulsing.
func (fe *FeatureExtractor) computePatternConfidence(entries []*state.HistoryEntry, currentPos *Position) float64 {
if len(entries) < 3 || currentPos == nil {
return 0
}

confidence := 0.0

// Detect loitering: check if positions are within 0.1 NM over > 30 min.
if len(entries) >= 2 {
duration := entries[len(entries)-1].Timestamp.Sub(entries[0].Timestamp)
if duration > 30*time.Minute {
maxDist := 0.0
for _, e := range entries {
d := haversineNM(e.Latitude, e.Longitude, currentPos.Latitude, currentPos.Longitude)
if d > maxDist {
maxDist = d
}
}
if maxDist < 0.1 {
confidence = math.Max(confidence, 0.85)
}
}
}

// Detect zigzag: heading reversals > 5 times in recent history.
if len(entries) >= 6 {
reversals := 0
for i := 2; i < len(entries); i++ {
prev := entries[i-1].Heading - entries[i-2].Heading
curr := entries[i].Heading - entries[i-1].Heading
if (prev > 30 && curr < -30) || (prev < -30 && curr > 30) {
reversals++
}
}
if reversals > 5 {
confidence = math.Max(confidence, 0.80)
}
}

// Detect speed pulsing: alternating fast/slow > 4 times.
if len(entries) >= 5 {
pulsing := 0
meanSpeed := 0.0
for _, e := range entries {
meanSpeed += e.SpeedKnots
}
meanSpeed /= float64(len(entries))
prevFast := entries[0].SpeedKnots > meanSpeed
for _, e := range entries[1:] {
currFast := e.SpeedKnots > meanSpeed
if currFast != prevFast {
pulsing++
}
prevFast = currFast
}
if pulsing > 4 {
confidence = math.Max(confidence, 0.78)
}
}

return confidence
}

// computeTemporalPValue returns a simulated p-value for time-of-day activity.
// MVP: returns low p-value for activity in early morning hours (0–5 UTC).
func (fe *FeatureExtractor) computeTemporalPValue(_ string, t time.Time) float64 {
hour := t.Hour()
// Simulate statistical test: early morning activity has low p-value.
switch {
case hour >= 0 && hour < 4:
return 0.01
case hour >= 4 && hour < 6:
return 0.03
default:
return 0.15
}
}

// computeProximityFeatures calculates distance to nearest exclusion zone.
func (fe *FeatureExtractor) computeProximityFeatures(fv *FeatureVector) {
if fv.FusedPosition == nil || len(fe.exclusionZones) == 0 {
fv.NearestExclusionZoneDistNM = math.MaxFloat64
return
}

minDist := math.MaxFloat64
fv.NearestExclusionZoneDistNM = math.MaxFloat64

for _, zone := range fe.exclusionZones {
dist := haversineNM(
fv.FusedPosition.Latitude, fv.FusedPosition.Longitude,
zone.CenterLat, zone.CenterLon,
)
distFromEdge := dist - zone.RadiusNM
if distFromEdge < minDist {
minDist = distFromEdge
fv.NearestExclusionZoneDistNM = math.Max(0, distFromEdge)
fv.NearestZoneName = zone.Name
fv.NearestZoneRadiusNM = zone.RadiusNM
if distFromEdge <= 0 {
fv.InExclusionZone = true
}
}
}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// haversineNM computes the great-circle distance in nautical miles.
func haversineNM(lat1, lon1, lat2, lon2 float64) float64 {
const earthRadiusNM = 3440.065
φ1 := lat1 * math.Pi / 180
φ2 := lat2 * math.Pi / 180
Δφ := (lat2 - lat1) * math.Pi / 180
Δλ := (lon2 - lon1) * math.Pi / 180

a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
return earthRadiusNM * c
}

// angularDifference returns the signed difference between two headings,
// normalised to [-180, +180].
func angularDifference(a, b float64) float64 {
diff := a - b
for diff > 180 {
diff -= 360
}
for diff < -180 {
diff += 360
}
return diff
}

// latOrZero returns the latitude or 0 if pos is nil.
func (p *Position) latOrZero() float64 {
if p == nil {
return 0
}
return p.Latitude
}

// lonOrZero returns the longitude or 0 if pos is nil.
func (p *Position) lonOrZero() float64 {
if p == nil {
return 0
}
return p.Longitude
}

// HaversineNM is exported for use in tests.
var HaversineNM = haversineNM
