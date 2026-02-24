// CLASSIFICATION: UNCLASSIFIED
package sensor

import (
"fmt"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/generator"
"github.com/google/uuid"
"google.golang.org/protobuf/types/known/timestamppb"
)

// isrSensorNames lists valid ISR sensor modalities.
var isrSensorNames = []string{"EO", "IR", "SAR", "MTI"}

// isrImageCounter provides sequential image IDs per sensor.
var isrImageCounter = make(map[string]int)

// isrPlatformIDs lists simulated ISR platform identifiers.
var isrPlatformIDs = []string{
"ISR-PLATFORM-ALPHA",
"ISR-PLATFORM-BRAVO",
"ISR-PLATFORM-CHARLIE",
}

// GenerateISRObservation creates a SensorObservation with an ISRObservation payload.
// A 4-vertex coverage polygon centred on the entity is generated.
// All generated observations are CLASSIFICATION_LEVEL_UNCLASSIFIED.
func GenerateISRObservation(entity *generator.SimEntity, sensorID string) *ingestionv1.SensorObservation {
isrImageCounter[sensorID]++
imageID := fmt.Sprintf("IMG-%s-%08d", sensorID, isrImageCounter[sensorID])

platformIdx := rng().Intn(len(isrPlatformIDs))
platformID := isrPlatformIDs[platformIdx]
sensorName := isrSensorNames[rng().Intn(len(isrSensorNames))]

// Coverage polygon: ±0.05° lat/lon around entity (~3-5 km box).
halfDeg := 0.05
polygon := []*commonv1.Position{
{Latitude: entity.Position.Lat - halfDeg, Longitude: entity.Position.Lon - halfDeg},
{Latitude: entity.Position.Lat + halfDeg, Longitude: entity.Position.Lon - halfDeg},
{Latitude: entity.Position.Lat + halfDeg, Longitude: entity.Position.Lon + halfDeg},
{Latitude: entity.Position.Lat - halfDeg, Longitude: entity.Position.Lon + halfDeg},
}

// GSD: 0.5-10 meters.
gsd := 0.5 + rng().Float64()*9.5

// One detection at the entity position.
detConf := 0.5 + rng().Float64()*0.5
detAlt := entity.Position.AltM

detections := []*ingestionv1.ISRDetection{
{
Position: &commonv1.Position{
Latitude:       entity.Position.Lat,
Longitude:      entity.Position.Lon,
AltitudeMeters: &detAlt,
},
EntityType:  entity.EntityType,
Confidence:  detConf,
Description: fmt.Sprintf("Simulated detection of entity %s", entity.ID),
},
}

obs := &ingestionv1.SensorObservation{
ObservationId:   uuid.New().String(),
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_ISR,
ObservationTime: timestamppb.New(time.Now()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:  entity.Position.Lat,
Longitude: entity.Position.Lon,
},
Metadata: map[string]string{
"sim_entity_id": entity.ID,
},
SensorData: &ingestionv1.SensorObservation_Isr{
Isr: &ingestionv1.ISRObservation{
PlatformId:      platformID,
SensorName:      sensorName,
ImageId:         imageID,
CoveragePolygon: polygon,
GsdMeters:       &gsd,
Detections:      detections,
},
},
}
return obs
}
