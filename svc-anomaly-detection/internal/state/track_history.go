// CLASSIFICATION: UNCLASSIFIED
package state

import (
"math"
"sync"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

// HistoryEntry is a snapshot of a track at a point in time.
type HistoryEntry struct {
Timestamp  time.Time
Latitude   float64
Longitude  float64
SpeedKnots float64
Heading    float64
EntityType commonv1.EntityType
}

// circularBuffer is a fixed-size circular buffer of HistoryEntry.
type circularBuffer struct {
entries []*HistoryEntry
head    int
size    int
cap     int
}

func newCircularBuffer(capacity int) *circularBuffer {
return &circularBuffer{
entries: make([]*HistoryEntry, capacity),
cap:     capacity,
}
}

// push adds an entry to the buffer, overwriting the oldest if full.
func (b *circularBuffer) push(e *HistoryEntry) {
b.entries[b.head] = e
b.head = (b.head + 1) % b.cap
if b.size < b.cap {
b.size++
}
}

// all returns all entries in chronological order (oldest to newest).
func (b *circularBuffer) all() []*HistoryEntry {
out := make([]*HistoryEntry, 0, b.size)
if b.size < b.cap {
// Buffer not yet full: entries[0..size-1] are in order.
for i := 0; i < b.size; i++ {
out = append(out, b.entries[i])
}
return out
}
// Buffer full: head points to oldest entry.
for i := 0; i < b.cap; i++ {
idx := (b.head + i) % b.cap
out = append(out, b.entries[idx])
}
return out
}

// TrackHistory maintains a sliding window of recent track states
// needed for feature extraction (speed averages, heading trends, etc.).
type TrackHistory struct {
mu         sync.RWMutex
history    map[string]*circularBuffer
maxEntries int
maxAge     time.Duration
}

// NewTrackHistory creates a new TrackHistory store.
func NewTrackHistory(maxEntries int, maxAge time.Duration) *TrackHistory {
return &TrackHistory{
history:    make(map[string]*circularBuffer),
maxEntries: maxEntries,
maxAge:     maxAge,
}
}

// Append adds a new entry to a track's history.
func (th *TrackHistory) Append(trackID string, entry *HistoryEntry) {
th.mu.Lock()
defer th.mu.Unlock()

buf, ok := th.history[trackID]
if !ok {
buf = newCircularBuffer(th.maxEntries)
th.history[trackID] = buf
}
buf.push(entry)
}

// GetHistory returns the history entries for a track within the given time window.
func (th *TrackHistory) GetHistory(trackID string, window time.Duration) []*HistoryEntry {
th.mu.RLock()
defer th.mu.RUnlock()

buf, ok := th.history[trackID]
if !ok {
return nil
}

cutoff := time.Now().Add(-window)
all := buf.all()
out := make([]*HistoryEntry, 0, len(all))
for _, e := range all {
if e.Timestamp.After(cutoff) {
out = append(out, e)
}
}
return out
}

// AvgSpeed calculates the average speed over the given window.
// Returns 0 if no history entries exist.
func (th *TrackHistory) AvgSpeed(trackID string, window time.Duration) float64 {
entries := th.GetHistory(trackID, window)
if len(entries) == 0 {
return 0
}
var sum float64
for _, e := range entries {
sum += e.SpeedKnots
}
return sum / float64(len(entries))
}

// SpeedStdDev calculates the standard deviation of speed over the given window.
// Returns 0 if fewer than 2 entries.
func (th *TrackHistory) SpeedStdDev(trackID string, window time.Duration) float64 {
entries := th.GetHistory(trackID, window)
if len(entries) < 2 {
return 0
}
var sum float64
for _, e := range entries {
sum += e.SpeedKnots
}
mean := sum / float64(len(entries))

var variance float64
for _, e := range entries {
diff := e.SpeedKnots - mean
variance += diff * diff
}
variance /= float64(len(entries) - 1) // sample std dev
return math.Sqrt(variance)
}

// RecentHeadings returns the last N headings for heading change analysis.
func (th *TrackHistory) RecentHeadings(trackID string, n int) []float64 {
th.mu.RLock()
defer th.mu.RUnlock()

buf, ok := th.history[trackID]
if !ok {
return nil
}

all := buf.all()
if len(all) == 0 {
return nil
}

start := 0
if len(all) > n {
start = len(all) - n
}

out := make([]float64, len(all)-start)
for i, e := range all[start:] {
out[i] = e.Heading
}
return out
}

// TrackAge returns the duration since the earliest history entry for a track.
// Returns 0 if no history exists.
func (th *TrackHistory) TrackAge(trackID string) time.Duration {
th.mu.RLock()
defer th.mu.RUnlock()

buf, ok := th.history[trackID]
if !ok {
return 0
}
all := buf.all()
if len(all) == 0 {
return 0
}
return time.Since(all[0].Timestamp)
}

// Cleanup removes entries older than maxAge across all tracks.
// Tracks with no remaining entries are removed from the map.
func (th *TrackHistory) Cleanup() {
th.mu.Lock()
defer th.mu.Unlock()

cutoff := time.Now().Add(-th.maxAge)
for trackID, buf := range th.history {
all := buf.all()
// Check if all entries are expired.
allExpired := true
for _, e := range all {
if e.Timestamp.After(cutoff) {
allExpired = false
break
}
}
if allExpired && len(all) > 0 {
delete(th.history, trackID)
}
}
}

// Count returns the number of tracks in history.
func (th *TrackHistory) Count() int {
th.mu.RLock()
defer th.mu.RUnlock()
return len(th.history)
}
