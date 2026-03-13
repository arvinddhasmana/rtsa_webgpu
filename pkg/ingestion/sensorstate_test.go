// CLASSIFICATION: UNCLASSIFIED
// pkg/ingestion/sensorstate_test.go — Unit tests for SensorStateTracker
//
// Run with: go test -race ./pkg/ingestion/...

package ingestion

import (
"context"
"sync"
"testing"
"time"
)

func TestRecordAccepted_UpdatesCounters(t *testing.T) {
tracker := NewSensorStateTracker()

tracker.RecordAccepted(1_000_000) // 1ms

if tracker.TotalReceived() != 1 {
t.Errorf("TotalReceived: got %d, want 1", tracker.TotalReceived())
}
if tracker.TotalAccepted() != 1 {
t.Errorf("TotalAccepted: got %d, want 1", tracker.TotalAccepted())
}
if tracker.TotalRejected() != 0 {
t.Errorf("TotalRejected: got %d, want 0", tracker.TotalRejected())
}

// LastObsTime should be set.
if tracker.LastObsTime().IsZero() {
t.Error("LastObsTime should be set after RecordAccepted")
}
// Connected should be true immediately after recording.
if !tracker.Connected() {
t.Error("Connected should be true immediately after RecordAccepted")
}
}

func TestRecordRejected_IncrementsDLQReason(t *testing.T) {
tracker := NewSensorStateTracker()

tracker.RecordRejected("invalid_timestamp")
tracker.RecordRejected("invalid_timestamp")
tracker.RecordRejected("coordinates_out_of_range")

if tracker.TotalReceived() != 3 {
t.Errorf("TotalReceived: got %d, want 3", tracker.TotalReceived())
}
if tracker.TotalRejected() != 3 {
t.Errorf("TotalRejected: got %d, want 3", tracker.TotalRejected())
}
if tracker.TotalAccepted() != 0 {
t.Errorf("TotalAccepted: got %d, want 0", tracker.TotalAccepted())
}

breakdown := tracker.DLQBreakdown()
if len(breakdown) != 2 {
t.Errorf("DLQBreakdown len: got %d, want 2", len(breakdown))
}

counts := make(map[string]int64, len(breakdown))
for _, d := range breakdown {
counts[d.GetReason()] = d.GetCount()
}
if counts["invalid_timestamp"] != 2 {
t.Errorf("invalid_timestamp count: got %d, want 2", counts["invalid_timestamp"])
}
if counts["coordinates_out_of_range"] != 1 {
t.Errorf("coordinates_out_of_range count: got %d, want 1", counts["coordinates_out_of_range"])
}
}

func TestSnapshotThroughput_RespectsLimit(t *testing.T) {
tracker := NewSensorStateTracker()
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Inject some accepted observations then manually add throughput samples.
for i := 0; i < 5; i++ {
tracker.RecordAccepted(500_000)
tracker.throughputMu.Lock()
tracker.throughputRing[tracker.throughputHead] = throughputEntry{
sampledAt:       time.Now().UTC(),
eventsPerSecond: float64(i + 1),
}
tracker.throughputHead = (tracker.throughputHead + 1) % throughputCap
tracker.throughputMu.Unlock()
}
_ = ctx

samples := tracker.SnapshotThroughput(3)
if len(samples) != 3 {
t.Errorf("SnapshotThroughput(3): got %d samples, want 3", len(samples))
}
}

func TestSnapshotEvents_RespectsLimit(t *testing.T) {
tracker := NewSensorStateTracker()

// Generate 10 accepted observations → 10 INFO events.
for i := 0; i < 10; i++ {
tracker.RecordAccepted(1_000_000)
}

events := tracker.SnapshotEvents(5)
if len(events) != 5 {
t.Errorf("SnapshotEvents(5): got %d events, want 5", len(events))
}
}

func TestDLQBreakdown_ReturnsAllReasons(t *testing.T) {
tracker := NewSensorStateTracker()

reasons := []string{"reason_a", "reason_b", "reason_c"}
for _, r := range reasons {
tracker.RecordRejected(r)
}

breakdown := tracker.DLQBreakdown()
if len(breakdown) != 3 {
t.Errorf("DLQBreakdown: got %d entries, want 3", len(breakdown))
}
}

func TestValidationPassRate_Correct(t *testing.T) {
tracker := NewSensorStateTracker()

// 10 accepted, 5 rejected → 10/15 ≈ 66.67%.
for i := 0; i < 10; i++ {
tracker.RecordAccepted(1_000_000)
}
for i := 0; i < 5; i++ {
tracker.RecordRejected("bad_data")
}

rate := tracker.ValidationPassRate()
// Allow ±0.1 tolerance.
const want = 66.67
if rate < want-0.1 || rate > want+0.1 {
t.Errorf("ValidationPassRate: got %.2f, want %.2f ± 0.1", rate, want)
}
}

func TestConcurrentAccess_NoRaceCondition(t *testing.T) {
tracker := NewSensorStateTracker()
const goroutines = 50
const opsPerGoroutine = 200

var wg sync.WaitGroup
wg.Add(goroutines)
for g := 0; g < goroutines; g++ {
g := g
go func() {
defer wg.Done()
for i := 0; i < opsPerGoroutine; i++ {
if (g+i)%3 == 0 {
tracker.RecordRejected("reason_" + string(rune('a'+g%5)))
} else {
tracker.RecordAccepted(int64((g + i) * 1_000))
}
// Also exercise reads concurrently.
_ = tracker.ValidationPassRate()
_ = tracker.Connected()
_ = tracker.LatencyMs()
}
}()
}
wg.Wait()

total := tracker.TotalReceived()
accepted := tracker.TotalAccepted()
rejected := tracker.TotalRejected()
if accepted+rejected != total {
t.Errorf("counter invariant violated: accepted(%d)+rejected(%d) != received(%d)",
accepted, rejected, total)
}
}
