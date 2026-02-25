// CLASSIFICATION: UNCLASSIFIED
package metrics

import (
"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for svc-alert.
type Metrics struct {
// AlertsReceived is a counter of alerts consumed from Redpanda.
// Labels: severity, anomaly_type
AlertsReceived *prometheus.CounterVec

// AlertsUnacknowledged is a gauge of unacknowledged alerts.
// Labels: severity
AlertsUnacknowledged *prometheus.GaugeVec

// StreamClients is a gauge of connected streaming gRPC clients.
StreamClients prometheus.Gauge

// TimeToAcknowledge is a histogram of time from detection to acknowledgment.
// Labels: severity
TimeToAcknowledge *prometheus.HistogramVec

// QueueSize is a gauge of total alerts in the priority queue.
QueueSize prometheus.Gauge
}

// New registers and returns all svc-alert metrics with the given registerer.
func New(reg prometheus.Registerer) *Metrics {
factory := promauto.With(reg)
return &Metrics{
AlertsReceived: factory.NewCounterVec(prometheus.CounterOpts{
Name: "rtsa_alert_service_alerts_received_total",
Help: "Total number of alerts consumed from Redpanda, by severity and anomaly type.",
}, []string{"severity", "anomaly_type"}),

AlertsUnacknowledged: factory.NewGaugeVec(prometheus.GaugeOpts{
Name: "rtsa_alert_service_alerts_unacknowledged",
Help: "Current number of unacknowledged alerts by severity.",
}, []string{"severity"}),

StreamClients: factory.NewGauge(prometheus.GaugeOpts{
Name: "rtsa_alert_service_stream_clients",
Help: "Current number of active StreamAlerts gRPC streaming clients.",
}),

TimeToAcknowledge: factory.NewHistogramVec(prometheus.HistogramOpts{
Name:    "rtsa_alert_service_time_to_acknowledge_seconds",
Help:    "Time from alert detection to operator acknowledgment in seconds.",
Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1s to ~68min
}, []string{"severity"}),

QueueSize: factory.NewGauge(prometheus.GaugeOpts{
Name: "rtsa_alert_service_queue_size",
Help: "Current total number of alerts in the priority queue.",
}),
}
}
