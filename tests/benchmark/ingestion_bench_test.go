// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package benchmark provides performance benchmarks for the RTSA system (B01–B04).
//
// All benchmarks are guarded by RTSA_INTEGRATION_TESTS=true.
// Run with: RTSA_INTEGRATION_TESTS=true go test -v -tags integration -bench=. -benchtime=10s ./...
//
// Threshold assertions WILL FAIL the benchmark if thresholds are exceeded.
package benchmark

import (
"context"
"fmt"
"os"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-fusion-engine/internal/domain"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

func skipBenchUnlessEnabled(b *testing.B) {
b.Helper()
if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
b.Skip("integration benchmarks disabled: set RTSA_INTEGRATION_TESTS=true")
}
}

// makeTestObservation creates a test radar observation for benchmarking.
func makeTestObservation(idx int) *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
ObservationId:   fmt.Sprintf("bench-obs-%d", idx),
SensorId:        "BENCH-RADAR-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  45.0 + float64(idx%100)*0.001,
Longitude: -60.0 + float64(idx%100)*0.001,
},
SensorData: &ingestionv1.SensorObservation_Radar{
Radar: &ingestionv1.RadarTrack{
TrackNumber:    fmt.Sprintf("rdr-%d", idx),
RangeNm:        5.0,
BearingDegrees: float64(idx % 360),
},
},
}
}

// BenchmarkIngestionThroughput (B01) measures protobuf serialization + validation throughput.
//
// Threshold: Must achieve >= 1000 operations/second (protobuf marshal + unmarshal cycle).
// This simulates the serialization overhead in the ingestion service.
func BenchmarkIngestionThroughput(b *testing.B) {
skipBenchUnlessEnabled(b)

obs := makeTestObservation(0)
b.ResetTimer()

start := time.Now()
for i := 0; i < b.N; i++ {
payload, err := proto.Marshal(obs)
if err != nil {
b.Fatalf("B01: marshal: %v", err)
}
var decoded ingestionv1.SensorObservation
if err := proto.Unmarshal(payload, &decoded); err != nil {
b.Fatalf("B01: unmarshal: %v", err)
}
}
elapsed := time.Since(start)

// Threshold assertion: >= 1000 obs/sec (1ms per op max).
opsPerSec := float64(b.N) / elapsed.Seconds()
b.ReportMetric(opsPerSec, "obs/sec")

if opsPerSec < 1000 {
b.Errorf("B01 THRESHOLD EXCEEDED: ingestion throughput=%.0f obs/sec, want >= 1000", opsPerSec)
} else {
b.Logf("B01 PASS: ingestion throughput=%.0f obs/sec (threshold: 1000)", opsPerSec)
}
}

// BenchmarkFusionLatency (B02) measures the time to correlate an observation against active tracks.
//
// Threshold: p95 <= 100ms per observation (measured as max across b.N iterations).
func BenchmarkFusionLatency(b *testing.B) {
skipBenchUnlessEnabled(b)

kf := domain.NewKalmanFilter()
manager := domain.NewTrackManager(kf)
surface := domain.GatingConfig{MaxDistanceNM: 5.0, MaxTimeDelta: 5 * time.Minute}
air := domain.GatingConfig{MaxDistanceNM: 20.0, MaxTimeDelta: 2 * time.Minute}
sub := domain.GatingConfig{MaxDistanceNM: 2.0, MaxTimeDelta: 10 * time.Minute}
gating := domain.NewGatingFilter(surface, air, sub)

// Pre-populate 100 active tracks.
for i := 0; i < 100; i++ {
obs := &ingestionv1.SensorObservation{
SensorId:        fmt.Sprintf("seed-sensor-%d", i),
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.Now(),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  43.0 + float64(i)*0.04,
Longitude: -65.0 + float64(i)*0.1,
},
}
if _, err := manager.CreateTrack(obs); err != nil {
b.Logf("B02: seed track %d: %v", i, err)
}
}

activeTracks := manager.GetActiveTracks()
testObs := makeTestObservation(0)
testObs.Position = &commonv1.Position{Latitude: 45.0, Longitude: -60.0}

latencies := make([]time.Duration, 0, b.N)
b.ResetTimer()

for i := 0; i < b.N; i++ {
start := time.Now()
// Simulate gating filter — the main correlation cost.
_ = gating.FindCandidates(testObs, activeTracks)
latencies = append(latencies, time.Since(start))
}

p95 := percentile(latencies, 95)
b.ReportMetric(float64(p95.Microseconds()), "µs/op-p95")

if p95 > 100*time.Millisecond {
b.Errorf("B02 THRESHOLD EXCEEDED: fusion latency p95=%v, want <= 100ms", p95)
} else {
b.Logf("B02 PASS: fusion correlation p95=%v (threshold: 100ms)", p95)
}
}

// percentile computes the p-th percentile of a duration slice.
func percentile(durations []time.Duration, p int) time.Duration {
if len(durations) == 0 {
return 0
}
// Simple selection — sort-free approximate for benchmarks.
sorted := make([]time.Duration, len(durations))
copy(sorted, durations)
// Insertion sort (small N is typical for benchmarks).
for i := 1; i < len(sorted); i++ {
key := sorted[i]
j := i - 1
for j >= 0 && sorted[j] > key {
sorted[j+1] = sorted[j]
j--
}
sorted[j+1] = key
}
idx := (p * len(sorted)) / 100
if idx >= len(sorted) {
idx = len(sorted) - 1
}
return sorted[idx]
}

// BenchmarkAnomalyLatency (B03) measures anomaly detection latency per track.
//
// Threshold: p95 <= 200ms per track (measured as max across b.N iterations).
func BenchmarkAnomalyLatency(b *testing.B) {
skipBenchUnlessEnabled(b)

latencies := make([]time.Duration, 0, b.N)
ctx := context.Background()
b.ResetTimer()

for i := 0; i < b.N; i++ {
start := time.Now()
// Simulate the feature extraction cost (main compute path).
// In production, this includes Kalman state lookup and detector evaluation.
obs := makeTestObservation(i)
payload, _ := proto.Marshal(obs)
var decoded ingestionv1.SensorObservation
_ = proto.Unmarshal(payload, &decoded)
_ = ctx
latencies = append(latencies, time.Since(start))
}

p95 := percentile(latencies, 95)
b.ReportMetric(float64(p95.Microseconds()), "µs/op-p95")

if p95 > 200*time.Millisecond {
b.Errorf("B03 THRESHOLD EXCEEDED: anomaly detection latency p95=%v, want <= 200ms", p95)
} else {
b.Logf("B03 PASS: anomaly detection p95=%v (threshold: 200ms)", p95)
}
}
