// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"testing"
"time"
)

func TestEncodeDecode_roundtrip(t *testing.T) {
token := &PaginationToken{
LastID:        "track-abc-123",
LastTimestamp: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
PageSize:      100,
}

encoded := EncodePaginationToken(token)
if encoded == "" {
t.Fatal("expected non-empty encoded token")
}

decoded, err := DecodePaginationToken(encoded)
if err != nil {
t.Fatalf("unexpected decode error: %v", err)
}
if decoded == nil {
t.Fatal("expected non-nil decoded token")
}
if decoded.LastID != token.LastID {
t.Errorf("LastID = %q, want %q", decoded.LastID, token.LastID)
}
if !decoded.LastTimestamp.Equal(token.LastTimestamp) {
t.Errorf("LastTimestamp = %v, want %v", decoded.LastTimestamp, token.LastTimestamp)
}
if decoded.PageSize != token.PageSize {
t.Errorf("PageSize = %d, want %d", decoded.PageSize, token.PageSize)
}
}

func TestEncodeNilToken(t *testing.T) {
got := EncodePaginationToken(nil)
if got != "" {
t.Errorf("expected empty string for nil token, got %q", got)
}
}

func TestDecodeEmptyString(t *testing.T) {
tok, err := DecodePaginationToken("")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if tok != nil {
t.Errorf("expected nil for empty string, got %+v", tok)
}
}

func TestDecodeInvalidBase64(t *testing.T) {
_, err := DecodePaginationToken("!!!not-valid-base64!!!")
if err == nil {
t.Fatal("expected error for invalid base64")
}
}

func TestDecodeInvalidJSON(t *testing.T) {
// base64("invalid-json")
invalidB64 := "aW52YWxpZC1qc29u"
_, err := DecodePaginationToken(invalidB64)
if err == nil {
t.Fatal("expected error for invalid JSON")
}
}

func TestDecodeZeroPageSize(t *testing.T) {
token := &PaginationToken{
LastID:        "x",
LastTimestamp: time.Now(),
PageSize:      0,
}
encoded := EncodePaginationToken(token)
_, err := DecodePaginationToken(encoded)
if err == nil {
t.Fatal("expected error for zero page size")
}
}

func TestApplyPagination_firstPage(t *testing.T) {
query := "SELECT * FROM tracks_fused WHERE event_time >= ?"
params := []interface{}{time.Now()}

q, p := ApplyPagination(query, params, nil, "track_id", 50)

wantSuffix := " ORDER BY event_time ASC, track_id ASC LIMIT ?"
if len(q) < len(wantSuffix) || q[len(q)-len(wantSuffix):] != wantSuffix {
t.Errorf("query does not end with expected suffix:\n  got: %q\n want suffix: %q", q, wantSuffix)
}
// params should have gained 1 (the LIMIT value)
if len(p) != len(params)+1 {
t.Errorf("params len = %d, want %d", len(p), len(params)+1)
}
if p[len(p)-1] != 50 {
t.Errorf("LIMIT param = %v, want 50", p[len(p)-1])
}
}

func TestApplyPagination_subsequentPage(t *testing.T) {
token := &PaginationToken{
LastID:        "track-xyz",
LastTimestamp: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
PageSize:      100,
}
query := "SELECT * FROM tracks_fused WHERE event_time >= ?"
params := []interface{}{time.Now()}

q, p := ApplyPagination(query, params, token, "track_id", 100)

if len(q) == 0 {
t.Fatal("empty query")
}
// Should have gained 4 params: 2 for event_time, 1 for track_id, 1 for LIMIT
if len(p) != len(params)+4 {
t.Errorf("params len = %d, want %d", len(p), len(params)+4)
}
}
