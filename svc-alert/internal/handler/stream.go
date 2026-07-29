// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"log/slog"

inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/mapper"
"github.com/prometheus/client_golang/prometheus"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// StreamMetrics holds Prometheus metrics for the StreamAlerts handler.
type StreamMetrics struct {
StreamClients        prometheus.Gauge
AlertsUnacknowledged *prometheus.GaugeVec
}

// StreamHandler implements the StreamAlerts gRPC server-streaming RPC.
type StreamHandler struct {
queue   *domain.AlertQueue
metrics *StreamMetrics
logger  *slog.Logger
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(q *domain.AlertQueue, m *StreamMetrics, logger *slog.Logger) *StreamHandler {
return &StreamHandler{queue: q, metrics: m, logger: logger}
}

// StreamAlerts implements AlertService.StreamAlerts.
//
// Flow:
//  1. Validate and parse the request filters.
//  2. Send all existing unacknowledged alerts that match the filters (initial dump).
//  3. Subscribe to the queue for real-time notifications.
//  4. Forward each matching new alert to the client until disconnect.
func (h *StreamHandler) StreamAlerts(
req *inferencev1.StreamAlertsRequest,
stream inferencev1.AlertService_StreamAlertsServer,
) error {
ctx := stream.Context()

if req == nil {
return status.Error(codes.InvalidArgument, "request must not be nil")
}

// Track connected clients.
if h.metrics != nil && h.metrics.StreamClients != nil {
h.metrics.StreamClients.Inc()
defer h.metrics.StreamClients.Dec()
}

h.logger.InfoContext(ctx, "StreamAlerts client connected",
"min_severity", req.GetMinSeverity().String(),
"clearance_level", req.GetClearanceLevel().String(),
)

// --- Phase 1: Initial dump of existing unacknowledged alerts ---
existingAlerts := h.queue.GetUnacknowledged()
for _, qa := range existingAlerts {
if !mapper.PassesStreamFilter(qa.Alert, req) {
continue
}
if err := stream.Send(mapper.QueuedAlertToProto(qa)); err != nil {
h.logger.WarnContext(ctx, "StreamAlerts initial send failed", "error", err.Error())
return status.Errorf(codes.Internal, "send failed: %v", err)
}
}

// Update unacknowledged gauge after initial dump.
h.updateUnacknowledgedGauge()

// --- Phase 2: Subscribe and forward incremental updates ---
ch := h.queue.Subscribe()
defer h.queue.Unsubscribe(ch)

for {
select {
case <-ctx.Done():
h.logger.InfoContext(ctx, "StreamAlerts client disconnected")
return nil

case alert, ok := <-ch:
if !ok {
// Channel closed — subscriber was removed.
return nil
}
if !mapper.PassesStreamFilter(alert, req) {
continue
}
if err := stream.Send(alert); err != nil {
h.logger.WarnContext(ctx, "StreamAlerts send failed", "error", err.Error())
return status.Errorf(codes.Internal, "send failed: %v", err)
}
h.updateUnacknowledgedGauge()
}
}
}

// updateUnacknowledgedGauge refreshes the unacknowledged count gauge per severity.
func (h *StreamHandler) updateUnacknowledgedGauge() {
if h.metrics == nil || h.metrics.AlertsUnacknowledged == nil {
return
}
counts := h.queue.UnacknowledgedCount()
for sev, count := range counts {
h.metrics.AlertsUnacknowledged.WithLabelValues(domain.SeverityLabel(sev)).Set(float64(count))
}
}
