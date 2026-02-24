// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"
"fmt"
"log/slog"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain/detectors"
"github.com/google/uuid"
"google.golang.org/protobuf/types/known/timestamppb"
)

const modelVersion = "rules-v1.0.0"

// AlertPublisher is the interface for publishing anomaly alerts.
type AlertPublisher interface {
Produce(ctx context.Context, topic string, alert *inferencev1.AnomalyAlert) error
}

// detectorEntry pairs an anomaly type with its detector.
type detectorEntry struct {
anomalyType commonv1.AnomalyType
detector    detectors.Detector
enabled     bool
}

// DetectionPipeline orchestrates the full anomaly detection flow:
// FusedTrack → FeatureExtractor → Detectors → Severity → Explainer → AlertProducer
type DetectionPipeline struct {
extractor  *domain.FeatureExtractor
detectors  []detectorEntry
publisher  AlertPublisher
logger     *slog.Logger
modelVer   string
}

// DetectionPipelineConfig configures the detection pipeline.
type DetectionPipelineConfig struct {
ModelVersion        string
SpeedDetector       *detectors.SpeedDetector
RouteDetector       *detectors.RouteDeviationDetector
AISDetector         *detectors.AISManipulationDetector
BehavioralDetector  *detectors.BehavioralDetector
TemporalDetector    *detectors.TemporalDetector
ProximityDetector   *detectors.ProximityDetector
SpeedEnabled        bool
RouteEnabled        bool
AISEnabled          bool
BehavioralEnabled   bool
TemporalEnabled     bool
ProximityEnabled    bool
}

// NewDetectionPipeline creates a new DetectionPipeline.
func NewDetectionPipeline(
extractor *domain.FeatureExtractor,
cfg DetectionPipelineConfig,
publisher AlertPublisher,
logger *slog.Logger,
) *DetectionPipeline {
mv := cfg.ModelVersion
if mv == "" {
mv = modelVersion
}
return &DetectionPipeline{
extractor: extractor,
publisher: publisher,
logger:    logger,
modelVer:  mv,
detectors: []detectorEntry{
{commonv1.AnomalyType_ANOMALY_TYPE_SPEED, cfg.SpeedDetector, cfg.SpeedEnabled},
{commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION, cfg.RouteDetector, cfg.RouteEnabled},
{commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION, cfg.AISDetector, cfg.AISEnabled},
{commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, cfg.BehavioralDetector, cfg.BehavioralEnabled},
{commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL, cfg.TemporalDetector, cfg.TemporalEnabled},
{commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY, cfg.ProximityDetector, cfg.ProximityEnabled},
},
}
}

// HandleTrack processes a single fused track through the detection pipeline.
func (dp *DetectionPipeline) HandleTrack(ctx context.Context, track *entityv1.FusedTrack) error {
if track == nil {
return fmt.Errorf("[handler.DetectionPipeline.HandleTrack]: track is nil")
}

// Extract features.
fv, err := dp.extractor.Extract(track)
if err != nil {
return fmt.Errorf("[handler.DetectionPipeline.HandleTrack](%s): feature extraction: %w",
track.GetTrackId(), err)
}

// Run all enabled detectors.
for _, entry := range dp.detectors {
if !entry.enabled || entry.detector == nil {
continue
}

result := entry.detector.Detect(fv)
if !result.Detected {
continue
}

// Map confidence to severity.
severity := domain.MapSeverity(result.Confidence)
topic := domain.SeverityTopic(severity)
if topic == "" {
// Below threshold — no alert.
continue
}

// Generate explanation.
explanation := domain.GenerateExplanation(entry.anomalyType, fv, result.Confidence)

// Build position from fused track.
var trackPos *commonv1.Position
if pos := track.GetEstimatedPosition(); pos != nil {
trackPos = pos
}

// Build the alert proto.
alert := &inferencev1.AnomalyAlert{
AlertId:         uuid.New().String(),
TrackId:         track.GetTrackId(),
AnomalyType:     entry.anomalyType,
Severity:        severity,
ConfidenceScore: result.Confidence,
Explanation:     explanation,
Features:        result.Features,
Classification:  track.GetClassification(),
DetectedAt:      timestamppb.New(time.Now().UTC()),
ModelVersion:    dp.modelVer,
TrackPosition:   trackPos,
EntityType:      track.GetEntityType(),
}

// Produce alert.
if err := dp.publisher.Produce(ctx, topic, alert); err != nil {
dp.logger.Error("failed to produce alert",
"alert_id", alert.GetAlertId(),
"track_id", track.GetTrackId(),
"anomaly_type", entry.anomalyType.String(),
"error", err,
)
// Continue processing remaining detectors even if one publish fails.
continue
}

dp.logger.Info("anomaly alert produced",
"track_id", track.GetTrackId(),
"anomaly_type", entry.anomalyType.String(),
"severity", severity.String(),
"confidence", result.Confidence,
)
}

return nil
}
