// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
)

func TestPaginationTokenRoundTrip(t *testing.T) {
token := &domain.PaginationToken{
LastID:        "track-001",
LastTimestamp: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
PageSize:      100,
}
encoded := domain.EncodePaginationToken(token)
if encoded == "" {
t.Fatal("expected non-empty encoded token")
}
decoded, err := domain.DecodePaginationToken(encoded)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if decoded.LastID != token.LastID {
t.Errorf("LastID mismatch: got %q, want %q", decoded.LastID, token.LastID)
}
if decoded.PageSize != token.PageSize {
t.Errorf("PageSize mismatch: got %d, want %d", decoded.PageSize, token.PageSize)
}
}

func TestDecodePaginationToken_Empty(t *testing.T) {
token, err := domain.DecodePaginationToken("")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if token != nil {
t.Error("expected nil token for empty string")
}
}

func TestDecodePaginationToken_Invalid(t *testing.T) {
_, err := domain.DecodePaginationToken("not-valid-base64!!!")
if err == nil {
t.Error("expected error for invalid token")
}
}

func TestEncodePaginationToken_Nil(t *testing.T) {
encoded := domain.EncodePaginationToken(nil)
if encoded != "" {
t.Errorf("expected empty string for nil token, got %q", encoded)
}
}
