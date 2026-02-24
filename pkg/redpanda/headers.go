// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"time"

"github.com/twmb/franz-go/pkg/kgo"
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

// GetHeader extracts a header value from a record. Returns "" if not found.
func GetHeader(record *kgo.Record, key string) string {
for _, h := range record.Headers {
if h.Key == key {
return string(h.Value)
}
}
return ""
}
