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

// vesselNames provides synthetic vessel name pool.
var vesselNames = []string{
"ATLANTIC SPIRIT", "NOVA STAR", "CABOT TRAIL", "SABLE ISLAND",
"FUNDY TIDE", "MARITIMER", "CAPE BRETON", "BLUENOSE TWO",
"ROYAL VENTURE", "OCEAN EXPLORER", "GRAND BANKS", "ISLE ROYALE",
"ST LAWRENCE", "BAY SPIRIT", "ACADIAN FISHER",
}

// vesselTypeCodes maps to AIS vessel type codes (must be 1-99).
// 30 = fishing, 70 = cargo, 80 = tanker, 60 = passenger, 90 = other
var vesselTypeCodes = []int32{30, 60, 70, 80, 90}

// aisMessageTypes lists valid AIS message types per the validator.
var aisMessageTypes = []int32{1, 2, 3, 18}

// navStatuses provides AIS navigation status strings.
var navStatuses = []string{
"under-way-using-engine",
"at-anchor",
"not-under-command",
"restricted-manoeuvrability",
"moored",
}

// entityMMSIs stores per-entity MMSI values for consistency across ticks.
// MMSI format: 9 digits, starting with 3-digit MID (Maritime ID).
// We use 316XXXXXX (Canada).
var entityMMSIs = make(map[string]string)

// GenerateAISObservation creates a SensorObservation with an AISPosition payload.
//
// If manipulatedPos is non-nil, that position is reported instead of the entity's
// true position (simulates AIS spoofing/manipulation anomaly).
//
// All generated observations are CLASSIFICATION_LEVEL_UNCLASSIFIED.
// is_bft is always false (BFT requires PROTECTED_B classification).
func GenerateAISObservation(entity *generator.SimEntity, manipulatedPos *generator.Position) *ingestionv1.SensorObservation {
mmsi := getOrCreateMMSI(entity.ID)
vesselName := vesselNames[stableIdx(entity.ID, len(vesselNames))]
vesselType := vesselTypeCodes[stableIdx(entity.ID, len(vesselTypeCodes))]
msgType := aisMessageTypes[rng().Intn(len(aisMessageTypes))]
navStatus := navStatuses[rng().Intn(len(navStatuses))]

// Use manipulated position if provided (for AIS spoofing), else true position.
pos := entity.Position
if manipulatedPos != nil {
pos = *manipulatedPos
}

speedKn := pos.SpeedKn
if speedKn < 0 {
speedKn = 0
}
heading := pos.Heading
rot := (rng().Float64()*20 - 10) // -10 to +10 deg/min
draught := 2.0 + rng().Float64()*10.0 // 2-12 m

callSign := fmt.Sprintf("VE%d%s", rng().Intn(9)+1, entity.ID[:3])
dest := "HALIFAX"

obs := &ingestionv1.SensorObservation{
ObservationId:   uuid.New().String(),
SensorId:        "AIS-SIM-001",
SensorType:      commonv1.SensorType_SENSOR_TYPE_AIS_BFT,
ObservationTime: timestamppb.New(time.Now()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Position: &commonv1.Position{
Latitude:       pos.Lat,
Longitude:      pos.Lon,
SpeedKnots:     &speedKn,
HeadingDegrees: &heading,
},
Metadata: map[string]string{
"sim_entity_id":     entity.ID,
"sim_entity_type":   entity.EntityType.String(),
"sim_hostile_class": entity.HostileClass.String(),
},
SensorData: &ingestionv1.SensorObservation_AisBft{
AisBft: &ingestionv1.AISPosition{
Mmsi:           mmsi,
VesselName:     vesselName,
VesselTypeCode: vesselType,
NavStatus:      navStatus,
RateOfTurn:     &rot,
DraughtMeters:  &draught,
Destination:    &dest,
IsBft:          false, // UNCLASSIFIED only; BFT requires PROTECTED_B
AisMessageType: msgType,
CallSign:       &callSign,
},
},
}
return obs
}

// getOrCreateMMSI returns a stable 9-digit MMSI for the given entity ID.
// Uses Canadian MID prefix 316.
func getOrCreateMMSI(entityID string) string {
if mmsi, ok := entityMMSIs[entityID]; ok {
return mmsi
}
// Generate a 6-digit suffix to append to "316".
suffix := 100000 + rng().Intn(899999) // 100000-999999 → 9 digits total
mmsi := fmt.Sprintf("316%06d", suffix)
entityMMSIs[entityID] = mmsi
return mmsi
}

// stableIdx derives a consistent index from an entity ID string.
func stableIdx(entityID string, n int) int {
h := 0
for _, c := range entityID {
h = h*31 + int(c)
}
if h < 0 {
h = -h
}
return h % n
}
