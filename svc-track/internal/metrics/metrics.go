// CLASSIFICATION: UNCLASSIFIED
// Package metrics defines Prometheus metrics for svc-track.
//
// Metric naming convention: rtsa_track_service_<name>_<unit>
//
// Feature: FEAT-13 Situational Awareness UI
// Requirements: NFR-PERF-001, NFR-AVAIL-001
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus instruments for svc-track.
type Metrics struct {
	// ActiveTracks is a gauge of currently cached active tracks,
	// labelled by entity_type and status.
	ActiveTracks *prometheus.GaugeVec

	// StreamClients is a gauge of currently connected streaming gRPC clients.
	StreamClients prometheus.Gauge

	// UpdatesSentTotal is a counter of TrackUpdate messages sent to clients,
	// labelled by entity_type and update_type.
	UpdatesSentTotal *prometheus.CounterVec

	// CacheUpdateDuration is a histogram of cache Put() latency in seconds.
	CacheUpdateDuration prometheus.Histogram
}

// New registers and returns the service metrics. Panics if registration fails
// (acceptable at init time — indicates programming error, not runtime error).
func New(reg prometheus.Registerer) *Metrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	factory := promauto.With(reg)

	return &Metrics{
		ActiveTracks: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rtsa_track_service_active_tracks",
			Help: "Number of active tracks in the in-memory cache.",
		}, []string{"entity_type", "status"}),

		StreamClients: factory.NewGauge(prometheus.GaugeOpts{
			Name: "rtsa_track_service_stream_clients",
			Help: "Number of currently connected StreamTracks gRPC clients.",
		}),

		UpdatesSentTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "rtsa_track_service_updates_sent_total",
			Help: "Total number of TrackUpdate messages sent to streaming clients.",
		}, []string{"entity_type", "update_type"}),

		CacheUpdateDuration: factory.NewHistogram(prometheus.HistogramOpts{
			Name:    "rtsa_track_service_cache_update_duration_seconds",
			Help:    "Latency of TrackCache.Put() operations in seconds.",
			Buckets: prometheus.DefBuckets,
		}),
	}
}
