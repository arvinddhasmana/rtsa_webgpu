// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/domain"
)

func TestEncodeDecode_RoundTrip(t *testing.T) {
original := &domain.PaginationToken{
LastID:        "audit-abc-123",
LastTimestamp: time.Now().Truncate(time.Millisecond),
PageSize:      100,
}
encoded := domain.EncodePaginationToken(original)
if encoded == "" {
t.Fatal("expected non-empty encoded token")
}

decoded, err := domain.DecodePaginationToken(encoded)
if err != nil {
t.Fatalf("unexpected decode error: %v", err)
}
if decoded.LastID != original.LastID {
t.Errorf("LastID: got %q, want %q", decoded.LastID, original.LastID)
}
if decoded.PageSize != original.PageSize {
t.Errorf("PageSize: got %d, want %d", decoded.PageSize, original.PageSize)
}
}

func TestEncode_NilToken(t *testing.T) {
encoded := domain.EncodePaginationToken(nil)
if encoded != "" {
t.Errorf("expected empty string for nil token, got %q", encoded)
}
}

func TestDecode_EmptyString(t *testing.T) {
token, err := domain.DecodePaginationToken("")
if err != nil {
t.Errorf("unexpected error for empty string: %v", err)
}
if token != nil {
t.Errorf("expected nil token for empty string, got %v", token)
}
}

func TestDecode_InvalidBase64(t *testing.T) {
_, err := domain.DecodePaginationToken("not-valid-base64!!")
if err == nil {
t.Error("expected error for invalid base64")
}
}

func TestDecode_InvalidJSON(t *testing.T) {
import64 := "bm90LWpzb24="  // base64 of "not-json"
_, err := domain.DecodePaginationToken(import64)
if err == nil {
t.Error("expected error for invalid JSON")
}
}
