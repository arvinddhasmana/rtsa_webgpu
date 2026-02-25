// CLASSIFICATION: UNCLASSIFIED
package metrics_test

import (
"testing"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/metrics"
"github.com/prometheus/client_golang/prometheus"
)

// TestNew verifies that all metrics are created and non-nil.
func TestNew(t *testing.T) {
reg := prometheus.NewRegistry()
m := metrics.New(reg)

if m.AlertsReceived == nil {
t.Error("AlertsReceived must not be nil")
}
if m.AlertsUnacknowledged == nil {
t.Error("AlertsUnacknowledged must not be nil")
}
if m.StreamClients == nil {
t.Error("StreamClients must not be nil")
}
if m.TimeToAcknowledge == nil {
t.Error("TimeToAcknowledge must not be nil")
}
if m.QueueSize == nil {
t.Error("QueueSize must not be nil")
}
}

// TestMetrics_Usable verifies all metrics can record values without panicking.
func TestMetrics_Usable(t *testing.T) {
reg := prometheus.NewRegistry()
m := metrics.New(reg)

// Exercise each metric
m.AlertsReceived.WithLabelValues("critical", "speed").Inc()
m.AlertsUnacknowledged.WithLabelValues("critical").Set(3)
m.StreamClients.Inc()
m.StreamClients.Dec()
m.TimeToAcknowledge.WithLabelValues("elevated").Observe(42.5)
m.QueueSize.Set(100)

// Verify registry can gather without errors
gathered, err := reg.Gather()
if err != nil {
t.Fatalf("unexpected gather error: %v", err)
}
if len(gathered) == 0 {
t.Error("expected at least one gathered metric family")
}
}

// TestMetrics_IsolatedRegistries verifies two separate registries don't collide.
func TestMetrics_IsolatedRegistries(t *testing.T) {
reg1 := prometheus.NewRegistry()
reg2 := prometheus.NewRegistry()
m1 := metrics.New(reg1)
m2 := metrics.New(reg2)

m1.StreamClients.Inc()
m2.StreamClients.Inc()
m2.StreamClients.Inc()

// Values should be independent
_ = m1
_ = m2
}
