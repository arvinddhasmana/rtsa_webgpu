// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
"time"

commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
entityv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/entity/v1"
inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
"google.golang.org/protobuf/types/known/timestamppb"
)

// ── Test data coordinates: Mid-Atlantic 43-47°N, 55-65°W (UNCLASSIFIED synthetic only) ──

// NewTestRadarObservation builds a valid RadarTrack SensorObservation for testing.
// Position: ~45.0°N, -60.0°W (Mid-Atlantic, synthetic).
func NewTestRadarObservation(sensorID string) *ingestionv1.SensorObservation {
speed := float64(15.0)
heading := float64(180.0)
return &ingestionv1.SensorObservation{
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_RADAR,
ObservationTime: timestamppb.New(time.Now().UTC()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:       45.0,
Longitude:      -60.0,
SpeedKnots:     &speed,
HeadingDegrees: &heading,
},
Metadata: map[string]string{
"test": "true",
},
SensorData: &ingestionv1.SensorObservation_Radar{
Radar: &ingestionv1.RadarTrack{
TrackNumber:    "TRK-001",
RangeNm:        25.0,
BearingDegrees: 090.0,
TrackQuality:   0.9,
},
},
}
}

// NewTestFusedTrack builds a valid FusedTrack for testing.
func NewTestFusedTrack(trackID string) *entityv1.FusedTrack {
return &entityv1.FusedTrack{
TrackId:        trackID,
Classification: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
}
}

// NewTestAnomalyAlert builds a valid AnomalyAlert for testing.
func NewTestAnomalyAlert(alertID, trackID string) *inferencev1.AnomalyAlert {
return &inferencev1.AnomalyAlert{
AlertId:    alertID,
TrackId:    trackID,
DetectedAt: timestamppb.New(time.Now().UTC()),
}
}

// NewTestPosition builds a Position at the given coordinates.
func NewTestPosition(lat, lon float64) *commonv1.Position {
return &commonv1.Position{
Latitude:  lat,
Longitude: lon,
}
}
