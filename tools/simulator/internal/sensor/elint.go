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

// elintRadarTypes lists plausible radar type identifiers for synthetic data.
var elintRadarTypes = []string{
"SURFACE-SEARCH",
"AIR-SEARCH",
"FIRE-CONTROL",
"NAVIGATION",
"WEATHER",
"IFF-INTERROGATOR",
}

// elintScanTypes lists the scan types accepted by the ELINT validator.
var elintScanTypes = []string{"circular", "sector", "track-while-scan"}

// elintEmitterCounter tracks per-sensor emitter numbering.
var elintEmitterCounter = make(map[string]int)

// GenerateELINTObservation creates a SensorObservation with an ELINTDetection payload.
// All generated observations are CLASSIFICATION_LEVEL_UNCLASSIFIED.
func GenerateELINTObservation(entity *generator.SimEntity, sensorID string) *ingestionv1.SensorObservation {
elintEmitterCounter[sensorID]++
emitterID := fmt.Sprintf("ELINT-EMITTER-%s-%04d", sensorID, elintEmitterCounter[sensorID])

radarType := elintRadarTypes[rng().Intn(len(elintRadarTypes))]
scanType := elintScanTypes[rng().Intn(len(elintScanTypes))]

// Frequency: 0.5-40000 MHz valid range; use common radar bands.
freqMHz := 500.0 + rng().Float64()*17500.0
// Pulse width: 0.1-100 microseconds.
pulseWidthUS := 0.1 + rng().Float64()*99.9
// CEP: 100-10000 meters.
cepMeters := 100.0 + rng().Float64()*9900.0
// Confidence: 0.4-1.0.
confidence := 0.4 + rng().Float64()*0.6

altMeters := entity.Position.AltM
speedKn := entity.Position.SpeedKn
heading := entity.Position.Heading

obs := &ingestionv1.SensorObservation{
ObservationId:   uuid.New().String(),
SensorId:        sensorID,
SensorType:      commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT,
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
			"sim_entity_id":               entity.ID,
			"sim_entity_type":             entity.EntityType.String(),
			"sim_hostile_class":           entity.HostileClass.String(),
			"rtsa.coverage.range_nm":      fmt.Sprintf("%.1f", 300.0), // Very long range for ELINT
			"rtsa.coverage.bearing_start": fmt.Sprintf("%.1f", 200.0), // Directional (SW sector)
			"rtsa.coverage.bearing_end":   fmt.Sprintf("%.1f", 290.0),
			"rtsa.coverage.center_lat":    fmt.Sprintf("%.6f", 58.4),
			"rtsa.coverage.center_lon":    fmt.Sprintf("%.6f", -7.6),
		},
SensorData: &ingestionv1.SensorObservation_ElintComint{
ElintComint: &ingestionv1.ELINTDetection{
EmitterId:              emitterID,
RadarType:              radarType,
FrequencyMhz:           freqMHz,
PulseWidthUs:           &pulseWidthUS,
ScanType:               scanType,
CepMeters:              cepMeters,
Confidence:             confidence,
ContentClassification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
},
},
}
return obs
}
