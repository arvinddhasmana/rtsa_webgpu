// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"context"
"testing"
"time"

"github.com/arvinddhasmana/rtsa_webgpu/svc-query/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateTimeRange_Valid(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 30)
now := time.Now()
start := timestamppb.New(now.Add(-24 * time.Hour))
end := timestamppb.New(now)
if err := g.ValidateTimeRange(start, end); err != nil {
t.Errorf("expected no error, got %v", err)
}
}

func TestValidateTimeRange_NilStart(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 30)
if err := g.ValidateTimeRange(nil, timestamppb.Now()); err == nil {
t.Error("expected error for nil start")
}
}

func TestValidateTimeRange_NilEnd(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 30)
if err := g.ValidateTimeRange(timestamppb.Now(), nil); err == nil {
t.Error("expected error for nil end")
}
}

func TestValidateTimeRange_EndBeforeStart(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 30)
now := time.Now()
start := timestamppb.New(now)
end := timestamppb.New(now.Add(-1 * time.Hour))
if err := g.ValidateTimeRange(start, end); err == nil {
t.Error("expected error for end before start")
}
}

func TestValidateTimeRange_ExceedsMax(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 30)
now := time.Now()
start := timestamppb.New(now.Add(-31 * 24 * time.Hour))
end := timestamppb.New(now)
if err := g.ValidateTimeRange(start, end); err == nil {
t.Error("expected error for range exceeding max")
}
}

func TestEnforceRowLimit(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 30)
tests := []struct {
requested int
expected  int
}{
{0, 100000},
{50, 50},
{100000, 100000},
{200000, 100000},
{-1, 100000},
}
for _, tt := range tests {
got := g.EnforceRowLimit(tt.requested)
if got != tt.expected {
t.Errorf("EnforceRowLimit(%d) = %d, want %d", tt.requested, got, tt.expected)
}
}
}

func TestQueryContext(t *testing.T) {
g := domain.NewQueryGuardrail(30, 100000, 5)
ctx, cancel := g.QueryContext(context.Background())
defer cancel()
deadline, ok := ctx.Deadline()
if !ok {
t.Error("expected context with deadline")
}
if time.Until(deadline) > 6*time.Second {
t.Error("deadline should be within 5 seconds")
}
}
