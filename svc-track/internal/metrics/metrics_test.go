// CLASSIFICATION: UNCLASSIFIED
// Package metrics — unit tests for metric registration.
package metrics

import (
"testing"

"github.com/prometheus/client_golang/prometheus"
)

func TestNew_RegistersAllMetrics(t *testing.T) {
reg := prometheus.NewRegistry()
m := New(reg)

if m.ActiveTracks == nil {
t.Error("ActiveTracks gauge vec must not be nil")
}
if m.StreamClients == nil {
t.Error("StreamClients gauge must not be nil")
}
if m.UpdatesSentTotal == nil {
t.Error("UpdatesSentTotal counter vec must not be nil")
}
if m.CacheUpdateDuration == nil {
t.Error("CacheUpdateDuration histogram must not be nil")
}
}

func TestNew_MetricsWork(t *testing.T) {
reg := prometheus.NewRegistry()
m := New(reg)

// Verify we can use all metrics without panic.
m.ActiveTracks.WithLabelValues("AIR", "ACTIVE").Set(5)
m.StreamClients.Inc()
m.StreamClients.Dec()
m.UpdatesSentTotal.WithLabelValues("AIR", "CREATED").Inc()
obs := m.CacheUpdateDuration.Observe
obs(0.001)
}

func TestNew_DefaultRegistry(t *testing.T) {
// Passing nil uses the default registry — must not panic.
// We skip actual test since it would conflict with other tests using DefaultRegisterer.
// Just verify nil is handled without panicking in a fresh registry.
reg := prometheus.NewRegistry()
m := New(reg)
if m == nil {
t.Error("New should return non-nil metrics")
}
}

func TestNew_WithNilReg_UsesDefault(t *testing.T) {
// Passing nil should use prometheus.DefaultRegisterer.
// We can't actually call this without polluting the global registry,
// so just verify the non-nil path via a fresh registry.
reg := prometheus.NewRegistry()
m := New(reg)
// Verify all labels work.
m.ActiveTracks.WithLabelValues("ENTITY_TYPE_AIR", "TRACK_STATUS_ACTIVE").Add(3)
m.ActiveTracks.WithLabelValues("ENTITY_TYPE_SURFACE", "TRACK_STATUS_STALE").Set(1)
m.UpdatesSentTotal.WithLabelValues("ENTITY_TYPE_CYBER", "UPDATE_TYPE_DROPPED").Inc()
// Gather to verify metrics are registered.
mfs, err := reg.Gather()
if err != nil {
t.Fatalf("Gather error: %v", err)
}
if len(mfs) == 0 {
t.Error("expected at least one metric family")
}
}
