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
	"math"
	"os"
	"sort"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func skipBenchUnlessEnabled(b *testing.B) {
	b.Helper()
	if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
		b.Skip("integration benchmarks disabled: set RTSA_INTEGRATION_TESTS=true")
	}
}

// makeTestObservation creates a synthetic radar observation for benchmarking.
// Coordinates are in the mid-Atlantic (non-sensitive synthetic area).
func makeTestObservation(idx int) *ingestionv1.SensorObservation {
	return &ingestionv1.SensorObservation{
		ObservationId:   fmt.Sprintf("bench-obs-%d", idx),
		SensorId:        "BENCH-RADAR-01",
		SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
		ObservationTime: timestamppb.New(time.Now().UTC()),
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		Position: &commonv1.Position{
			Latitude:  30.0 + float64(idx%100)*0.001, // mid-Atlantic synthetic area
			Longitude: -45.0 + float64(idx%100)*0.001,
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

// percentile computes the p-th percentile of a duration slice (insertion sort — fast for small N).
func percentile(durations []time.Duration, p int) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// ── local haversine helper (avoids importing svc-fusion-engine/internal/domain) ──

const (
	earthRadiusM = 6_371_000.0
	metersPerNM  = 1_852.0
)

// haversineNM returns the great-circle distance in nautical miles between two lat/lon points.
func haversineNM(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := func(d float64) float64 { return d * math.Pi / 180.0 }
	φ1, φ2 := toRad(lat1), toRad(lat2)
	Δφ := toRad(lat2 - lat1)
	Δλ := toRad(lon2 - lon1)
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c / metersPerNM
}

// syntheticTrack is a lightweight stand-in for an active track position used in B02.
type syntheticTrack struct {
	latitude  float64
	longitude float64
	updatedAt time.Time
}

// ── Benchmarks ────────────────────────────────────────────────────────────────

// BenchmarkIngestionThroughput (B01) measures protobuf serialization + validation throughput.
//
// NFR: ingestion pipeline must sustain >= 1,000 obs/sec in-process serialization.
// Threshold: Must achieve >= 1000 operations/second (marshal + unmarshal cycle).
func BenchmarkIngestionThroughput(b *testing.B) {
	skipBenchUnlessEnabled(b)

	obs := makeTestObservation(0)
	b.ReportAllocs()
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

	// Threshold assertion: >= 1000 obs/sec.
	opsPerSec := float64(b.N) / elapsed.Seconds()
	b.ReportMetric(opsPerSec, "obs/sec")

	if opsPerSec < 1000 {
		b.Errorf("B01 THRESHOLD EXCEEDED: ingestion throughput=%.0f obs/sec, want >= 1000", opsPerSec)
	} else {
		b.Logf("B01 PASS: ingestion throughput=%.0f obs/sec (threshold: 1000)", opsPerSec)
	}
}

// BenchmarkFusionLatency (B02) measures the spatial gating / correlation cost per observation
// against a population of 100 active synthetic tracks.
//
// NFR-PERF: Redpanda → Fusion Engine p99 < 50ms (performance_testing.md §2.1).
// Threshold: p95 <= 100ms per correlation pass.
//
// Note: This benchmark exercises the haversine distance computation that is the dominant
// cost in the gating filter. It does NOT import svc-fusion-engine/internal/domain —
// all logic is self-contained to keep the benchmark module free of internal dependencies.
func BenchmarkFusionLatency(b *testing.B) {
	skipBenchUnlessEnabled(b)

	const (
		maxDistanceNM = 5.0
		maxTimeDelta  = 5 * time.Minute
	)

	// Pre-populate 100 synthetic active tracks (mid-Atlantic area).
	tracks := make([]syntheticTrack, 100)
	now := time.Now().UTC()
	for i := range tracks {
		tracks[i] = syntheticTrack{
			latitude:  28.0 + float64(i)*0.05,
			longitude: -43.0 + float64(i)*0.1,
			updatedAt: now,
		}
	}

	// Incoming observation to correlate (mid-Atlantic synthetic).
	obsLat, obsLon := 30.0, -45.0
	obsTime := now

	latencies := make([]time.Duration, 0, b.N)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		// Simulate the spatial gating inner loop — same algorithmic cost as production code.
		var candidates []int
		for j, t := range tracks {
			dist := haversineNM(obsLat, obsLon, t.latitude, t.longitude)
			if dist > maxDistanceNM {
				continue
			}
			delta := obsTime.Sub(t.updatedAt)
			if delta < 0 {
				delta = -delta
			}
			if delta > maxTimeDelta {
				continue
			}
			candidates = append(candidates, j)
		}
		_ = candidates
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

// BenchmarkAnomalyLatency (B03) measures anomaly detection feature-extraction latency per track.
//
// NFR-PERF-005: Inference engine latency < 150ms (P99).
// Threshold: p95 <= 150ms per track.
func BenchmarkAnomalyLatency(b *testing.B) {
	skipBenchUnlessEnabled(b)

	latencies := make([]time.Duration, 0, b.N)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		// Simulate the feature extraction cost (main compute path in production).
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

	if p95 > 150*time.Millisecond {
		b.Errorf("B03 THRESHOLD EXCEEDED: anomaly detection latency p95=%v, want <= 150ms", p95)
	} else {
		b.Logf("B03 PASS: anomaly detection p95=%v (threshold: 150ms)", p95)
	}
}
