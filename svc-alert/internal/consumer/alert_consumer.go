// CLASSIFICATION: UNCLASSIFIED
package consumer

import (
"context"
"fmt"
"log/slog"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
"github.com/prometheus/client_golang/prometheus"
"google.golang.org/protobuf/proto"
)

// MessageHandler is called for each message consumed from Redpanda.
// topic is the source topic name, key and value are the raw bytes.
type MessageHandler func(ctx context.Context, topic string, key, value []byte) error

// ConsumerClient is the interface for consuming messages from Redpanda.
// A real implementation (FranzConsumerClient) uses franz-go; tests use mocks.
type ConsumerClient interface {
// Consume starts consuming from the given topics and calls handler for each message.
// It blocks until ctx is cancelled or a fatal error occurs.
Consume(ctx context.Context, topics []string, handler MessageHandler) error
// Close releases underlying resources.
Close() error
}

// ConsumerMetrics holds Prometheus metrics for the consumer.
type ConsumerMetrics struct {
AlertsReceived *prometheus.CounterVec
QueueSize      prometheus.Gauge
}

// AlertConsumer consumes anomaly alerts from Redpanda topics
// and enqueues them into the domain AlertQueue.
type AlertConsumer struct {
client  ConsumerClient
queue   *domain.AlertQueue
topics  []string
metrics *ConsumerMetrics
logger  *slog.Logger
}

// NewAlertConsumer creates a new AlertConsumer.
// metrics may be nil.
func NewAlertConsumer(
client ConsumerClient,
queue *domain.AlertQueue,
topics []string,
metrics *ConsumerMetrics,
logger *slog.Logger,
) *AlertConsumer {
return &AlertConsumer{
client:  client,
queue:   queue,
topics:  topics,
metrics: metrics,
logger:  logger,
}
}

// Start begins consuming from the configured topics.
// Blocks until ctx is cancelled or a fatal error occurs.
func (c *AlertConsumer) Start(ctx context.Context) error {
c.logger.InfoContext(ctx, "alert consumer starting", "topics", c.topics)
err := c.client.Consume(ctx, c.topics, c.handleMessage)
if err != nil && ctx.Err() == nil {
return fmt.Errorf("[consumer].[AlertConsumer.Start]: %w", err)
}
c.logger.InfoContext(ctx, "alert consumer stopped")
return nil
}

// Close shuts down the underlying consumer client.
func (c *AlertConsumer) Close() error {
if err := c.client.Close(); err != nil {
return fmt.Errorf("[consumer].[AlertConsumer.Close]: %w", err)
}
return nil
}

// handleMessage processes a single consumed message.
// It deserializes the protobuf-encoded AnomalyAlert and enqueues it.
func (c *AlertConsumer) handleMessage(ctx context.Context, topic string, key, value []byte) error {
if len(value) == 0 {
c.logger.WarnContext(ctx, "received empty message body", "topic", topic)
return nil
}

var alert inferencev1.AnomalyAlert
if err := proto.Unmarshal(value, &alert); err != nil {
c.logger.ErrorContext(ctx, "failed to unmarshal alert",
"topic", topic,
"error", err.Error(),
)
// Return nil to avoid poisoning the consumer — log and continue.
return nil
}

if alert.GetAlertId() == "" {
c.logger.WarnContext(ctx, "alert missing alert_id; skipping", "topic", topic)
return nil
}

c.queue.Enqueue(&alert)
c.recordMetrics(&alert)

c.logger.DebugContext(ctx, "alert enqueued",
"alert_id", alert.GetAlertId(),
"severity", alert.GetSeverity().String(),
"topic", topic,
)

return nil
}

// recordMetrics updates Prometheus counters after a successful enqueue.
func (c *AlertConsumer) recordMetrics(alert *inferencev1.AnomalyAlert) {
if c.metrics == nil {
return
}
if c.metrics.AlertsReceived != nil {
c.metrics.AlertsReceived.WithLabelValues(
severityLabel(alert.GetSeverity()),
anomalyTypeLabel(alert.GetAnomalyType()),
).Inc()
}
if c.metrics.QueueSize != nil {
c.metrics.QueueSize.Set(float64(c.queue.Size()))
}
}

// severityLabel returns the metric label for an AlertSeverity.
func severityLabel(s commonv1.AlertSeverity) string {
return domain.SeverityLabel(s)
}

// anomalyTypeLabel returns the metric label for an AnomalyType.
func anomalyTypeLabel(t commonv1.AnomalyType) string {
switch t {
case commonv1.AnomalyType_ANOMALY_TYPE_SPEED:
return "speed"
case commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION:
return "route_deviation"
case commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION:
return "ais_manipulation"
case commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL:
return "behavioral"
case commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL:
return "temporal"
case commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY:
return "proximity"
default:
return "unspecified"
}
}
