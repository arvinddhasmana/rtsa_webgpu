// CLASSIFICATION: UNCLASSIFIED

package handler

import (
"context"
"errors"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/audit"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-query/internal/security"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/metadata"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

// --- mocks ---

type mockTracksRepo struct {
resp *queryv1.QueryTracksResponse
err  error
}

func (m *mockTracksRepo) QueryTracks(_ context.Context, _ *queryv1.QueryTracksRequest, _ commonv1.ClassificationLevel) (*queryv1.QueryTracksResponse, error) {
return m.resp, m.err
}

type mockAnomalyRepo struct {
resp *queryv1.QueryAnomaliesResponse
err  error
}

func (m *mockAnomalyRepo) QueryAnomalies(_ context.Context, _ *queryv1.QueryAnomaliesRequest, _ commonv1.ClassificationLevel) (*queryv1.QueryAnomaliesResponse, error) {
return m.resp, m.err
}

type mockAuditRepo struct {
resp *queryv1.QueryAuditLogResponse
err  error
}

func (m *mockAuditRepo) QueryAuditLog(_ context.Context, _ *queryv1.QueryAuditLogRequest, _ commonv1.ClassificationLevel) (*queryv1.QueryAuditLogResponse, error) {
return m.resp, m.err
}

// --- test helpers ---

func makeGuardrail(t *testing.T) *domain.QueryGuardrail {
t.Helper()
g, err := domain.NewQueryGuardrail(30, 100000, 30, 100, 1000)
if err != nil {
t.Fatalf("NewQueryGuardrail: %v", err)
}
return g
}

func makeHandler(
tracks *mockTracksRepo,
anomaly *mockAnomalyRepo,
auditR *mockAuditRepo,
emitter audit.Emitter,
guard *domain.QueryGuardrail,
) *Handler {
return New(tracks, anomaly, auditR, emitter, guard, "svc-query-test")
}

func makeTimeRange(start, end time.Time) *commonv1.TimeRange {
return &commonv1.TimeRange{
StartTime: timestamppb.New(start),
EndTime:   timestamppb.New(end),
}
}

func ctxWithClearance(clearance commonv1.ClassificationLevel) context.Context {
md := metadata.Pairs(security.MetadataKeyClassification, clearance.String())
return metadata.NewIncomingContext(context.Background(), md)
}

// --- QueryTracks tests ---

func TestQueryTracks_nilTimeRange(t *testing.T) {
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))
_, err := h.QueryTracks(context.Background(), &queryv1.QueryTracksRequest{})
assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestQueryTracks_rangeExceeds30Days(t *testing.T) {
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))
now := time.Now()
req := &queryv1.QueryTracksRequest{
TimeRange: makeTimeRange(now.Add(-31*24*time.Hour), now),
}
_, err := h.QueryTracks(context.Background(), req)
assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestQueryTracks_endBeforeStart(t *testing.T) {
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))
now := time.Now()
req := &queryv1.QueryTracksRequest{
TimeRange: makeTimeRange(now, now.Add(-time.Hour)),
}
_, err := h.QueryTracks(context.Background(), req)
assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestQueryTracks_success(t *testing.T) {
now := time.Now()
expectedTracks := []*entityv1.FusedTrack{
{TrackId: "track-001", EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE},
}
repo := &mockTracksRepo{
resp: &queryv1.QueryTracksResponse{
Tracks: expectedTracks,
Pagination: &commonv1.PaginationResponse{TotalCount: 1},
},
}
emitter := &audit.NoopEmitter{}
h := makeHandler(repo, &mockAnomalyRepo{}, &mockAuditRepo{}, emitter, makeGuardrail(t))

ctx := ctxWithClearance(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B)
req := &queryv1.QueryTracksRequest{
TimeRange: makeTimeRange(now.Add(-24*time.Hour), now),
}

resp, err := h.QueryTracks(ctx, req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(resp.Tracks) != 1 {
t.Errorf("expected 1 track, got %d", len(resp.Tracks))
}
// Audit event must have been emitted
if len(emitter.Events) != 1 {
t.Errorf("expected 1 audit event, got %d", len(emitter.Events))
}
if emitter.Events[0].ResourceType != "tracks" {
t.Errorf("audit event ResourceType = %q, want %q", emitter.Events[0].ResourceType, "tracks")
}
}

func TestQueryTracks_repoError(t *testing.T) {
repo := &mockTracksRepo{err: errors.New("clickhouse timeout")}
h := makeHandler(repo, &mockAnomalyRepo{}, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))

now := time.Now()
req := &queryv1.QueryTracksRequest{
TimeRange: makeTimeRange(now.Add(-24*time.Hour), now),
}
_, err := h.QueryTracks(context.Background(), req)
assertGRPCCode(t, err, codes.Internal)
}

func TestQueryTracks_clearanceDefaultsToUnclassified(t *testing.T) {
// No clearance metadata => UNCLASSIFIED
captured := commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED
repo := &mockTracksRepo{}
repo.resp = &queryv1.QueryTracksResponse{
Pagination: &commonv1.PaginationResponse{},
}
emitter := &audit.NoopEmitter{}
h := makeHandler(repo, &mockAnomalyRepo{}, &mockAuditRepo{}, emitter, makeGuardrail(t))

now := time.Now()
// plain context — no metadata
_, err := h.QueryTracks(context.Background(), &queryv1.QueryTracksRequest{
TimeRange: makeTimeRange(now.Add(-time.Hour), now),
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
// Captured clearance in audit event should be UNCLASSIFIED
if len(emitter.Events) > 0 {
captured = emitter.Events[0].ClassificationLevel
}
if captured != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("clearance = %v, want UNCLASSIFIED", captured)
}
}

// --- QueryAnomalies tests ---

func TestQueryAnomalies_nilTimeRange(t *testing.T) {
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))
_, err := h.QueryAnomalies(context.Background(), &queryv1.QueryAnomaliesRequest{})
assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestQueryAnomalies_success(t *testing.T) {
now := time.Now()
expectedAlerts := []*inferencev1.AnomalyAlert{
{AlertId: "alert-001", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL},
}
repo := &mockAnomalyRepo{
resp: &queryv1.QueryAnomaliesResponse{
Alerts: expectedAlerts,
Pagination: &commonv1.PaginationResponse{TotalCount: 1},
},
}
emitter := &audit.NoopEmitter{}
h := makeHandler(&mockTracksRepo{}, repo, &mockAuditRepo{}, emitter, makeGuardrail(t))

ctx := ctxWithClearance(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
req := &queryv1.QueryAnomaliesRequest{
TimeRange: makeTimeRange(now.Add(-24*time.Hour), now),
}

resp, err := h.QueryAnomalies(ctx, req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(resp.Alerts) != 1 {
t.Errorf("expected 1 alert, got %d", len(resp.Alerts))
}
if len(emitter.Events) != 1 || emitter.Events[0].ResourceType != "anomalies" {
t.Errorf("audit event not emitted correctly")
}
}

func TestQueryAnomalies_repoError(t *testing.T) {
repo := &mockAnomalyRepo{err: errors.New("db error")}
h := makeHandler(&mockTracksRepo{}, repo, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))
now := time.Now()
_, err := h.QueryAnomalies(context.Background(), &queryv1.QueryAnomaliesRequest{
TimeRange: makeTimeRange(now.Add(-time.Hour), now),
})
assertGRPCCode(t, err, codes.Internal)
}

// --- QueryAuditLog tests ---

func TestQueryAuditLog_nilTimeRange(t *testing.T) {
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, &mockAuditRepo{}, &audit.NoopEmitter{}, makeGuardrail(t))
_, err := h.QueryAuditLog(context.Background(), &queryv1.QueryAuditLogRequest{})
assertGRPCCode(t, err, codes.InvalidArgument)
}

func TestQueryAuditLog_success(t *testing.T) {
now := time.Now()
expectedEntries := []*queryv1.AuditLogEntry{
{AuditId: "audit-001", ServiceId: "svc-query"},
}
repo := &mockAuditRepo{
resp: &queryv1.QueryAuditLogResponse{
Entries: expectedEntries,
Pagination: &commonv1.PaginationResponse{TotalCount: 1},
},
}
emitter := &audit.NoopEmitter{}
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, repo, emitter, makeGuardrail(t))

ctx := ctxWithClearance(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C)
req := &queryv1.QueryAuditLogRequest{
TimeRange: makeTimeRange(now.Add(-24*time.Hour), now),
}

resp, err := h.QueryAuditLog(ctx, req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(resp.Entries) != 1 {
t.Errorf("expected 1 entry, got %d", len(resp.Entries))
}
// Meta-audit: querying audit log is itself audited
if len(emitter.Events) != 1 || emitter.Events[0].ResourceType != "audit_log" {
t.Errorf("meta-audit event not emitted correctly")
}
}

func TestQueryAuditLog_repoError(t *testing.T) {
repo := &mockAuditRepo{err: errors.New("db error")}
h := makeHandler(&mockTracksRepo{}, &mockAnomalyRepo{}, repo, &audit.NoopEmitter{}, makeGuardrail(t))
now := time.Now()
_, err := h.QueryAuditLog(context.Background(), &queryv1.QueryAuditLogRequest{
TimeRange: makeTimeRange(now.Add(-time.Hour), now),
})
assertGRPCCode(t, err, codes.Internal)
}

// TestAuditFailureDoesNotAbortQuery verifies that an audit emission failure
// does NOT cause the primary query to fail.
func TestAuditFailureDoesNotAbortQuery(t *testing.T) {
now := time.Now()
repo := &mockTracksRepo{
resp: &queryv1.QueryTracksResponse{
Tracks: []*entityv1.FusedTrack{{TrackId: "t1"}},
Pagination: &commonv1.PaginationResponse{TotalCount: 1},
},
}

// errEmitter always fails
errEmitter := &errAuditEmitter{}
h := makeHandler(repo, &mockAnomalyRepo{}, &mockAuditRepo{}, errEmitter, makeGuardrail(t))

resp, err := h.QueryTracks(context.Background(), &queryv1.QueryTracksRequest{
TimeRange: makeTimeRange(now.Add(-time.Hour), now),
})
// Query must succeed even though audit failed
if err != nil {
t.Fatalf("expected no error despite audit failure, got: %v", err)
}
if len(resp.Tracks) != 1 {
t.Errorf("expected 1 track, got %d", len(resp.Tracks))
}
}

type errAuditEmitter struct{}

func (e *errAuditEmitter) Emit(_ context.Context, _ string, _ audit.AuditParams) error {
return errors.New("audit broker unavailable")
}
func (e *errAuditEmitter) Close() error { return nil }

// --- helpers ---

func assertGRPCCode(t *testing.T, err error, wantCode codes.Code) {
t.Helper()
if err == nil {
t.Fatalf("expected error with code %v, got nil", wantCode)
}
st, ok := status.FromError(err)
if !ok {
t.Fatalf("expected gRPC status error, got: %v", err)
}
if st.Code() != wantCode {
t.Errorf("code = %v, want %v: %s", st.Code(), wantCode, st.Message())
}
}
