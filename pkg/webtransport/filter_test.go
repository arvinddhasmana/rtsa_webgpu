// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/filter_test.go — Tests for record filtering and origin validation

package webtransport_test

import (
"encoding/binary"
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/pkg/webtransport"
)

// buildRecord creates a 128-byte record with the given classification and threat levels.
func buildRecord(classLevel, threatLevel uint32) []byte {
rec := make([]byte, webtransport.RecordSize)
binary.LittleEndian.PutUint32(rec[webtransport.OffClassificationLevel:], classLevel)
binary.LittleEndian.PutUint32(rec[webtransport.OffThreatLevel:], threatLevel)
return rec
}

func TestRecordFilter_ClassificationBlock(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 2} // PROTECTED_A

// Track at PROTECTED_B (3) — above clearance → drop
rec := buildRecord(3, 0)
if f.ShouldSendRecord(rec, false) {
t.Error("PROTECTED_B track must not be sent to PROTECTED_A operator")
}
}

func TestRecordFilter_ClassificationAllow(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 5} // SECRET

// Track at UNCLASSIFIED (1) — below clearance → send
rec := buildRecord(1, 2) // Friendly
if !f.ShouldSendRecord(rec, false) {
t.Error("UNCLASSIFIED track must be sent to SECRET operator")
}
}

func TestRecordFilter_PrioritySheddingAllowed(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 5}

// Hostile (5) under congestion → always send
rec := buildRecord(1, 5)
if !f.ShouldSendRecord(rec, true) {
t.Error("Hostile track must be sent even under congestion")
}
}

func TestRecordFilter_PrioritySheddingDropped(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 5}

// Friendly (2) under congestion → drop
rec := buildRecord(1, 2)
if f.ShouldSendRecord(rec, true) {
t.Error("Friendly track must be dropped under congestion")
}
}

func TestRecordFilter_ShortRecord(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 5}
shortRec := make([]byte, 64)
if f.ShouldSendRecord(shortRec, false) {
t.Error("short record must be rejected")
}
}

func TestReadU32LE(t *testing.T) {
// Verify little-endian reading via the record filter path
rec := make([]byte, webtransport.RecordSize)
// Write 0xDEADBEEF at OffClassificationLevel in LE
binary.LittleEndian.PutUint32(rec[webtransport.OffClassificationLevel:], 0xDEADBEEF)
// Use RecordFilter to exercise readU32LE indirectly
f := &webtransport.RecordFilter{ClearanceLevel: 0xDEADBEEF}
rec2 := buildRecord(0xDEADBEEF, 0)
// classLevel == clearance → should pass classification filter
if !f.ShouldSendRecord(rec2, false) {
t.Error("record with classLevel == clearance should pass classification filter")
}
_ = rec
}

func TestIsAllowedOrigin_EmptyList(t *testing.T) {
// Empty list = dev mode, all origins allowed
if !webtransport.IsAllowedOrigin("https://anything.example.com", nil) {
t.Error("empty allowed list must permit all origins")
}
}

func TestIsAllowedOrigin_MatchFound(t *testing.T) {
allowed := []string{"https://rtsa.mil.ca", "https://rtsa-dev.local"}
if !webtransport.IsAllowedOrigin("https://rtsa.mil.ca", allowed) {
t.Error("matching origin must be allowed")
}
}

func TestIsAllowedOrigin_NoMatch(t *testing.T) {
allowed := []string{"https://rtsa.mil.ca"}
if webtransport.IsAllowedOrigin("https://evil.com", allowed) {
t.Error("non-matching origin must be denied")
}
}
