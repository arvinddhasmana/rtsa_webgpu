// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc/metadata"
)

// Standard RTSA message header keys.
const (
	HeaderClassification = "rtsa-classification"
	HeaderSourceService  = "rtsa-source-service"
	HeaderTraceID        = "rtsa-trace-id"
	HeaderTimestamp      = "rtsa-timestamp"
	HeaderSchemaVersion  = "rtsa-schema-version"
)

// StandardHeaders returns the required headers for every RTSA message.
func StandardHeaders(classification, sourceService, traceID, schemaVersion string) []kgo.RecordHeader {
	return []kgo.RecordHeader{
		{Key: HeaderClassification, Value: []byte(classification)},
		{Key: HeaderSourceService, Value: []byte(sourceService)},
		{Key: HeaderTraceID, Value: []byte(traceID)},
		{Key: HeaderTimestamp, Value: []byte(time.Now().UTC().Format(time.RFC3339Nano))},
		{Key: HeaderSchemaVersion, Value: []byte(schemaVersion)},
	}
}

// StandardHeadersFromContext extracts classification and trace ID from context if available.
func StandardHeadersFromContext(ctx context.Context, sourceService, schemaVersion string) []kgo.RecordHeader {
	classification := "UNCLASSIFIED"
	traceID := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("x-classification"); len(v) > 0 {
			classification = v[0]
		}
		if v := md.Get("x-trace-id"); len(v) > 0 {
			traceID = v[0]
		}
	}
	// Fallback to outgoing if incoming not present
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if classification == "UNCLASSIFIED" {
			if v := md.Get("x-classification"); len(v) > 0 {
				classification = v[0]
			}
		}
		if traceID == "" {
			if v := md.Get("x-trace-id"); len(v) > 0 {
				traceID = v[0]
			}
		}
	}
	return StandardHeaders(classification, sourceService, traceID, schemaVersion)
}

// GetHeader extracts a header value from a record. Returns "" if not found.
func GetHeader(record *kgo.Record, key string) string {
	for _, h := range record.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
