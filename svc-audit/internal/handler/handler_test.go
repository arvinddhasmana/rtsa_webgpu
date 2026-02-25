// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
"context"
"testing"
"time"

auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-audit/internal/handler"
"go.uber.org/zap"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

func newTestServer(t *testing.T) *handler.AuditServer {
t.Helper()
guardrail := domain.NewQueryGuardrail(90, 10000, 30)
return handler.NewAuditServerForTest(guardrail, zap.NewNop())
}

func validTimeRange() *commonv1.TimeRange {
now := time.Now()
return &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-1 * time.Hour)),
EndTime:   timestamppb.New(now),
}
}

// TestGetAuditEntry_EmptyAuditID verifies that an empty audit_id returns INVALID_ARGUMENT.
func TestGetAuditEntry_EmptyAuditID(t *testing.T) {
srv := newTestServer(t)
_, err := srv.GetAuditEntry(context.Background(), &auditv1.GetAuditEntryRequest{AuditId: ""})
if err == nil {
t.Fatal("expected error for empty audit_id")
}
st, ok := status.FromError(err)
if !ok || st.Code() != codes.InvalidArgument {
t.Errorf("expected InvalidArgument, got %v", err)
}
}

// TestStreamAuditLog_NilTimeRange verifies that nil time_range returns INVALID_ARGUMENT.
func TestStreamAuditLog_NilTimeRange(t *testing.T) {
srv := newTestServer(t)
req := &auditv1.StreamAuditLogRequest{TimeRange: nil}
err := srv.StreamAuditLogValidateOnly(req)
if err == nil {
t.Fatal("expected error for nil time_range")
}
st, ok := status.FromError(err)
if !ok || st.Code() != codes.InvalidArgument {
t.Errorf("expected InvalidArgument, got %v", err)
}
}

// TestStreamAuditLog_EndBeforeStart verifies that end < start returns INVALID_ARGUMENT.
func TestStreamAuditLog_EndBeforeStart(t *testing.T) {
srv := newTestServer(t)
now := time.Now()
req := &auditv1.StreamAuditLogRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now),
EndTime:   timestamppb.New(now.Add(-1 * time.Hour)),
},
}
err := srv.StreamAuditLogValidateOnly(req)
if err == nil {
t.Fatal("expected error for end before start")
}
st, ok := status.FromError(err)
if !ok || st.Code() != codes.InvalidArgument {
t.Errorf("expected InvalidArgument, got %v", err)
}
}

// TestStreamAuditLog_ExceedsMaxRange verifies that >90 day range returns INVALID_ARGUMENT.
func TestStreamAuditLog_ExceedsMaxRange(t *testing.T) {
srv := newTestServer(t)
now := time.Now()
req := &auditv1.StreamAuditLogRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-91 * 24 * time.Hour)),
EndTime:   timestamppb.New(now),
},
}
err := srv.StreamAuditLogValidateOnly(req)
if err == nil {
t.Fatal("expected error for range exceeding 90 days")
}
st, ok := status.FromError(err)
if !ok || st.Code() != codes.InvalidArgument {
t.Errorf("expected InvalidArgument, got %v", err)
}
}

// TestStreamAuditLog_ValidRange verifies that a valid time range passes validation.
func TestStreamAuditLog_ValidRange(t *testing.T) {
srv := newTestServer(t)
req := &auditv1.StreamAuditLogRequest{
TimeRange: validTimeRange(),
}
err := srv.StreamAuditLogValidateOnly(req)
if err != nil {
t.Errorf("expected no error for valid time range, got %v", err)
}
}

// TestStreamAuditLog_ExactlyMaxRange verifies that exactly 90 days is allowed.
func TestStreamAuditLog_ExactlyMaxRange(t *testing.T) {
srv := newTestServer(t)
now := time.Now()
req := &auditv1.StreamAuditLogRequest{
TimeRange: &commonv1.TimeRange{
StartTime: timestamppb.New(now.Add(-90 * 24 * time.Hour)),
EndTime:   timestamppb.New(now),
},
}
err := srv.StreamAuditLogValidateOnly(req)
if err != nil {
t.Errorf("expected no error for exactly 90 day range, got %v", err)
}
}

// TestGetAuditEntry_BlankAuditID verifies whitespace-only ID is also rejected.
func TestGetAuditEntry_BlankAuditID(t *testing.T) {
srv := newTestServer(t)
_, err := srv.GetAuditEntry(context.Background(), &auditv1.GetAuditEntryRequest{AuditId: ""})
if err == nil {
t.Fatal("expected error for blank audit_id")
}
st, ok := status.FromError(err)
if !ok {
t.Fatalf("expected gRPC status error, got %T: %v", err, err)
}
if st.Code() != codes.InvalidArgument {
t.Errorf("expected InvalidArgument, got %v", st.Code())
}
}
