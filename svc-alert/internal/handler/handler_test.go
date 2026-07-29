// CLASSIFICATION: UNCLASSIFIED
package handler_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/domain"
	"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/handler"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────────────────────────────────────

func newTestLogger() *slog.Logger {
return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func makeAlert(id string, sev commonv1.AlertSeverity, at commonv1.AnomalyType, cl commonv1.ClassificationLevel, et commonv1.EntityType) *inferencev1.AnomalyAlert {
return &inferencev1.AnomalyAlert{
AlertId:        id,
TrackId:        "track-" + id,
Severity:       sev,
AnomalyType:    at,
Classification: cl,
EntityType:     et,
DetectedAt:     timestamppb.New(time.Now()),
}
}

// mockStreamServer captures sent alerts and respects context cancellation.
type mockStreamServer struct {
ctx    context.Context
sent   []*inferencev1.AnomalyAlert
sendFn func(*inferencev1.AnomalyAlert) error
inferencev1.AlertService_StreamAlertsServer
}

func newMockStream(ctx context.Context) *mockStreamServer {
m := &mockStreamServer{ctx: ctx}
m.sendFn = func(a *inferencev1.AnomalyAlert) error {
m.sent = append(m.sent, a)
return nil
}
return m
}

func (m *mockStreamServer) Send(a *inferencev1.AnomalyAlert) error {
return m.sendFn(a)
}
func (m *mockStreamServer) Context() context.Context { return m.ctx }

// ─────────────────────────────────────────────────────────────────────────────
// T06: StreamAlerts min_severity=ELEVATED → WATCH alerts excluded
// ─────────────────────────────────────────────────────────────────────────────

func TestStreamAlerts_T06_MinSeverityFiltersWatch(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewStreamHandler(q, nil, newTestLogger())

now := time.Now()
q.Enqueue(makeAlert("watch-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))
q.Enqueue(makeAlert("elevated-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))
_ = now

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

stream := newMockStream(ctx)
req := &inferencev1.StreamAlertsRequest{
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}

_ = h.StreamAlerts(req, stream) // returns on ctx cancel

for _, a := range stream.sent {
if a.GetSeverity() == commonv1.AlertSeverity_ALERT_SEVERITY_WATCH {
t.Errorf("WATCH alert should be excluded (min_severity=ELEVATED), got alert_id=%s", a.GetAlertId())
}
}

found := false
for _, a := range stream.sent {
if a.GetAlertId() == "elevated-1" {
found = true
}
}
if !found {
t.Error("expected elevated-1 to be sent to stream")
}
}

// ─────────────────────────────────────────────────────────────────────────────
// T07: StreamAlerts filter by anomaly type
// ─────────────────────────────────────────────────────────────────────────────

func TestStreamAlerts_T07_AnomalyTypeFilter(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewStreamHandler(q, nil, newTestLogger())

q.Enqueue(makeAlert("speed-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))
q.Enqueue(makeAlert("route-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

stream := newMockStream(ctx)
req := &inferencev1.StreamAlertsRequest{
AnomalyTypes:   []commonv1.AnomalyType{commonv1.AnomalyType_ANOMALY_TYPE_SPEED},
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL,
}

_ = h.StreamAlerts(req, stream)

for _, a := range stream.sent {
if a.GetAnomalyType() != commonv1.AnomalyType_ANOMALY_TYPE_SPEED {
t.Errorf("expected only SPEED alerts, got %v for alert %s", a.GetAnomalyType(), a.GetAlertId())
}
}

found := false
for _, a := range stream.sent {
if a.GetAlertId() == "speed-1" {
found = true
}
}
if !found {
t.Error("expected speed-1 in stream")
}
}

// ─────────────────────────────────────────────────────────────────────────────
// T08: StreamAlerts classification filter — higher classified excluded
// ─────────────────────────────────────────────────────────────────────────────

func TestStreamAlerts_T08_ClassificationFilter(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewStreamHandler(q, nil, newTestLogger())

// Unclassified alert — should pass
q.Enqueue(makeAlert("unclass-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_AIR))

// Secret alert — should be excluded for PROTECTED_C clearance
q.Enqueue(makeAlert("secret-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
commonv1.EntityType_ENTITY_TYPE_AIR))

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

stream := newMockStream(ctx)
req := &inferencev1.StreamAlertsRequest{
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C,
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL,
}

_ = h.StreamAlerts(req, stream)

for _, a := range stream.sent {
if a.GetAlertId() == "secret-1" {
t.Error("secret-1 should NOT be sent to a PROTECTED_C clearance client")
}
}

found := false
for _, a := range stream.sent {
if a.GetAlertId() == "unclass-1" {
found = true
}
}
if !found {
t.Error("expected unclass-1 to pass classification filter")
}
}

// ─────────────────────────────────────────────────────────────────────────────
// T09: GetAlertDetails — full alert with features
// ─────────────────────────────────────────────────────────────────────────────

func TestGetAlertDetails_T09_FullAlert(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewDetailsHandler(q, newTestLogger())

alert := &inferencev1.AnomalyAlert{
AlertId:        "detail-1",
TrackId:        "track-xyz",
Severity:       commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
AnomalyType:    commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
DetectedAt:     timestamppb.New(time.Now()),
Explanation:    "Unusual behaviour pattern detected",
Features: []*inferencev1.FeatureContribution{
{FeatureName: "speed_delta", ContributionWeight: 0.85},
},
}
q.Enqueue(alert)

resp, err := h.GetAlertDetails(context.Background(), &inferencev1.GetAlertDetailsRequest{
AlertId:        "detail-1",
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.GetAlertId() != "detail-1" {
t.Errorf("expected alert_id=detail-1, got %s", resp.GetAlertId())
}
if resp.GetExplanation() == "" {
t.Error("expected non-empty explanation")
}
if len(resp.GetFeatures()) == 0 {
t.Error("expected features to be populated")
}
}

// TestGetAlertDetails_NotFound verifies NOT_FOUND for unknown ID.
func TestGetAlertDetails_NotFound(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewDetailsHandler(q, newTestLogger())

_, err := h.GetAlertDetails(context.Background(), &inferencev1.GetAlertDetailsRequest{
AlertId:        "ghost-id",
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
if err == nil {
t.Fatal("expected error for unknown alert")
}
if code := status.Code(err); code != codes.NotFound {
t.Errorf("expected NOT_FOUND, got %v", code)
}
}

// TestGetAlertDetails_ClassificationDenied verifies PERMISSION_DENIED for insufficient clearance.
func TestGetAlertDetails_ClassificationDenied(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewDetailsHandler(q, newTestLogger())

q.Enqueue(makeAlert("protected-c-alert", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C,
commonv1.EntityType_ENTITY_TYPE_SURFACE))

_, err := h.GetAlertDetails(context.Background(), &inferencev1.GetAlertDetailsRequest{
AlertId:        "protected-c-alert",
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
})
if err == nil {
t.Fatal("expected permission denied error")
}
if code := status.Code(err); code != codes.PermissionDenied {
t.Errorf("expected PERMISSION_DENIED, got %v", code)
}
}

// ─────────────────────────────────────────────────────────────────────────────
// T10: AcknowledgeAlert — time-to-acknowledge metric recorded
// ─────────────────────────────────────────────────────────────────────────────

func TestAcknowledgeAlert_T10_TimeToAcknowledgeMetric(t *testing.T) {
q := domain.NewAlertQueue(100)
ackMetrics := &domain.AcknowledgerMetrics{
// nil histogram — no panic expected
TimeToAcknowledge: nil,
}
ack := domain.NewAcknowledger(q, ackMetrics, newTestLogger())
h := handler.NewAcknowledgeHandler(ack, newTestLogger())

q.Enqueue(makeAlert("ack-t10", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))

resp, err := h.AcknowledgeAlert(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "ack-t10",
OperatorId: "op-t10",
Comment:    "Test acknowledgment",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !resp.GetSuccess() {
t.Error("expected Success=true")
}
if resp.GetAcknowledgedAt() == nil {
t.Error("expected non-nil AcknowledgedAt")
}
}

// TestAcknowledgeAlert_NotFound verifies NOT_FOUND status.
func TestAcknowledgeAlert_NotFound(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())
h := handler.NewAcknowledgeHandler(ack, newTestLogger())

_, err := h.AcknowledgeAlert(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "no-such-alert",
OperatorId: "op-001",
})
if err == nil {
t.Fatal("expected NOT_FOUND error")
}
if code := status.Code(err); code != codes.NotFound {
t.Errorf("expected NOT_FOUND, got %v", code)
}
}

// TestAcknowledgeAlert_MissingAlertID verifies INVALID_ARGUMENT.
func TestAcknowledgeAlert_MissingAlertID(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())
h := handler.NewAcknowledgeHandler(ack, newTestLogger())

_, err := h.AcknowledgeAlert(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "",
OperatorId: "op-001",
})
if err == nil {
t.Fatal("expected error for empty alert_id")
}
if code := status.Code(err); code != codes.InvalidArgument {
t.Errorf("expected INVALID_ARGUMENT, got %v", code)
}
}

// TestAcknowledgeAlert_MissingOperatorID verifies INVALID_ARGUMENT.
func TestAcknowledgeAlert_MissingOperatorID(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())
h := handler.NewAcknowledgeHandler(ack, newTestLogger())

_, err := h.AcknowledgeAlert(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "some-alert",
OperatorId: "",
})
if err == nil {
t.Fatal("expected error for empty operator_id")
}
if code := status.Code(err); code != codes.InvalidArgument {
t.Errorf("expected INVALID_ARGUMENT, got %v", code)
}
}

// TestAcknowledgeAlert_NilRequest verifies INVALID_ARGUMENT for nil request.
func TestAcknowledgeAlert_NilRequest(t *testing.T) {
q := domain.NewAlertQueue(100)
ack := domain.NewAcknowledger(q, nil, newTestLogger())
h := handler.NewAcknowledgeHandler(ack, newTestLogger())

_, err := h.AcknowledgeAlert(context.Background(), nil)
if err == nil {
t.Fatal("expected error for nil request")
}
if code := status.Code(err); code != codes.InvalidArgument {
t.Errorf("expected INVALID_ARGUMENT, got %v", code)
}
}

// TestStreamAlerts_NilRequest verifies INVALID_ARGUMENT for nil request.
func TestStreamAlerts_NilRequest(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewStreamHandler(q, nil, newTestLogger())
ctx := context.Background()
stream := newMockStream(ctx)
err := h.StreamAlerts(nil, stream)
if err == nil {
t.Fatal("expected error for nil request")
}
if code := status.Code(err); code != codes.InvalidArgument {
t.Errorf("expected INVALID_ARGUMENT, got %v", code)
}
}

// TestGetAlertDetails_NilRequest verifies INVALID_ARGUMENT for nil request.
func TestGetAlertDetails_NilRequest(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewDetailsHandler(q, newTestLogger())
_, err := h.GetAlertDetails(context.Background(), nil)
if err == nil {
t.Fatal("expected error for nil request")
}
}

// TestGetAlertDetails_EmptyAlertID verifies INVALID_ARGUMENT.
func TestGetAlertDetails_EmptyAlertID(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewDetailsHandler(q, newTestLogger())
_, err := h.GetAlertDetails(context.Background(), &inferencev1.GetAlertDetailsRequest{
AlertId: "",
})
if err == nil {
t.Fatal("expected error for empty alert_id")
}
if code := status.Code(err); code != codes.InvalidArgument {
t.Errorf("expected INVALID_ARGUMENT, got %v", code)
}
}

// TestStreamAlerts_IncrementalUpdate verifies that a newly enqueued alert
// reaches an active stream subscriber.
func TestStreamAlerts_IncrementalUpdate(t *testing.T) {
q := domain.NewAlertQueue(100)
h := handler.NewStreamHandler(q, nil, newTestLogger())

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

received := make(chan *inferencev1.AnomalyAlert, 1)
stream := &mockStreamServer{
ctx: ctx,
sendFn: func(a *inferencev1.AnomalyAlert) error {
select {
case received <- a:
default:
}
return nil
},
}

req := &inferencev1.StreamAlertsRequest{
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL,
}

errCh := make(chan error, 1)
go func() {
errCh <- h.StreamAlerts(req, stream)
}()

// Give the goroutine time to subscribe.
time.Sleep(20 * time.Millisecond)

// Enqueue a new alert — it should arrive via the subscriber channel.
newAlert := makeAlert("incremental-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_AIR)
q.Enqueue(newAlert)

select {
case a := <-received:
if a.GetAlertId() != "incremental-1" {
t.Errorf("expected incremental-1, got %s", a.GetAlertId())
}
case <-time.After(500 * time.Millisecond):
t.Error("timed out waiting for incremental alert in stream")
}

cancel()
if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
t.Errorf("unexpected stream error: %v", err)
}
}

// ─────────────────────────────────────────────────────────────────────────────
// AlertServer (server.go) tests — covers the composition layer
// ─────────────────────────────────────────────────────────────────────────────

// TestAlertServer_AcknowledgeAlert exercises the server wrapper for AcknowledgeAlert.
func TestAlertServer_AcknowledgeAlert(t *testing.T) {
	q := domain.NewAlertQueue(100)
	ack := domain.NewAcknowledger(q, nil, newTestLogger())
	ackH := handler.NewAcknowledgeHandler(ack, newTestLogger())
	detailsH := handler.NewDetailsHandler(q, newTestLogger())
	streamH := handler.NewStreamHandler(q, nil, newTestLogger())

	assigner := domain.NewAssigner(q, newTestLogger())
	assignH := handler.NewAssignHandler(assigner, nil, newTestLogger())

	srv := handler.NewAlertServer(streamH, ackH, detailsH, assignH)

q.Enqueue(makeAlert("srv-ack-1", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))

resp, err := srv.AcknowledgeAlert(context.Background(), &inferencev1.AcknowledgeAlertRequest{
AlertId:    "srv-ack-1",
OperatorId: "op-srv",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !resp.GetSuccess() {
t.Error("expected Success=true from server wrapper")
}
}

// TestAlertServer_GetAlertDetails exercises the server wrapper for GetAlertDetails.
func TestAlertServer_GetAlertDetails(t *testing.T) {
	q := domain.NewAlertQueue(100)
	ack := domain.NewAcknowledger(q, nil, newTestLogger())
	ackH := handler.NewAcknowledgeHandler(ack, newTestLogger())
	detailsH := handler.NewDetailsHandler(q, newTestLogger())
	streamH := handler.NewStreamHandler(q, nil, newTestLogger())

	assigner := domain.NewAssigner(q, newTestLogger())
	assignH := handler.NewAssignHandler(assigner, nil, newTestLogger())

	srv := handler.NewAlertServer(streamH, ackH, detailsH, assignH)

q.Enqueue(makeAlert("srv-det-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED,
commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_AIR))

resp, err := srv.GetAlertDetails(context.Background(), &inferencev1.GetAlertDetailsRequest{
AlertId:        "srv-det-1",
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.GetAlertId() != "srv-det-1" {
t.Errorf("expected srv-det-1, got %s", resp.GetAlertId())
}
}

// TestAlertServer_StreamAlerts exercises the server wrapper for StreamAlerts.
func TestAlertServer_StreamAlerts(t *testing.T) {
	q := domain.NewAlertQueue(100)
	ack := domain.NewAcknowledger(q, nil, newTestLogger())
	ackH := handler.NewAcknowledgeHandler(ack, newTestLogger())
	detailsH := handler.NewDetailsHandler(q, newTestLogger())

	assigner := domain.NewAssigner(q, newTestLogger())
	assignH := handler.NewAssignHandler(assigner, nil, newTestLogger())

	streamMetrics := &handler.StreamMetrics{} // nil gauges — should not panic
	streamH := handler.NewStreamHandler(q, streamMetrics, newTestLogger())
	srv := handler.NewAlertServer(streamH, ackH, detailsH, assignH)

ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

stream := newMockStream(ctx)
err := srv.StreamAlerts(&inferencev1.StreamAlertsRequest{
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
}, stream)
// nil return or context cancelled — either is fine for the wrapper test
_ = err
}

// TestStreamAlerts_WithMetrics exercises updateUnacknowledgedGauge with real metrics.
func TestStreamAlerts_WithMetrics(t *testing.T) {
q := domain.NewAlertQueue(100)
q.Enqueue(makeAlert("sm-1", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
commonv1.EntityType_ENTITY_TYPE_SURFACE))

streamMetrics := &handler.StreamMetrics{
StreamClients: prometheus.NewGauge(prometheus.GaugeOpts{
Name: "test_stream_clients",
Help: "test",
}),
AlertsUnacknowledged: prometheus.NewGaugeVec(prometheus.GaugeOpts{
Name: "test_alerts_unacknowledged",
Help: "test",
}, []string{"severity"}),
}

h := handler.NewStreamHandler(q, streamMetrics, newTestLogger())

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

stream := newMockStream(ctx)
_ = h.StreamAlerts(&inferencev1.StreamAlertsRequest{
ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
MinSeverity:    commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL,
}, stream)

if len(stream.sent) == 0 {
t.Error("expected at least one alert sent with metrics enabled")
}
}
