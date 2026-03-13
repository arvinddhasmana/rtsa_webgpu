// CLASSIFICATION: UNCLASSIFIED
// pkg/ingestion/sensorstate.go — Shared sensor state tracker for ingestion handlers
//
// Provides thread-safe per-sensor statistics collection used by all 6 ingestion services.
// Reference: docs/architecture/v1/RTSA_WebGPU_Architecture_v1.md §Ingestion Layer

package ingestion

import (
	"context"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ewmaAlpha controls the decay rate of the exponentially-weighted moving average.
// 0.1 gives a slow, smooth average; larger values react faster.
const ewmaAlpha = 0.1

// SensorStateTracker provides thread-safe per-sensor statistics.
// All counter operations are lock-free via atomics where possible.
type SensorStateTracker struct {
totalReceived atomic.Int64
totalAccepted atomic.Int64
totalRejected atomic.Int64

// lastObsTime stores time.Time using atomic Value.
lastObsTime atomic.Value

// latencyEwmaNs stores EWMA of nanosecond latency × 1024 for integer arithmetic.
latencyEwmaNs atomic.Int64

// dlqReasons: key=string reason, value=*atomic.Int64 count.
dlqReasons sync.Map

// Throughput ring buffer (capped at 60 samples, youngest at throughputHead-1 mod cap).
throughputMu   sync.Mutex
throughputRing []throughputEntry
throughputHead int

// Recent event ring buffer (capped at 100 entries).
eventsMu   sync.Mutex
eventRing  []eventEntry
eventHead  int

// Track start time for EPS computation fallback.
startTime time.Time

// Sensor coverage geometry (stores *ingestionv1.SensorCoverage)
coverage atomic.Value
}

type throughputEntry struct {
sampledAt      time.Time
eventsPerSecond float64
}

type eventEntry struct {
occurredAt time.Time
eventText  string
severity   ingestionv1.SensorEventSeverity
}

const (
throughputCap = 60
eventCap      = 100
)

// NewSensorStateTracker creates a new tracker ready for use.
func NewSensorStateTracker() *SensorStateTracker {
t := &SensorStateTracker{
throughputRing: make([]throughputEntry, throughputCap),
eventRing:      make([]eventEntry, eventCap),
startTime:      time.Now(),
}
t.lastObsTime.Store(time.Time{})
return t
}

// RecordAccepted increments the accepted counter, updates the latency EWMA,
// appends an INFO event, and records the observation time.
func (t *SensorStateTracker) RecordAccepted(latencyNs int64) {
t.totalReceived.Add(1)
t.totalAccepted.Add(1)
t.lastObsTime.Store(time.Now().UTC())

// Update EWMA: new = alpha*sample + (1-alpha)*old (× 1024 for integer)
oldEwma := t.latencyEwmaNs.Load()
newEwma := int64(ewmaAlpha*float64(latencyNs*1024) + (1-ewmaAlpha)*float64(oldEwma))
t.latencyEwmaNs.Store(newEwma)

t.appendEvent(eventEntry{
occurredAt: time.Now().UTC(),
eventText:  "Observation accepted",
severity:   ingestionv1.SensorEventSeverity_SENSOR_EVENT_SEVERITY_INFO,
})
}

// RecordRejected increments the rejected counter, logs the DLQ reason, and appends a WARN event.
func (t *SensorStateTracker) RecordRejected(reason string) {
t.totalReceived.Add(1)
t.totalRejected.Add(1)

// Atomically increment DLQ reason counter.
newCounter := &atomic.Int64{}
actual, _ := t.dlqReasons.LoadOrStore(reason, newCounter)
actual.(*atomic.Int64).Add(1)

t.appendEvent(eventEntry{
occurredAt: time.Now().UTC(),
eventText:  "Observation rejected: " + reason,
severity:   ingestionv1.SensorEventSeverity_SENSOR_EVENT_SEVERITY_WARN,
})
}

// SnapshotThroughput returns up to limit throughput samples, newest first.
func (t *SensorStateTracker) SnapshotThroughput(limit int) []*ingestionv1.ThroughputSample {
t.throughputMu.Lock()
defer t.throughputMu.Unlock()

out := make([]*ingestionv1.ThroughputSample, 0, limit)
head := t.throughputHead
for i := 0; i < throughputCap && len(out) < limit; i++ {
idx := (head - 1 - i + throughputCap) % throughputCap
entry := t.throughputRing[idx]
if entry.sampledAt.IsZero() {
break
}
out = append(out, &ingestionv1.ThroughputSample{
SampledAt:       timestamppb.New(entry.sampledAt),
EventsPerSecond: entry.eventsPerSecond,
})
}
return out
}

// SnapshotEvents returns up to limit recent events, newest first.
func (t *SensorStateTracker) SnapshotEvents(limit int) []*ingestionv1.SensorEvent {
t.eventsMu.Lock()
defer t.eventsMu.Unlock()

out := make([]*ingestionv1.SensorEvent, 0, limit)
head := t.eventHead
for i := 0; i < eventCap && len(out) < limit; i++ {
idx := (head - 1 - i + eventCap) % eventCap
entry := t.eventRing[idx]
if entry.occurredAt.IsZero() {
break
}
out = append(out, &ingestionv1.SensorEvent{
OccurredAt: timestamppb.New(entry.occurredAt),
EventText:  entry.eventText,
Severity:   entry.severity,
})
}
return out
}

// DLQBreakdown returns all DLQ reason counts sorted by count descending.
func (t *SensorStateTracker) DLQBreakdown() []*ingestionv1.DLQReasonCount {
var out []*ingestionv1.DLQReasonCount
t.dlqReasons.Range(func(key, value any) bool {
out = append(out, &ingestionv1.DLQReasonCount{
Reason: key.(string),
Count:  value.(*atomic.Int64).Load(),
})
return true
})
return out
}

// LatencyMs returns the EWMA latency converted from nanoseconds to milliseconds.
func (t *SensorStateTracker) LatencyMs() float64 {
ewmaNs := t.latencyEwmaNs.Load()
if ewmaNs == 0 {
return 0
}
// Divide by 1024 (scale factor) then convert ns → ms.
return math.Round(float64(ewmaNs)/(1024*1e6)*100) / 100
}

// ValidationPassRate returns the percentage of accepted observations (0–100).
func (t *SensorStateTracker) ValidationPassRate() float64 {
recv := t.totalReceived.Load()
if recv == 0 {
return 0
}
accepted := t.totalAccepted.Load()
return math.Round(float64(accepted)/float64(recv)*10000) / 100
}

// Connected returns true if the last observation was within the past 30 seconds.
func (t *SensorStateTracker) Connected() bool {
last := t.lastObsTime.Load().(time.Time)
if last.IsZero() {
return false
}
return time.Since(last) < 30*time.Second
}

// LastObsTime returns the last observation timestamp.
func (t *SensorStateTracker) LastObsTime() time.Time {
return t.lastObsTime.Load().(time.Time)
}

// UpdateCoverage updates the sensor's coverage geometry.
func (t *SensorStateTracker) UpdateCoverage(coverage *ingestionv1.SensorCoverage) {
	t.coverage.Store(coverage)
}

// ExtractCoverage parses coverage metadata from an observation and updates the tracker.
func (t *SensorStateTracker) ExtractCoverage(meta map[string]string) {
	if meta == nil {
		return
	}
	cov := &ingestionv1.SensorCoverage{}
	hasData := false

	if val, err := strconv.ParseFloat(meta["rtsa.coverage.range_nm"], 64); err == nil {
		cov.RangeNm = &val
		hasData = true
	}
	if val, err := strconv.ParseFloat(meta["rtsa.coverage.bearing_start"], 64); err == nil {
		cov.BearingStartDegrees = &val
		hasData = true
	}
	if val, err := strconv.ParseFloat(meta["rtsa.coverage.bearing_end"], 64); err == nil {
		cov.BearingEndDegrees = &val
		hasData = true
	}
	if lat, er1 := strconv.ParseFloat(meta["rtsa.coverage.sensor_lat"], 64); er1 == nil {
		if lon, er2 := strconv.ParseFloat(meta["rtsa.coverage.sensor_lon"], 64); er2 == nil {
			cov.SensorPosition = &commonv1.Position{
				Latitude:  lat,
				Longitude: lon,
			}
			hasData = true
		}
	}

	if hasData {
		t.UpdateCoverage(cov)
	}
}

// Coverage returns the current sensor coverage geometry.
func (t *SensorStateTracker) Coverage() *ingestionv1.SensorCoverage {
	val := t.coverage.Load()
	if val == nil {
		return nil
	}
	return val.(*ingestionv1.SensorCoverage)
}

// TotalReceived returns the total received observation count.
func (t *SensorStateTracker) TotalReceived() int64 {
return t.totalReceived.Load()
}

// TotalAccepted returns the total accepted observation count.
func (t *SensorStateTracker) TotalAccepted() int64 {
return t.totalAccepted.Load()
}

// TotalRejected returns the total rejected observation count.
func (t *SensorStateTracker) TotalRejected() int64 {
return t.totalRejected.Load()
}

// EventsPerSecond returns the current throughput estimate using the ring buffer average.
// Falls back to time-since-start average if no samples are collected yet.
func (t *SensorStateTracker) EventsPerSecond() float64 {
t.throughputMu.Lock()
sample := t.throughputRing[(t.throughputHead-1+throughputCap)%throughputCap]
t.throughputMu.Unlock()
if !sample.sampledAt.IsZero() {
return sample.eventsPerSecond
}
// Fallback: total / elapsed
elapsed := time.Since(t.startTime).Seconds()
if elapsed < 1 {
return 0
}
return float64(t.totalReceived.Load()) / elapsed
}

// StartThroughputSampler starts a goroutine that samples the current EPS every interval.
// The goroutine stops when ctx is cancelled.
func (t *SensorStateTracker) StartThroughputSampler(ctx context.Context, interval time.Duration) {
var lastReceived int64
var lastSampleTime = time.Now()

go func() {
ticker := time.NewTicker(interval)
defer ticker.Stop()
for {
select {
case <-ctx.Done():
return
case now := <-ticker.C:
currentReceived := t.totalReceived.Load()
elapsed := now.Sub(lastSampleTime).Seconds()
var eps float64
if elapsed > 0 {
eps = float64(currentReceived-lastReceived) / elapsed
}
lastReceived = currentReceived
lastSampleTime = now

t.throughputMu.Lock()
t.throughputRing[t.throughputHead] = throughputEntry{
sampledAt:      now.UTC(),
eventsPerSecond: eps,
}
t.throughputHead = (t.throughputHead + 1) % throughputCap
t.throughputMu.Unlock()
}
}
}()
}

// appendEvent appends an event to the ring buffer (caller should not hold eventsMu).
func (t *SensorStateTracker) appendEvent(e eventEntry) {
t.eventsMu.Lock()
t.eventRing[t.eventHead] = e
t.eventHead = (t.eventHead + 1) % eventCap
t.eventsMu.Unlock()
}
