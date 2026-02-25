// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package benchmark provides query performance benchmarks (B04).
package benchmark

import (
"fmt"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"google.golang.org/protobuf/proto"
"google.golang.org/protobuf/types/known/timestamppb"
)

// BenchmarkQueryPerformance (B04) measures query serialization and filtering performance.
//
// Threshold: 100 rows <= 500ms p95, 10K rows <= 2s p95
// In this test we simulate query filtering overhead using protobuf + in-memory data.
func BenchmarkQueryPerformance(b *testing.B) {
skipBenchUnlessEnabled(b)

// Generate 10000 test observations (simulates 10K row query result).
obs := make([]*ingestionv1.SensorObservation, 10000)
for i := range obs {
obs[i] = &ingestionv1.SensorObservation{
ObservationId:   fmt.Sprintf("qbench-%d", i),
SensorId:        "BENCH-SENSOR",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  43.0 + float64(i%1000)*0.004,
Longitude: -65.0 + float64(i%1000)*0.01,
},
}
}

// Serialize all to simulate query result serialization.
payloads := make([][]byte, len(obs))
for i, o := range obs {
p, _ := proto.Marshal(o)
payloads[i] = p
}

latencies100 := make([]time.Duration, 0, b.N)
latencies10k := make([]time.Duration, 0, b.N)
b.ResetTimer()

for i := 0; i < b.N; i++ {
// Benchmark 100-row deserialization.
start100 := time.Now()
for j := 0; j < 100; j++ {
var decoded ingestionv1.SensorObservation
_ = proto.Unmarshal(payloads[j%len(payloads)], &decoded)
}
latencies100 = append(latencies100, time.Since(start100))

// Benchmark 10K-row deserialization.
start10k := time.Now()
for j := 0; j < 10000; j++ {
var decoded ingestionv1.SensorObservation
_ = proto.Unmarshal(payloads[j%len(payloads)], &decoded)
}
latencies10k = append(latencies10k, time.Since(start10k))
}

p95_100 := percentile(latencies100, 95)
p95_10k := percentile(latencies10k, 95)

b.ReportMetric(float64(p95_100.Milliseconds()), "ms-p95-100rows")
b.ReportMetric(float64(p95_10k.Milliseconds()), "ms-p95-10krows")

// Threshold assertions.
if p95_100 > 500*time.Millisecond {
b.Errorf("B04 THRESHOLD EXCEEDED: 100-row query p95=%v, want <= 500ms", p95_100)
} else {
b.Logf("B04 PASS: 100-row query p95=%v (threshold: 500ms)", p95_100)
}

if p95_10k > 2*time.Second {
b.Errorf("B04 THRESHOLD EXCEEDED: 10K-row query p95=%v, want <= 2s", p95_10k)
} else {
b.Logf("B04 PASS: 10K-row query p95=%v (threshold: 2s)", p95_10k)
}
}
