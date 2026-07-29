// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

"github.com/arvinddhasmana/rtsa_webgpu/svc-audit/internal/domain"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateTimeRange_NilStart(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
err := g.ValidateTimeRange(nil, timestamppb.New(time.Now()))
if err == nil {
t.Error("expected error for nil start")
}
st, _ := status.FromError(err)
if st.Code() != codes.InvalidArgument {
t.Errorf("expected InvalidArgument, got %v", st.Code())
}
}

func TestValidateTimeRange_NilEnd(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
err := g.ValidateTimeRange(timestamppb.New(time.Now()), nil)
if err == nil {
t.Error("expected error for nil end")
}
}

func TestValidateTimeRange_EndBeforeStart(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
now := time.Now()
err := g.ValidateTimeRange(
timestamppb.New(now),
timestamppb.New(now.Add(-1*time.Hour)),
)
if err == nil {
t.Error("expected error for end before start")
}
}

func TestValidateTimeRange_ExceedsMaxDays(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
now := time.Now()
err := g.ValidateTimeRange(
timestamppb.New(now.Add(-91*24*time.Hour)),
timestamppb.New(now),
)
if err == nil {
t.Error("expected error for range exceeding 90 days")
}
}

func TestValidateTimeRange_Valid(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
now := time.Now()
err := g.ValidateTimeRange(
timestamppb.New(now.Add(-1*time.Hour)),
timestamppb.New(now),
)
if err != nil {
t.Errorf("expected no error for valid range, got %v", err)
}
}

func TestEnforceRowLimit_Zero(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
if got := g.EnforceRowLimit(0); got != 10000 {
t.Errorf("expected 10000 for 0 requested, got %d", got)
}
}

func TestEnforceRowLimit_ExceedsMax(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
if got := g.EnforceRowLimit(99999); got != 10000 {
t.Errorf("expected 10000 for 99999 requested, got %d", got)
}
}

func TestEnforceRowLimit_WithinBounds(t *testing.T) {
g := domain.NewQueryGuardrail(90, 10000, 30)
if got := g.EnforceRowLimit(500); got != 500 {
t.Errorf("expected 500, got %d", got)
}
}
