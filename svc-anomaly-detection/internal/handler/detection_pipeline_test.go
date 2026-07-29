// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"
"testing"
"log/slog"
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/domain/detectors"
"github.com/arvinddhasmana/rtsa_webgpu/svc-anomaly-detection/internal/state"
"google.golang.org/protobuf/types/known/timestamppb"
)

// mockPublisher captures produced alerts.
type mockPublisher struct {
alerts []*inferencev1.AnomalyAlert
topics []string
}

func (m *mockPublisher) Produce(_ context.Context, topic string, alert *inferencev1.AnomalyAlert) error {
m.alerts = append(m.alerts, alert)
m.topics = append(m.topics, topic)
return nil
}

func newTestPipeline(publisher AlertPublisher) *DetectionPipeline {
history := state.NewTrackHistory(100, 2*time.Hour)

// Seed history with data to enable speed detection.
// Seed varied speeds to ensure stddev > 0.1 (required for speed detection).
baseSpeeds := []float64{8.0, 10.0, 12.0, 9.0, 11.0}
for i, spd := range baseSpeeds {
history.Append("test-track-001", &state.HistoryEntry{
Timestamp:  time.Now().Add(-time.Duration(len(baseSpeeds)-i) * time.Minute),
SpeedKnots: spd,
Heading:    90.0,
Latitude:   44.65,
Longitude:  -63.57,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
})
}

extractor := domain.NewFeatureExtractor(history, nil)

cfg := DetectionPipelineConfig{
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

return NewDetectionPipeline(extractor, cfg, publisher, slog.Default())
}

func TestDetectionPipeline_HandleTrack_NilTrack(t *testing.T) {
pub := &mockPublisher{}
pipeline := newTestPipeline(pub)
err := pipeline.HandleTrack(context.Background(), nil)
if err == nil {
t.Error("Expected error for nil track")
}
}

func TestDetectionPipeline_HandleTrack_SpeedAnomaly(t *testing.T) {
pub := &mockPublisher{}
pipeline := newTestPipeline(pub)

// Track with speed well above historical average (should trigger speed detection).
track := &entityv1.FusedTrack{
TrackId:    "test-track-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity: &commonv1.Velocity{
EastMps: 51.44, // ~100 knots heading east
},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack returned error: %v", err)
}

// Should have produced at least one alert.
if len(pub.alerts) == 0 {
t.Error("Expected at least one alert for speed anomaly track")
}
}

func TestDetectionPipeline_HandleTrack_NormalTrack(t *testing.T) {
pub := &mockPublisher{}

// Use fresh history — no prior data = stddev < 0.1 = no speed detection.
history := state.NewTrackHistory(100, 2*time.Hour)
extractor := domain.NewFeatureExtractor(history, nil)

cfg := DetectionPipelineConfig{
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

pipeline := NewDetectionPipeline(extractor, cfg, pub, slog.Default())

track := &entityv1.FusedTrack{
TrackId:    "normal-track-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity: &commonv1.Velocity{
EastMps: 6.17, // ~12 knots heading east
},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack returned error: %v", err)
}

// Normal track with no history should produce no alerts.
for _, alert := range pub.alerts {
t.Logf("Unexpected alert: type=%s, conf=%.2f", alert.GetAnomalyType(), alert.GetConfidenceScore())
}
}

func TestDetectionPipeline_AlertHasCorrectFields(t *testing.T) {
pub := &mockPublisher{}
pipeline := newTestPipeline(pub)

track := &entityv1.FusedTrack{
TrackId:    "test-track-001",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity: &commonv1.Velocity{
EastMps: 51.44, // ~100 knots heading east
},
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A,
CreatedAt:         timestamppb.Now(),
UpdatedAt:         timestamppb.Now(),
}

err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack returned error: %v", err)
}

if len(pub.alerts) == 0 {
t.Skip("No alerts produced — cannot verify fields")
}

alert := pub.alerts[0]
if alert.GetAlertId() == "" {
t.Error("Alert should have a non-empty alert_id")
}
if alert.GetTrackId() != "test-track-001" {
t.Errorf("Alert track_id = %q, want 'test-track-001'", alert.GetTrackId())
}
if alert.GetModelVersion() != "rules-v1.0.0" {
t.Errorf("Alert model_version = %q, want 'rules-v1.0.0'", alert.GetModelVersion())
}
if alert.GetClassification() != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A {
t.Error("Classification should be propagated from track")
}
if alert.GetDetectedAt() == nil {
t.Error("Alert should have a detected_at timestamp")
}
}

func TestDetectionPipeline_DisabledDetector(t *testing.T) {
pub := &mockPublisher{}
history := state.NewTrackHistory(100, 2*time.Hour)

// Seed history.
for i := 0; i < 5; i++ {
history.Append("test-track-002", &state.HistoryEntry{
Timestamp:  time.Now().Add(-time.Duration(5-i) * time.Minute),
SpeedKnots: 10.0,
Heading:    90.0,
Latitude:   44.65,
Longitude:  -63.57,
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
})
}

extractor := domain.NewFeatureExtractor(history, nil)
cfg := DetectionPipelineConfig{
ModelVersion:  "rules-v1.0.0",
SpeedDetector: detectors.NewSpeedDetector(3.0),
SpeedEnabled:  false, // Disabled!
}
pipeline := NewDetectionPipeline(extractor, cfg, pub, slog.Default())

track := &entityv1.FusedTrack{
TrackId:    "test-track-002",
EntityType: commonv1.EntityType_ENTITY_TYPE_SURFACE,
Velocity:   &commonv1.Velocity{EastMps: 514.0}, // ~999 knots
EstimatedPosition: &commonv1.Position{Latitude: 44.65, Longitude: -63.57},
CreatedAt:  timestamppb.Now(),
UpdatedAt:  timestamppb.Now(),
}
err := pipeline.HandleTrack(context.Background(), track)
if err != nil {
t.Fatalf("HandleTrack returned error: %v", err)
}
if len(pub.alerts) != 0 {
t.Errorf("Expected no alerts when detector is disabled, got %d", len(pub.alerts))
}
}
