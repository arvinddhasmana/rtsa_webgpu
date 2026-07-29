// CLASSIFICATION: UNCLASSIFIED
//go:build integration

package testutil

import (
"fmt"
"math/rand"
"time"

auditv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/audit/v1"
commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"google.golang.org/protobuf/types/known/timestamppb"
)

// Mid-Atlantic operational area bounds (43–47°N, 55–65°W).
const (
MinLat = 43.0
MaxLat = 47.0
MinLon = -65.0
MaxLon = -55.0
)

// MidAtlanticPosition returns a random position within the Mid-Atlantic operational area.
// Uses the provided rand.Rand for deterministic output when seeded.
func MidAtlanticPosition(rng *rand.Rand) *commonv1.Position {
lat := MinLat + rng.Float64()*(MaxLat-MinLat)
lon := MinLon + rng.Float64()*(MaxLon-MinLon)
return &commonv1.Position{
Latitude:  lat,
Longitude: lon,
}
}

// NewSeededRand returns a new deterministic random source with the given seed.
func NewSeededRand(seed int64) *rand.Rand {
return rand.New(rand.NewSource(seed)) //nolint:gosec // test-only, not cryptographic
}

// ValidRadarObservation returns a fully valid radar SensorObservation using a
// fixed Mid-Atlantic position. Uses seed 42 for determinism.
func ValidRadarObservation() *ingestionv1.SensorObservation {
rng := NewSeededRand(42)
pos := MidAtlanticPosition(rng)
return &ingestionv1.SensorObservation{
ObservationId:   fmt.Sprintf("obs-radar-%d", time.Now().UnixNano()),
SensorId:        "RADAR-TEST-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position:        pos,
SensorData: &ingestionv1.SensorObservation_Radar{
Radar: &ingestionv1.RadarTrack{
TrackNumber:    "rdr-001",
RangeNm:        5.4,
BearingDegrees: 45.0,
},
},
}
}

// ValidAISObservation returns a fully valid AIS SensorObservation.
func ValidAISObservation() *ingestionv1.SensorObservation {
rng := NewSeededRand(99)
pos := MidAtlanticPosition(rng)
speedKnots := 12.0
heading := 90.0
pos.SpeedKnots = &speedKnots
pos.HeadingDegrees = &heading
return &ingestionv1.SensorObservation{
ObservationId:   fmt.Sprintf("obs-ais-%d", time.Now().UnixNano()),
SensorId:        "AIS-TEST-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position:        pos,
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:       "123456789",
VesselName: "TEST VESSEL",
NavStatus:  "UNDERWAY",
},
},
}
}

// FusedTrackFixture returns a valid fused track for testing.
func FusedTrackFixture(entityType commonv1.EntityType) *entityv1.FusedTrack {
rng := NewSeededRand(7)
pos := MidAtlanticPosition(rng)
return &entityv1.FusedTrack{
TrackId:         fmt.Sprintf("track-fixture-%d", time.Now().UnixNano()),
EntityType:      entityType,
HostileClass:    commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
Status:          commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
ConfidenceScore: 0.85,
SourceCount:     2,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
EstimatedPosition: pos,
Velocity: &commonv1.Velocity{
NorthMps: 2.0,
EastMps:  3.0,
},
CreatedAt: timestamppb.New(time.Now().UTC()),
UpdatedAt: timestamppb.New(time.Now().UTC()),
}
}

// AnomalyAlertFixture returns a valid anomaly alert for testing.
func AnomalyAlertFixture(severity commonv1.AlertSeverity) *inferencev1.AnomalyAlert {
rng := NewSeededRand(13)
pos := MidAtlanticPosition(rng)
conf := 0.75
switch severity {
case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL:
conf = 0.95
case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED:
conf = 0.80
case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:
conf = 0.60
}
return &inferencev1.AnomalyAlert{
AlertId:         fmt.Sprintf("alert-fixture-%d", time.Now().UnixNano()),
TrackId:         "track-test-001",
AnomalyType:     commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
Severity:        severity,
ConfidenceScore: conf,
Explanation:     "test anomaly alert",
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
TrackPosition:   pos,
DetectedAt:      timestamppb.New(time.Now().UTC()),
ModelVersion:    "rules-v1.0.0",
}
}

// OperatorFeedbackFixture returns a valid operator feedback for testing.
func OperatorFeedbackFixture(feedbackType commonv1.FeedbackType) *feedbackv1.OperatorFeedback {
return &feedbackv1.OperatorFeedback{
FeedbackId:        fmt.Sprintf("fb-fixture-%d", time.Now().UnixNano()),
TrackId:           "track-test-001",
OperatorId:        "operator-test-01",
FeedbackType:      feedbackType,
Justification:     "test feedback",
Classification:    commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
OperatorClearance: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
SubmittedAt:       timestamppb.New(time.Now().UTC()),
}
}

// AuditEventFixture returns a valid audit event for testing.
func AuditEventFixture(eventType, service string) *auditv1.AuditEvent {
return &auditv1.AuditEvent{
AuditId:             fmt.Sprintf("audit-fixture-%d", time.Now().UnixNano()),
ServiceId:           service,
EventType:           eventType,
ActorId:             service,
ActorType:           auditv1.ActorType_ACTOR_TYPE_SERVICE,
ResourceType:        "track",
ResourceId:          "track-test-001",
Action:              auditv1.AuditAction_AUDIT_ACTION_CREATE,
DetailJson:          `{"test": true}`,
ClassificationLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
EventTime:           timestamppb.New(time.Now().UTC()),
}
}
