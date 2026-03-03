// CLASSIFICATION: UNCLASSIFIED
//go:build e2e

// Package e2e provides shared test helpers for all E2E tests.
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// skipE2E skips the test unless RTSA_INTEGRATION_TESTS=true.
// All E2E tests must call this as their first statement.
func skipE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("RTSA_INTEGRATION_TESTS") != "true" {
		t.Skip("E2E tests disabled: set RTSA_INTEGRATION_TESTS=true to enable")
	}
}

// redpandaBroker returns the Kafka broker address for the test stack.
// Defaults to localhost:19092 (host-mapped from docker-compose.yml).
func redpandaBroker() string {
	if b := os.Getenv("RTSA_REDPANDA_BROKERS"); b != "" {
		return b
	}
	return "localhost:19092"
}

// clickhouseEndpoint returns the ClickHouse HTTP endpoint.
// Defaults to localhost:8123.
func clickhouseEndpoint() string {
	if ep := os.Getenv("RTSA_CLICKHOUSE_ENDPOINT"); ep != "" {
		return ep
	}
	return "localhost:8123"
}

// radarIngestionEndpoint returns the gRPC endpoint of svc-radar-ingestion.
// Defaults to localhost:50051 (host-mapped from docker-compose.services.yml).
func radarIngestionEndpoint() string {
	if ep := os.Getenv("RTSA_RADAR_ENDPOINT"); ep != "" {
		return ep
	}
	return "localhost:50051"
}

// aisIngestionEndpoint returns the gRPC endpoint of svc-ais-ingestion.
// Defaults to localhost:50055 (host-mapped from docker-compose.services.yml).
func aisIngestionEndpoint() string {
	if ep := os.Getenv("RTSA_AIS_ENDPOINT"); ep != "" {
		return ep
	}
	return "localhost:50055"
}

// grpcDialCtx dials a gRPC endpoint using insecure credentials.
// It respects the provided context deadline and fatally fails the test on error.
func grpcDialCtx(ctx context.Context, t *testing.T, endpoint string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpcDialCtx: dial %s: %v", endpoint, err)
	}
	return conn
}

// newKafkaProducer creates a Kafka producer client and fatally fails the test on error.
func newKafkaProducer(t *testing.T, broker string) *kgo.Client {
	t.Helper()
	c, err := kgo.NewClient(kgo.SeedBrokers(broker))
	if err != nil {
		t.Fatalf("newKafkaProducer: %v", err)
	}
	return c
}

// newKafkaConsumer creates a Kafka consumer client with the given group and topics,
// resetting to the given offset strategy, and fatally fails the test on error.
func newKafkaConsumer(t *testing.T, broker, group string, offset kgo.Offset, topics ...string) *kgo.Client {
	t.Helper()
	c, err := kgo.NewClient(
		kgo.SeedBrokers(broker),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		kgo.ConsumeResetOffset(offset),
	)
	if err != nil {
		t.Fatalf("newKafkaConsumer: %v", err)
	}
	return c
}

// pollUntil polls the given Kafka consumer until cond returns true or deadline
// is exceeded. Returns true if cond was satisfied before the deadline.
func pollUntil(ctx context.Context, consumer *kgo.Client, timeout time.Duration, cond func(*kgo.Record) bool) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			fetches := consumer.PollRecords(ctx, 10)
			var satisfied bool
			fetches.EachRecord(func(r *kgo.Record) {
				if cond(r) {
					satisfied = true
				}
			})
			if satisfied {
				return true
			}
		}
	}
}

// pollCount polls the consumer for the given duration and returns the total
// number of records received. Used for counting-based assertions.
func pollCount(ctx context.Context, consumer *kgo.Client, duration time.Duration) int {
	deadline := time.After(duration)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	var total int
	for {
		select {
		case <-deadline:
			return total
		case <-ticker.C:
			fetches := consumer.PollRecords(ctx, 100)
			fetches.EachRecord(func(_ *kgo.Record) { total++ })
		}
	}
}

// generateObsID creates a deterministic observation ID: "<prefix>-<idx>".
func generateObsID(prefix string, idx int) string {
	return prefix + "-" + itoa(idx)
}

// itoa converts a non-negative integer to its decimal string representation.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
