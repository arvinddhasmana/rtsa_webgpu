// CLASSIFICATION: UNCLASSIFIED
package state

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
)

func newEntry(speedKnots float64, heading float64, ageAgo time.Duration) *HistoryEntry {
return &HistoryEntry{
Timestamp:  time.Now().Add(-ageAgo),
Latitude:   44.0,
Longitude:  -63.0,
SpeedKnots: speedKnots,
Heading:    heading,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
}
}

func TestTrackHistory_AvgSpeed(t *testing.T) {
th := NewTrackHistory(100, 2*time.Hour)

// T23: Test correct average speed calculation.
th.Append("track-1", newEntry(10.0, 90.0, 20*time.Minute))
th.Append("track-1", newEntry(20.0, 90.0, 15*time.Minute))
th.Append("track-1", newEntry(30.0, 90.0, 10*time.Minute))

avg := th.AvgSpeed("track-1", 30*time.Minute)
want := 20.0
if avg != want {
t.Errorf("AvgSpeed = %v, want %v", avg, want)
}
}

func TestTrackHistory_AvgSpeed_NoHistory(t *testing.T) {
th := NewTrackHistory(100, 2*time.Hour)
avg := th.AvgSpeed("unknown-track", 30*time.Minute)
if avg != 0 {
t.Errorf("AvgSpeed for unknown track = %v, want 0", avg)
}
}

func TestTrackHistory_SpeedStdDev(t *testing.T) {
th := NewTrackHistory(100, 2*time.Hour)
th.Append("track-2", newEntry(10.0, 90.0, 25*time.Minute))
th.Append("track-2", newEntry(20.0, 90.0, 20*time.Minute))
th.Append("track-2", newEntry(30.0, 90.0, 15*time.Minute))

stddev := th.SpeedStdDev("track-2", 30*time.Minute)
if stddev <= 0 {
t.Errorf("SpeedStdDev = %v, want > 0", stddev)
}
}

func TestTrackHistory_SpeedStdDev_InsufficientHistory(t *testing.T) {
th := NewTrackHistory(100, 2*time.Hour)
th.Append("track-3", newEntry(10.0, 90.0, 5*time.Minute))

stddev := th.SpeedStdDev("track-3", 30*time.Minute)
if stddev != 0 {
t.Errorf("SpeedStdDev with 1 entry = %v, want 0", stddev)
}
}

func TestTrackHistory_RecentHeadings(t *testing.T) {
th := NewTrackHistory(100, 2*time.Hour)
th.Append("track-4", newEntry(10.0, 90.0, 20*time.Minute))
th.Append("track-4", newEntry(10.0, 100.0, 15*time.Minute))
th.Append("track-4", newEntry(10.0, 110.0, 10*time.Minute))
th.Append("track-4", newEntry(10.0, 120.0, 5*time.Minute))

headings := th.RecentHeadings("track-4", 3)
if len(headings) != 3 {
t.Fatalf("RecentHeadings returned %d entries, want 3", len(headings))
}
// Should be the last 3: 100, 110, 120.
want := []float64{100.0, 110.0, 120.0}
for i, h := range headings {
if h != want[i] {
t.Errorf("heading[%d] = %v, want %v", i, h, want[i])
}
}
}

func TestTrackHistory_Cleanup(t *testing.T) {
// T24: Old entries should be removed.
th := NewTrackHistory(100, 1*time.Hour)

// Add old entry (over 1 hour ago - should be cleaned).
th.Append("old-track", &HistoryEntry{
Timestamp:  time.Now().Add(-2 * time.Hour),
SpeedKnots: 10.0,
Heading:    90.0,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
})

// Add fresh entry.
th.Append("fresh-track", newEntry(10.0, 90.0, 5*time.Minute))

th.Cleanup()

count := th.Count()
if count != 1 {
t.Errorf("After cleanup, Count = %d, want 1 (old-track removed)", count)
}

// Verify old-track is gone.
history := th.GetHistory("old-track", 3*time.Hour)
if len(history) != 0 {
t.Errorf("Old track should have 0 entries after cleanup, got %d", len(history))
}
}

func TestTrackHistory_CircularBuffer_Overflow(t *testing.T) {
th := NewTrackHistory(3, 24*time.Hour) // tiny buffer

th.Append("track-5", newEntry(10.0, 90.0, 50*time.Minute))
th.Append("track-5", newEntry(20.0, 100.0, 40*time.Minute))
th.Append("track-5", newEntry(30.0, 110.0, 30*time.Minute))
// This should overwrite the oldest.
th.Append("track-5", newEntry(40.0, 120.0, 20*time.Minute))

entries := th.GetHistory("track-5", 24*time.Hour)
if len(entries) != 3 {
t.Fatalf("Expected 3 entries (buffer full), got %d", len(entries))
}
// Oldest (10.0) should be gone, newest should be 40.0.
lastEntry := entries[len(entries)-1]
if lastEntry.SpeedKnots != 40.0 {
t.Errorf("Last entry SpeedKnots = %v, want 40.0", lastEntry.SpeedKnots)
}
}

func TestTrackHistory_GetHistory_WindowFilter(t *testing.T) {
th := NewTrackHistory(100, 24*time.Hour)
th.Append("track-6", newEntry(10.0, 90.0, 60*time.Minute))
th.Append("track-6", newEntry(20.0, 90.0, 15*time.Minute))

// Only get entries within 30 minutes.
entries := th.GetHistory("track-6", 30*time.Minute)
if len(entries) != 1 {
t.Errorf("Expected 1 entry within 30min window, got %d", len(entries))
}
if entries[0].SpeedKnots != 20.0 {
t.Errorf("Entry SpeedKnots = %v, want 20.0", entries[0].SpeedKnots)
}
}

func TestTrackHistory_TrackAge(t *testing.T) {
th := NewTrackHistory(100, 24*time.Hour)
th.Append("track-7", newEntry(10.0, 90.0, 45*time.Minute))
th.Append("track-7", newEntry(20.0, 90.0, 5*time.Minute))

age := th.TrackAge("track-7")
if age < 44*time.Minute || age > 46*time.Minute {
t.Errorf("TrackAge = %v, want ~45 min", age)
}
}

func TestTrackHistory_TrackAge_NoHistory(t *testing.T) {
th := NewTrackHistory(100, 24*time.Hour)
age := th.TrackAge("nonexistent")
if age != 0 {
t.Errorf("TrackAge for nonexistent = %v, want 0", age)
}
}
