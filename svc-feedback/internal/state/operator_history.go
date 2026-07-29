// CLASSIFICATION: UNCLASSIFIED

// Package state provides thread-safe in-memory tracking of operator
// feedback history for trust scoring and anti-poisoning analysis.
package state

import (
"sync"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
)

// FeedbackEntry records a single feedback submission.
type FeedbackEntry struct {
FeedbackID   string
TrackID      string
FeedbackType commonv1.FeedbackType
Timestamp    time.Time
TrustScore   float64
LabelFlipped bool   // true if contradicts previous feedback on same track
SensorSource string // sensor source associated with this track
Validated    bool   // true if independently validated (trust_score >= 0.5)
}

// OperatorStats tracks an operator's accumulated feedback history.
type OperatorStats struct {
OperatorID         string
TotalFeedback      int
ConfirmedCorrect   int                     // count of validated submissions
LabelFlipCount     int                     // count of label-flipped submissions
FeedbackByType     map[string]int          // FeedbackType.String() → count
FeedbackByTrack    map[string][]FeedbackEntry
RecentFeedback     []time.Time             // timestamps of all submissions
TrackSensorSources map[string]bool         // unique sensor source identifiers
}

// TypeDistribution returns the normalised distribution of feedback types.
func (os *OperatorStats) TypeDistribution() map[string]float64 {
dist := make(map[string]float64, len(os.FeedbackByType))
total := os.TotalFeedback
if total == 0 {
return dist
}
for k, v := range os.FeedbackByType {
dist[k] = float64(v) / float64(total)
}
return dist
}

// LabelFlipRate returns the proportion of label-flipped feedback.
func (os *OperatorStats) LabelFlipRate() float64 {
if os.TotalFeedback == 0 {
return 0.0
}
return float64(os.LabelFlipCount) / float64(os.TotalFeedback)
}

// UniqueSensorSources returns the count of unique sensor sources.
func (os *OperatorStats) UniqueSensorSources() int {
return len(os.TrackSensorSources)
}

// ValidatedRatio returns the ratio of validated submissions to total.
func (os *OperatorStats) ValidatedRatio() float64 {
if os.TotalFeedback == 0 {
return 0.0
}
return float64(os.ConfirmedCorrect) / float64(os.TotalFeedback)
}

// OperatorHistory maintains per-operator feedback statistics and a full
// proto log used to serve GetFeedbackHistory queries.
// All exported methods are safe for concurrent use.
type OperatorHistory struct {
mu          sync.RWMutex
operators   map[string]*OperatorStats
allFeedback []*feedbackv1.OperatorFeedback
}

// NewOperatorHistory returns an initialised, empty OperatorHistory.
func NewOperatorHistory() *OperatorHistory {
return &OperatorHistory{
operators:   make(map[string]*OperatorStats),
allFeedback: make([]*feedbackv1.OperatorFeedback, 0, 64),
}
}

// RecordFeedback records a feedback entry for the given operator.
// It auto-detects label flips by comparing against the previous
// feedback type submitted for the same track.
func (oh *OperatorHistory) RecordFeedback(operatorID string, entry FeedbackEntry) {
oh.mu.Lock()
defer oh.mu.Unlock()

stats := oh.getOrCreateLocked(operatorID)

// Detect label flip: check if operator previously submitted a different
// feedback type for this same track.
if prevEntries := stats.FeedbackByTrack[entry.TrackID]; len(prevEntries) > 0 {
lastType := prevEntries[len(prevEntries)-1].FeedbackType
if lastType != entry.FeedbackType {
entry.LabelFlipped = true
stats.LabelFlipCount++
}
}

stats.TotalFeedback++
typeKey := entry.FeedbackType.String()
stats.FeedbackByType[typeKey]++
stats.FeedbackByTrack[entry.TrackID] = append(stats.FeedbackByTrack[entry.TrackID], entry)
stats.RecentFeedback = append(stats.RecentFeedback, entry.Timestamp)

if entry.SensorSource != "" {
stats.TrackSensorSources[entry.SensorSource] = true
}
if entry.Validated {
stats.ConfirmedCorrect++
}
}

// RecordProto stores a complete OperatorFeedback proto for history queries.
// This is called after RecordFeedback so the proto persists independently.
func (oh *OperatorHistory) RecordProto(fb *feedbackv1.OperatorFeedback) {
oh.mu.Lock()
defer oh.mu.Unlock()
oh.allFeedback = append(oh.allFeedback, fb)
}

// GetStats returns a shallow copy of operator statistics, or nil if unknown.
func (oh *OperatorHistory) GetStats(operatorID string) *OperatorStats {
oh.mu.RLock()
defer oh.mu.RUnlock()
s, ok := oh.operators[operatorID]
if !ok {
return nil
}
cp := *s
return &cp
}

// GetFeedbackByTrack returns all FeedbackEntry records for a track ID,
// across all operators, for deviation scoring.
func (oh *OperatorHistory) GetFeedbackByTrack(trackID string) []FeedbackEntry {
oh.mu.RLock()
defer oh.mu.RUnlock()
var results []FeedbackEntry
for _, stats := range oh.operators {
if entries, ok := stats.FeedbackByTrack[trackID]; ok {
results = append(results, entries...)
}
}
return results
}

// QueryHistory returns all stored OperatorFeedback protos, optionally
// filtered by operatorID and/or trackID (empty string means no filter).
func (oh *OperatorHistory) QueryHistory(operatorID, trackID string) []*feedbackv1.OperatorFeedback {
oh.mu.RLock()
defer oh.mu.RUnlock()

results := make([]*feedbackv1.OperatorFeedback, 0, len(oh.allFeedback))
for _, fb := range oh.allFeedback {
if operatorID != "" && fb.GetOperatorId() != operatorID {
continue
}
if trackID != "" && fb.GetTrackId() != trackID {
continue
}
results = append(results, fb)
}
return results
}

// getOrCreateLocked returns or initialises OperatorStats for the given ID.
// Caller MUST hold oh.mu write lock.
func (oh *OperatorHistory) getOrCreateLocked(operatorID string) *OperatorStats {
s, ok := oh.operators[operatorID]
if !ok {
s = &OperatorStats{
OperatorID:         operatorID,
FeedbackByType:     make(map[string]int),
FeedbackByTrack:    make(map[string][]FeedbackEntry),
RecentFeedback:     make([]time.Time, 0, 16),
TrackSensorSources: make(map[string]bool),
}
oh.operators[operatorID] = s
}
return s
}
