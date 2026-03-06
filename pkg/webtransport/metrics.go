// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/metrics.go — OpenTelemetry metric attribute constants

package webtransport

import "go.opentelemetry.io/otel/attribute"

// RecordSize is the fixed size in bytes of one FlatBuffer track record.
// Matches flatbuf.RecordSize.
const RecordSize = 128

// OffClassificationLevel and OffThreatLevel are byte offsets within the 128-byte
// record for in-flight header access by the server.
const OffClassificationLevel = 0x1C
const OffThreatLevel = 0x20

// OTel attribute values for drop reason labels.
var (
attrReasonCongestion      = attribute.String("reason", "congestion")
attrReasonClassification  = attribute.String("reason", "classification")
attrPriorityAll           = attribute.String("priority", "all")
)

// readU32LE reads a little-endian uint32 from b at offset off.
func readU32LE(b []byte, off int) uint32 {
if len(b) < off+4 {
return 0
}
return uint32(b[off]) |
uint32(b[off+1])<<8 |
uint32(b[off+2])<<16 |
uint32(b[off+3])<<24
}
