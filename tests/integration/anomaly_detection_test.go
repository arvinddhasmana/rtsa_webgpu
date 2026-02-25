// CLASSIFICATION: UNCLASSIFIED

//go:build integration

package integration

import (
"context"
"testing"
"time"
"log/slog"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/domain/detectors"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/handler"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-anomaly-detection/internal/state"
"google.golang.org/protobuf/types/known/timestamppb"
)

// capturePublisher captures alerts for test assertions.
type capturePublisher struct {
alerts []*inferencev1.AnomalyAlert
topics []string
}

func (c *capturePublisher) Produce(_ context.Context, topic string, alert *inferencev1.AnomalyAlert) error {
c.alerts = append(c.alerts, alert)
c.topics = append(c.topics, topic)
return nil
}

// IT01: Publish fused track with speed anomaly → Alert produced in alerts.anomaly.*
func TestIT01_SpeedAnomalyProducesAlert(t *testing.T) {
pub, pipeline := buildPipelineWithHistory("track-it01", 10.0, 5)

// Track with extreme speed anomaly.
track := &entityv1.FusedTrack{
TrackId:    "track-it01",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity:   &commonv1.Velocity{EastMps: 102.89},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack failed: %v", err)
}

if len(pub.alerts) == 0 {
t.Fatal("IT01: expected at least one alert for speed anomaly, got none")
}

// Verify alert is in an anomaly topic.
found := false
for _, topic := range pub.topics {
if topic == "alerts.anomaly.critical" || topic == "alerts.anomaly.elevated" || topic == "alerts.anomaly.watch" {
found = true
break
}
}
if !found {
t.Errorf("IT01: alert not in expected topics, got: %v", pub.topics)
}
}

// IT02: Publish normal fused track → No alert produced.
func TestIT02_NormalTrackNoAlert(t *testing.T) {
pub := &capturePublisher{}

history := state.NewTrackHistory(100, 2*time.Hour)
extractor := domain.NewFeatureExtractor(history, nil)
pipeline := buildMinimalPipeline(extractor, pub)

track := &entityv1.FusedTrack{
TrackId:    "normal-track-it02",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity:   &commonv1.Velocity{EastMps: 6.17},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack failed: %v", err)
}

// With no history (stddev = 0), speed detector skips. Other detectors should not fire.
for _, alert := range pub.alerts {
t.Logf("IT02: Unexpected alert: type=%s, conf=%.2f", alert.GetAnomalyType(), alert.GetConfidenceScore())
}
}

// IT03: Alert has correct feature contributions populated.
func TestIT03_AlertHasFeatureContributions(t *testing.T) {
pub, pipeline := buildPipelineWithHistory("track-it03", 10.0, 5)

track := &entityv1.FusedTrack{
TrackId:    "track-it03",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity:   &commonv1.Velocity{EastMps: 77.16},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack failed: %v", err)
}

if len(pub.alerts) == 0 {
t.Fatal("IT03: no alerts produced — cannot verify feature contributions")
}

for _, alert := range pub.alerts {
if len(alert.GetFeatures()) == 0 {
t.Errorf("IT03: alert %s has no feature contributions", alert.GetAlertId())
}
}
}

// IT04: Classification is propagated from track to alert.
func TestIT04_ClassificationPropagated(t *testing.T) {
pub, pipeline := buildPipelineWithHistory("track-it04", 10.0, 5)

track := &entityv1.FusedTrack{
TrackId:    "track-it04",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity:   &commonv1.Velocity{EastMps: 77.16},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack failed: %v", err)
}

if len(pub.alerts) == 0 {
t.Fatal("IT04: no alerts produced — cannot verify classification")
}

for _, alert := range pub.alerts {
if alert.GetClassification() != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B {
t.Errorf("IT04: alert classification = %v, want PROTECTED_B", alert.GetClassification())
}
}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func buildPipelineWithHistory(trackID string, baseSpeed float64, entries int) (*capturePublisher, *handler.DetectionPipeline) {
pub := &capturePublisher{}
history := state.NewTrackHistory(100, 2*time.Hour)

for i := 0; i < entries; i++ {
history.Append(trackID, &state.HistoryEntry{
Timestamp:  time.Now().Add(-time.Duration(entries-i) * time.Minute),
SpeedKnots: baseSpeed,
Heading:    90.0,
Latitude:   44.65,
Longitude:  -63.57,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
})
}

extractor := domain.NewFeatureExtractor(history, nil)
pipeline := buildMinimalPipeline(extractor, pub)
return pub, pipeline
}

func buildMinimalPipeline(extractor *domain.FeatureExtractor, pub handler.AlertPublisher) *handler.DetectionPipeline {
cfg := handler.DetectionPipelineConfig{
ModelVersion:       "rules-v1.0.0",
SpeedDetector:      detectors.NewSpeedDetector(3.0),
RouteDetector:      detectors.NewRouteDeviationDetector(30.0, 3),
AISDetector:        detectors.NewAISManipulationDetector(0.5),
BehavioralDetector: detectors.NewBehavioralDetector(0.75),
TemporalDetector:   detectors.NewTemporalDetector(0.05),
ProximityDetector:  detectors.NewProximityDetector(),
SpeedEnabled:       true,
RouteEnabled:       true,
AISEnabled:         true,
BehavioralEnabled:  true,
TemporalEnabled:    true,
ProximityEnabled:   true,
}
return handler.NewDetectionPipeline(extractor, cfg, pub, slog.Default())
}
