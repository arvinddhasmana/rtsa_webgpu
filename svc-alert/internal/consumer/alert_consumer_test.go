// CLASSIFICATION: UNCLASSIFIED
package consumer_test

import (
"context"
"log/slog"
"os"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/consumer"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
"google.golang.org/protobuf/encoding/protojson"
"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/prometheus/client_golang/prometheus"
)

// mockConsumerClient is a test double for ConsumerClient.
// It delivers a fixed set of messages then blocks until ctx is cancelled.
type mockConsumerClient struct {
messages []mockMessage
closed   bool
}

type mockMessage struct {
topic string
key   []byte
value []byte
}

func (m *mockConsumerClient) Consume(ctx context.Context, topics []string, handler consumer.MessageHandler) error {
for _, msg := range m.messages {
if err := handler(ctx, msg.topic, msg.key, msg.value); err != nil {
return err
}
}
// Block until cancelled.
<-ctx.Done()
return nil
}

func (m *mockConsumerClient) Close() error {
m.closed = true
return nil
}

func marshalAlert(t *testing.T, alert *inferencev1.AnomalyAlert) []byte {
t.Helper()
b, err := protojson.Marshal(alert)
if err != nil {
t.Fatalf("marshal alert: %v", err)
}
return b
}

func newTestLogger() *slog.Logger {
return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestAlertConsumer_EnqueuesValidAlert verifies that a valid message is unmarshalled and enqueued.
func TestAlertConsumer_EnqueuesValidAlert(t *testing.T) {
alert := &inferencev1.AnomalyAlert{
AlertId:        "consumer-alert-1",
TrackId:        "track-abc",
Severity:       commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
AnomalyType:    commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
DetectedAt:     timestamppb.New(time.Now()),
}

mock := &mockConsumerClient{
messages: []mockMessage{
{
topic: "alerts.anomaly.critical",
key:   []byte(alert.AlertId),
value: marshalAlert(t, alert),
},
},
}

q := domain.NewAlertQueue(100)
c := consumer.NewAlertConsumer(mock, q, []string{"alerts.anomaly.critical"}, nil, newTestLogger())

ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
defer cancel()

_ = c.Start(ctx) // blocks until ctx cancelled

qa, found := q.Get("consumer-alert-1")
if !found {
t.Fatal("expected alert to be enqueued")
}
if qa.Alert.GetSeverity() != commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL {
t.Errorf("expected CRITICAL severity, got %v", qa.Alert.GetSeverity())
}
}

func TestAlertConsumer_IgnoresEmptyBody(t *testing.T) {
	mock := &mockConsumerClient{
		messages: []mockMessage{
			{topic: "alerts.anomaly.watch", key: []byte("k"), value: []byte{}},
		},
	}

q := domain.NewAlertQueue(100)
c := consumer.NewAlertConsumer(mock, q, nil, nil, newTestLogger())

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

_ = c.Start(ctx)

if q.Size() != 0 {
t.Errorf("expected empty queue, got %d", q.Size())
}
}

func TestAlertConsumer_IgnoresInvalidProto(t *testing.T) {
	mock := &mockConsumerClient{
		messages: []mockMessage{
			{topic: "alerts.anomaly.elevated", key: []byte("k"), value: []byte("not-proto-data")},
		},
	}

q := domain.NewAlertQueue(100)
c := consumer.NewAlertConsumer(mock, q, nil, nil, newTestLogger())

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

_ = c.Start(ctx)

if q.Size() != 0 {
t.Errorf("expected empty queue for invalid proto, got %d", q.Size())
}
}

// TestAlertConsumer_IgnoresMissingAlertID skips alerts with no alert_id.
func TestAlertConsumer_IgnoresMissingAlertID(t *testing.T) {
alert := &inferencev1.AnomalyAlert{
AlertId:  "", // missing
Severity: commonv1.AlertSeverity_ALERT_SEVERITY_WATCH,
}

mock := &mockConsumerClient{
messages: []mockMessage{
{topic: "alerts.anomaly.watch", value: marshalAlert(t, alert)},
},
}

q := domain.NewAlertQueue(100)
c := consumer.NewAlertConsumer(mock, q, nil, nil, newTestLogger())

ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()
_ = c.Start(ctx)

if q.Size() != 0 {
t.Errorf("expected empty queue for missing alert_id, got %d", q.Size())
}
}

// TestAlertConsumer_Close verifies Close propagates to the underlying client.
func TestAlertConsumer_Close(t *testing.T) {
mock := &mockConsumerClient{}
q := domain.NewAlertQueue(100)
c := consumer.NewAlertConsumer(mock, q, nil, nil, newTestLogger())

if err := c.Close(); err != nil {
t.Errorf("unexpected error on Close: %v", err)
}
if !mock.closed {
t.Error("expected underlying client to be closed")
}
}

// ─────────────────────────────────────────────────────────────────────────────
// Additional coverage tests
// ─────────────────────────────────────────────────────────────────────────────

// TestAlertConsumer_WithMetrics exercises the recordMetrics path including
// the anomalyTypeLabel and severityLabel helpers.
func TestAlertConsumer_WithMetrics(t *testing.T) {
alerts := []*inferencev1.AnomalyAlert{
{AlertId: "m-speed", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_SPEED, DetectedAt: timestamppb.New(time.Now())},
{AlertId: "m-route", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION, DetectedAt: timestamppb.New(time.Now())},
{AlertId: "m-ais", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION, DetectedAt: timestamppb.New(time.Now())},
{AlertId: "m-beh", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, DetectedAt: timestamppb.New(time.Now())},
{AlertId: "m-tmp", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL, DetectedAt: timestamppb.New(time.Now())},
{AlertId: "m-prx", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY, DetectedAt: timestamppb.New(time.Now())},
{AlertId: "m-unspec", Severity: commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED, AnomalyType: commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED, DetectedAt: timestamppb.New(time.Now())},
}

var messages []mockMessage
for _, a := range alerts {
messages = append(messages, mockMessage{
topic: "alerts.anomaly.watch",
value: marshalAlert(t, a),
})
}

mock := &mockConsumerClient{messages: messages}
q := domain.NewAlertQueue(100)

metricsObj := &consumer.ConsumerMetrics{
AlertsReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
Name: "test_alerts_received_total",
Help: "test",
}, []string{"severity", "anomaly_type"}),
QueueSize: prometheus.NewGauge(prometheus.GaugeOpts{
Name: "test_queue_size",
Help: "test",
}),
}

c := consumer.NewAlertConsumer(mock, q, nil, metricsObj, newTestLogger())

ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
defer cancel()

_ = c.Start(ctx)

if q.Size() != len(alerts) {
t.Errorf("expected %d alerts in queue, got %d", len(alerts), q.Size())
}
}

// TestFranzConsumerStub verifies the stub consumer returns without panic.
func TestFranzConsumerStub(t *testing.T) {
cl, err := consumer.NewFranzConsumerClient([]string{"localhost:9092"}, "test-group", newTestLogger())
if err != nil {
t.Fatalf("unexpected error creating stub client: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

_ = cl.Consume(ctx, []string{"topic"}, func(_ context.Context, _ string, _, _ []byte) error { return nil })

if err := cl.Close(); err != nil {
t.Errorf("unexpected error on Close: %v", err)
}
}
