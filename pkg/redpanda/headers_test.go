// CLASSIFICATION: UNCLASSIFIED
package redpanda_test

import (
	"context"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc/metadata"
)

func TestStandardHeaders_WithMetadata(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-classification", "PROTECTED_A",
		"x-trace-id", "trace-1",
	))
	h := redpanda.StandardHeadersFromContext(ctx, "test-svc", "1.0.0")
	if len(h) != 5 {
		t.Errorf("expected 5 headers, got %d", len(h))
	}
}

func TestStandardHeaders_AllPresent(t *testing.T) {
	headers := redpanda.StandardHeaders("UNCLASSIFIED", "svc-test", "trace-123", "1.0.0")
	if len(headers) != 5 {
		t.Errorf("expected 5 headers, got %d", len(headers))
	}
	keys := map[string]bool{}
	for _, h := range headers {
		keys[h.Key] = true
	}
	for _, expected := range []string{
		redpanda.HeaderClassification,
		redpanda.HeaderSourceService,
		redpanda.HeaderTraceID,
		redpanda.HeaderTimestamp,
		redpanda.HeaderSchemaVersion,
	} {
		if !keys[expected] {
			t.Errorf("missing header: %s", expected)
		}
	}
}

func TestGetHeader_Found(t *testing.T) {
	record := &kgo.Record{
		Headers: []kgo.RecordHeader{
			{Key: "rtsa-classification", Value: []byte("UNCLASSIFIED")},
		},
	}
	val := redpanda.GetHeader(record, "rtsa-classification")
	if val != "UNCLASSIFIED" {
		t.Errorf("expected UNCLASSIFIED, got %s", val)
	}
}

func TestGetHeader_NotFound(t *testing.T) {
	record := &kgo.Record{
		Headers: []kgo.RecordHeader{},
	}
	val := redpanda.GetHeader(record, "missing-key")
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
}
