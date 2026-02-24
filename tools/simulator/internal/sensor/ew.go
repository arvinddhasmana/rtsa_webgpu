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

// ewSensorPositions are fixed positions for simulated EW sensors.
var ewSensorPositions = map[string]generator.Position{
"EW-SIM-001": {Lat: 44.65, Lon: -63.57},
"EW-SIM-002": {Lat: 45.20, Lon: -59.50},
}

// ewModulationTypes lists valid EW modulation types.
var ewModulationTypes = []string{"PULSE", "CW", "FMCW", "AM", "FM"}

// emitterCounter tracks per-sensor emitter numbering.
var emitterCounter = make(map[string]int)

// GenerateEWObservation creates a SensorObservation with an EWIntercept payload.
// Only generated for entities that could plausibly have radar emissions.
// All generated observations are CLASSIFICATION_LEVEL_UNCLASSIFIED.
func GenerateEWObservation(entity *generator.SimEntity, sensorID string) *ingestionv1.SensorObservation {
emitterCounter[sensorID]++
emitterID := fmt.Sprintf("EW-EMITTER-%s-%04d", sensorID, emitterCounter[sensorID])

sensorPos := ewSensorPositions[sensorID]
if sensorPos.Lat == 0 && sensorPos.Lon == 0 {
sensorPos = generator.Position{Lat: 44.65, Lon: -63.57}
}

bearing := bearingFromPositions(sensorPos, entity.Position)

// Frequency: S-band to Ku-band (2000-18000 MHz). Within 0.5-40000 MHz range.
freqMHz := 2000.0 + rng().Float64()*16000.0
// Signal strength: -80 to -20 dBm.
powerDBM := -80.0 + rng().Float64()*60.0
// PRI: 10-5000 microseconds.
priMicrosec := 10.0 + rng().Float64()*4990.0
// Confidence: 0.5-1.0.
confidence := 0.5 + rng().Float64()*0.5
// Modulation.
modType := ewModulationTypes[rng().Intn(len(ewModulationTypes))]
// Bandwidth: 1-500 MHz.
bwMHz := 1.0 + rng().Float64()*499.0

altMeters := entity.Position.AltM
speedKn := entity.Position.SpeedKn
heading := entity.Position.Heading

obs := &ingestionv1.SensorObservation{
ObservationId:   uuid.New().String(),
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_EW_SIGINT,
ObservationTime: timestamppb.New(time.Now()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:       entity.Position.Lat,
Longitude:      entity.Position.Lon,
AltitudeMeters: &altMeters,
SpeedKnots:     &speedKn,
HeadingDegrees: &heading,
},
Metadata: map[string]string{
"sim_entity_id": entity.ID,
},
SensorData: &ingestionv1.SensorObservation_EwSigint{
EwSigint: &ingestionv1.EWIntercept{
EmitterId:       emitterID,
FrequencyMhz:    freqMHz,
BandwidthMhz:    &bwMHz,
PowerDbm:        &powerDBM,
BearingDegrees:  bearing,
PriMicroseconds: &priMicrosec,
ModulationType:  modType,
Confidence:      confidence,
},
},
}
return obs
}
