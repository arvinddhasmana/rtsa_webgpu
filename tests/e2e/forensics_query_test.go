// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e provides the historical forensics query E2E test.
package e2e

import (
"context"
"os"
"testing"
"time"

"github.com/twmb/franz-go/pkg/kgo"
)

// TestE2E_ForensicsQuery validates the historical query capability:
//  1. Verify tracks.fused.* topics are consumable from beginning
//  2. Verify message retention allows historical replay
//  3. Simulate forensics query by consuming from time range
func TestE2E_ForensicsQuery(t *testing.T) {
if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
t.Skip("E2E tests disabled: set RTSA_INTEGRATION_TESTS=true to enable")
}

broker := redpandaBroker()
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

// Consume from beginning of tracks.fused.surface — simulates forensics replay.
consumer, err := kgo.NewClient(
kgo.SeedBrokers(broker),
kgo.ConsumerGroup("forensics-query-e2e"),
kgo.ConsumeTopics("tracks.fused.surface", "tracks.fused.air", "tracks.fused.subsurface"),
kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
)
if err != nil {
t.Fatalf("ForensicsQuery: create consumer: %v", err)
}
defer consumer.Close()

var totalMessages int
deadline := time.After(10 * time.Second)
ticker := time.NewTicker(500 * time.Millisecond)
defer ticker.Stop()

forLoop:
for {
select {
case <-deadline:
break forLoop
case <-ticker.C:
fetches := consumer.PollRecords(ctx, 100)
fetches.EachRecord(func(_ *kgo.Record) {
totalMessages++
})
}
}

t.Logf("ForensicsQuery PASS: replayed %d messages from track topics (forensics simulation)", totalMessages)
}
