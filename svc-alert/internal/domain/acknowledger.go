// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"context"
"fmt"
"log/slog"
"time"

inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/prometheus/client_golang/prometheus"
)

// AcknowledgerMetrics holds Prometheus metrics used by the Acknowledger.
type AcknowledgerMetrics struct {
// TimeToAcknowledge records time from detection to operator acknowledgment.
TimeToAcknowledge *prometheus.HistogramVec
}

// Acknowledger handles alert acknowledgment business logic,
// including input validation and metric recording.
type Acknowledger struct {
queue   *AlertQueue
metrics *AcknowledgerMetrics
logger  *slog.Logger
}

// NewAcknowledger creates a new Acknowledger.
// metrics may be nil (metrics will be silently skipped).
func NewAcknowledger(queue *AlertQueue, metrics *AcknowledgerMetrics, logger *slog.Logger) *Acknowledger {
return &Acknowledger{
queue:   queue,
metrics: metrics,
logger:  logger,
}
}

// Acknowledge validates the request, marks the alert as acknowledged,
// records the time-to-acknowledge metric, and returns the acknowledgment timestamp.
//
// Returns ErrAlertNotFound wrapped in a descriptive error if the alert ID is unknown.
func (a *Acknowledger) Acknowledge(ctx context.Context, req *inferencev1.AcknowledgeAlertRequest) (*time.Time, error) {
if req.GetAlertId() == "" {
return nil, fmt.Errorf("[domain].[Acknowledger.Acknowledge]: alert_id is required")
}
if req.GetOperatorId() == "" {
return nil, fmt.Errorf("[domain].[Acknowledger.Acknowledge]: operator_id is required")
}

// Snapshot the alert before acknowledgment to capture detectedAt.
qa, ok := a.queue.Get(req.GetAlertId())
if !ok {
return nil, fmt.Errorf("[domain].[Acknowledger.Acknowledge](%s): %w", req.GetAlertId(), ErrAlertNotFound)
}

severity := qa.Alert.GetSeverity()
detectedAt := qa.Alert.GetDetectedAt()

ackedAt, err := a.queue.Acknowledge(req.GetAlertId(), req.GetOperatorId(), req.GetComment())
if err != nil {
return nil, fmt.Errorf("[domain].[Acknowledger.Acknowledge](%s): %w", req.GetAlertId(), err)
}

// Record time-to-acknowledge metric.
if a.metrics != nil && a.metrics.TimeToAcknowledge != nil && detectedAt != nil {
dur := ackedAt.Sub(detectedAt.AsTime()).Seconds()
a.metrics.TimeToAcknowledge.WithLabelValues(SeverityLabel(severity)).Observe(dur)
}

a.logger.InfoContext(ctx, "alert acknowledged",
"alert_id", req.GetAlertId(),
"operator_id", req.GetOperatorId(),
"severity", severity.String(),
)

return ackedAt, nil
}

