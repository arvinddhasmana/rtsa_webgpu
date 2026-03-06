// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/filter.go — Record-level filtering logic
//
// Extracts the classification and priority filtering into a testable function
// independent of the WebTransport network layer.

package webtransport

// RecordFilter determines whether a 128-byte FlatBuffer record should be sent
// to a specific operator session. It combines classification filtering and
// priority shedding into a single decision.
type RecordFilter struct {
// ClearanceLevel is the operator's clearance level from the JWT claims.
ClearanceLevel uint32
}

// ShouldSendRecord returns true if the record should be forwarded to the
// operator. It applies both classification filtering and priority shedding.
//
// Parameters:
//   - rec: the 128-byte FlatBuffer record (read-only)
//   - congested: true if the QUIC connection is experiencing backpressure
func (f *RecordFilter) ShouldSendRecord(rec []byte, congested bool) bool {
if len(rec) < RecordSize {
return false
}
classLevel := readU32LE(rec, OffClassificationLevel)
threatLevel := readU32LE(rec, OffThreatLevel)
return ShouldSendByClassification(classLevel, f.ClearanceLevel) &&
ShouldSend(threatLevel, congested)
}

// IsAllowedOrigin checks whether the given origin header value is in the
// allowed origins list. An empty list means all origins are allowed (dev mode).
func IsAllowedOrigin(origin string, allowedOrigins []string) bool {
if len(allowedOrigins) == 0 {
return true
}
for _, allowed := range allowedOrigins {
if origin == allowed {
return true
}
}
return false
}
