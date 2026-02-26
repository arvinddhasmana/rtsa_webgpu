// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"fmt"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-fusion-engine/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

// TrackProducer abstracts producing FusedTrack messages (enables test mocking).
type TrackProducer interface {
	Produce(ctx context.Context, track *entityv1.FusedTrack) error
}

// FusionMetrics holds Prometheus metrics for the fusion pipeline.
type FusionMetrics struct {
	ObservationsProcessed *prometheus.CounterVec
	TracksActive          *prometheus.GaugeVec
	CorrelationScore      *prometheus.HistogramVec
	CorrelationDuration   *prometheus.HistogramVec
	TracksCreated         *prometheus.CounterVec
	TracksMerged          *prometheus.CounterVec
	TracksDropped         *prometheus.CounterVec
	KalmanUpdateDuration  prometheus.Histogram
}

// NewFusionMetrics creates and registers Prometheus metrics for the fusion pipeline.
func NewFusionMetrics(reg prometheus.Registerer) *FusionMetrics {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	m := &FusionMetrics{
		ObservationsProcessed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rtsa_fusion_observations_processed_total",
			Help: "Total sensor observations processed.",
		}, []string{"sensor_type"}),
		TracksActive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "rtsa_fusion_tracks_active",
			Help: "Active tracks by entity type and status.",
		}, []string{"entity_type", "status"}),
		CorrelationScore: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rtsa_fusion_correlation_score",
			Help:    "Correlation score distribution.",
			Buckets: prometheus.LinearBuckets(0, 0.1, 11),
		}, []string{"entity_type", "action"}),
		CorrelationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "rtsa_fusion_correlation_duration_seconds",
			Help:    "Time spent on correlation per observation.",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10),
		}, []string{"entity_type"}),
		TracksCreated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rtsa_fusion_tracks_created_total",
			Help: "Total new tracks created.",
		}, []string{"entity_type"}),
		TracksMerged: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rtsa_fusion_tracks_merged_total",
			Help: "Total tracks merged.",
		}, []string{"entity_type"}),
		TracksDropped: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rtsa_fusion_tracks_dropped_total",
			Help: "Total tracks dropped.",
		}, []string{"entity_type"}),
		KalmanUpdateDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "rtsa_fusion_kalman_update_duration_seconds",
			Help:    "Time spent on Kalman filter updates.",
			Buckets: prometheus.ExponentialBuckets(0.00001, 2, 10),
		}),
	}
	reg.MustRegister(
		m.ObservationsProcessed,
		m.TracksActive,
		m.CorrelationScore,
		m.CorrelationDuration,
		m.TracksCreated,
		m.TracksMerged,
		m.TracksDropped,
		m.KalmanUpdateDuration,
	)
	return m
}

// FusionPipeline orchestrates the end-to-end fusion flow for each sensor observation.
type FusionPipeline struct {
	gating    *domain.GatingFilter
	scorer    *domain.CorrelationScorer
	manager   *domain.TrackManager
	producer  TrackProducer
	audit     *audit.Emitter
	logger    *zap.Logger
	metrics   *FusionMetrics
	autoThresh float64
	tentThresh float64
	mergeThresh float64
	gate      domain.GatingConfig // default gate for merge checks
}

// NewFusionPipeline creates a FusionPipeline with all required dependencies.
func NewFusionPipeline(
	gating *domain.GatingFilter,
	scorer *domain.CorrelationScorer,
	manager *domain.TrackManager,
	producer TrackProducer,
	auditEmitter *audit.Emitter,
	logger *zap.Logger,
	metrics *FusionMetrics,
	autoThresh, tentThresh float64,
	gate domain.GatingConfig,
) *FusionPipeline {
	return &FusionPipeline{
		gating:      gating,
		scorer:      scorer,
		manager:     manager,
		producer:    producer,
		audit:       auditEmitter,
		logger:      logger,
		metrics:     metrics,
		autoThresh:  autoThresh,
		tentThresh:  tentThresh,
		mergeThresh: autoThresh,
		gate:        gate,
	}
}

// HandleObservation processes a single Kafka record through the fusion pipeline.
func (fp *FusionPipeline) HandleObservation(ctx context.Context, record *kgo.Record) error {
	var obs ingestionv1.SensorObservation
	if err := protojson.Unmarshal(record.Value, &obs); err != nil {
		return fmt.Errorf("fusion_pipeline: unmarshal: %w", err)
	}

	// Skip observations without position (e.g., Cyber)
	if obs.GetPosition() == nil {
		fp.logger.Debug("skipping positionless observation",
			zap.String("sensor_id", obs.GetSensorId()),
			zap.String("sensor_type", obs.GetSensorType().String()),
		)
		if fp.metrics != nil {
			fp.metrics.ObservationsProcessed.WithLabelValues(obs.GetSensorType().String()).Inc()
		}
		return nil
	}

	correlationStart := time.Now()
	candidates := fp.gating.FindCandidates(&obs, fp.manager.GetActiveTracks())
	etLabel := entityTypeLabel(&obs)

	var track *domain.TrackState
	action := "CREATED"

	if len(candidates) == 0 {
		// No candidates → create new track
		var err error
		track, err = fp.manager.CreateTrack(&obs)
		if err != nil {
			return fmt.Errorf("fusion_pipeline: create track: %w", err)
		}
		if fp.metrics != nil {
			fp.metrics.TracksCreated.WithLabelValues(etLabel).Inc()
		}
	} else {
		// Score all candidates and pick best
		gate := fp.gating.GatingConfigFor(candidates[0].EntityType)
		var best *domain.CorrelationResult
		var bestTrack *domain.TrackState
		for _, candidate := range candidates {
			result := fp.scorer.Score(&obs, candidate, gate)
			if best == nil || result.Score > best.Score {
				best = result
				bestTrack = candidate
			}
			if fp.metrics != nil {
				fp.metrics.CorrelationScore.WithLabelValues(etLabel, result.Action.String()).Observe(result.Score)
			}
		}

		switch {
		case best.Score >= fp.autoThresh:
			action = "UPDATED"
			var err error
			track, err = fp.manager.UpdateTrack(bestTrack.TrackID, &obs)
			if err != nil {
				return fmt.Errorf("fusion_pipeline: update track: %w", err)
			}
		case best.Score >= fp.tentThresh:
			action = "UPDATED"
			var err error
			track, err = fp.manager.UpdateTrack(bestTrack.TrackID, &obs)
			if err != nil {
				return fmt.Errorf("fusion_pipeline: update track (tentative): %w", err)
			}
		default:
			// Score too low → create new track
			var err error
			track, err = fp.manager.CreateTrack(&obs)
			if err != nil {
				return fmt.Errorf("fusion_pipeline: create new track: %w", err)
			}
			action = "CREATED"
			if fp.metrics != nil {
				fp.metrics.TracksCreated.WithLabelValues(etLabel).Inc()
			}
		}
	}

	if fp.metrics != nil {
		fp.metrics.CorrelationDuration.WithLabelValues(etLabel).Observe(time.Since(correlationStart).Seconds())
		fp.metrics.ObservationsProcessed.WithLabelValues(obs.GetSensorType().String()).Inc()
	}

	// Check for merge candidates among all active tracks
	activeTracks := fp.manager.GetActiveTracks()
	if len(activeTracks) >= 2 {
		pairs := domain.FindMergeCandidates(activeTracks, fp.scorer, fp.gate, fp.mergeThresh)
		for _, pair := range pairs {
			if _, err := fp.manager.MergeTracks(pair[0], pair[1]); err == nil {
				fp.emitAudit(ctx, "TRACK_MERGE", pair[0])
				if fp.metrics != nil {
					fp.metrics.TracksMerged.WithLabelValues(etLabel).Inc()
				}
			}
		}
	}

	// Produce the fused track event
	if err := fp.producer.Produce(ctx, track.ToFusedTrack()); err != nil {
		fp.logger.Warn("produce failed", zap.Error(err))
	}

	fp.emitAudit(ctx, "TRACK_LIFECYCLE", track.TrackID)
	fp.logger.Debug("observation processed",
		zap.String("track_id", track.TrackID),
		zap.String("action", action),
	)

	return nil
}

func (fp *FusionPipeline) emitAudit(ctx context.Context, eventType, resourceID string) {
	if fp.audit != nil {
		fp.audit.Emit(ctx, audit.AuditParams{
			EventType:    eventType,
			ResourceType: "FusedTrack",
			ResourceID:   resourceID,
		})
	}
}

func entityTypeLabel(obs *ingestionv1.SensorObservation) string {
	return obs.GetSensorType().String()
}

// OnTrackDropped is a callback for the StaleMonitor to update dropped metrics.
func (fp *FusionPipeline) OnTrackDropped(track *domain.TrackState, _, _ commonv1.TrackStatus) {
	if fp.metrics != nil {
		fp.metrics.TracksDropped.WithLabelValues(track.EntityType.String()).Inc()
	}
}
