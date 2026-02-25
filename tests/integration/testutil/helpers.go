// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package testutil

import (
"context"
"testing"
"time"

"github.com/twmb/franz-go/pkg/kgo"
)

// WaitForTopicMessages polls a Kafka topic until at least count messages are received
// or the timeout is reached. Returns all collected records.
func WaitForTopicMessages(t *testing.T, consumer *kgo.Client, count int, timeout time.Duration) []*kgo.Record {
t.Helper()
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

var records []*kgo.Record
ticker := time.NewTicker(100 * time.Millisecond)
defer ticker.Stop()

for {
select {
case <-ctx.Done():
t.Logf("WaitForTopicMessages: timeout after %v (got %d/%d messages)", timeout, len(records), count)
return records
case <-ticker.C:
fetches := consumer.PollRecords(ctx, count)
fetches.EachRecord(func(r *kgo.Record) {
records = append(records, r)
})
if len(records) >= count {
return records
}
}
}
}

// AssertMessageOnTopic asserts that at least one message arrives on the consumer's topic
// within the timeout.
func AssertMessageOnTopic(t *testing.T, consumer *kgo.Client, timeout time.Duration) *kgo.Record {
t.Helper()
records := WaitForTopicMessages(t, consumer, 1, timeout)
if len(records) == 0 {
t.Fatal("AssertMessageOnTopic: no messages received within timeout")
}
return records[0]
}

// AssertHeaderPresent asserts that a message record contains a header with the given key.
func AssertHeaderPresent(t *testing.T, record *kgo.Record, headerKey string) string {
t.Helper()
for _, h := range record.Headers {
if h.Key == headerKey {
return string(h.Value)
}
}
t.Errorf("AssertHeaderPresent: header %q not found in record headers %v", headerKey, record.Headers)
return ""
}

// AssertHeaderValue asserts a header has the expected value.
func AssertHeaderValue(t *testing.T, record *kgo.Record, headerKey, expectedValue string) {
t.Helper()
val := AssertHeaderPresent(t, record, headerKey)
if val != expectedValue {
t.Errorf("AssertHeaderValue: header %q = %q, want %q", headerKey, val, expectedValue)
}
}

// AssertAtLeast fails if actual < min.
func AssertAtLeast(t *testing.T, min, actual int, msg string) {
t.Helper()
if actual < min {
t.Errorf("%s: expected at least %d, got %d", msg, min, actual)
}
}

// AssertFloatGTE fails if actual < expected.
func AssertFloatGTE(t *testing.T, expected, actual float64, msg string) {
t.Helper()
if actual < expected {
t.Errorf("%s: expected >= %.4f, got %.4f", msg, expected, actual)
}
}

// AssertFloatLTE fails if actual > expected.
func AssertFloatLTE(t *testing.T, expected, actual float64, msg string) {
t.Helper()
if actual > expected {
t.Errorf("%s: expected <= %.4f, got %.4f", msg, expected, actual)
}
}

// AssertNotEmpty fails if the slice is empty.
func AssertNotEmpty[T any](t *testing.T, slice []T, msg string) {
t.Helper()
if len(slice) == 0 {
t.Errorf("%s: expected non-empty slice, got empty", msg)
}
}

// AssertEqual fails if a != b.
func AssertEqual[T comparable](t *testing.T, expected, actual T, msg string) {
t.Helper()
if expected != actual {
t.Errorf("%s: expected %v, got %v", msg, expected, actual)
}
}
