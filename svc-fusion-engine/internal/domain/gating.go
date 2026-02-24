// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"math"
	"sort"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
)

const (
	earthRadiusM = 6_371_000.0
	metersPerNM  = 1_852.0
)

// GatingConfig holds spatial and temporal gate parameters for one entity type.
type GatingConfig struct {
	MaxDistanceNM float64
	MaxTimeDelta  time.Duration
}

// GatingFilter determines candidate tracks for a new sensor observation.
type GatingFilter struct {
	configs map[commonv1.EntityType]GatingConfig
}

// NewGatingFilter creates a GatingFilter with per-entity-type thresholds.
func NewGatingFilter(surface, air, sub GatingConfig) *GatingFilter {
	return &GatingFilter{
		configs: map[commonv1.EntityType]GatingConfig{
			commonv1.EntityType_ENTITY_TYPE_SURFACE:    surface,
			commonv1.EntityType_ENTITY_TYPE_AIR:        air,
			commonv1.EntityType_ENTITY_TYPE_SUBSURFACE: sub,
			commonv1.EntityType_ENTITY_TYPE_LAND:       surface,
			commonv1.EntityType_ENTITY_TYPE_CYBER:      sub,
		},
	}
}

// GatingConfigFor returns the gating config for the given entity type.
func (g *GatingFilter) GatingConfigFor(et commonv1.EntityType) GatingConfig {
	if cfg, ok := g.configs[et]; ok {
		return cfg
	}
	return g.configs[commonv1.EntityType_ENTITY_TYPE_SURFACE]
}

// FindCandidates returns all compatible active tracks within the spatial/temporal gate,
// sorted nearest-first.
func (g *GatingFilter) FindCandidates(obs *ingestionv1.SensorObservation, tracks []*TrackState) []*TrackState {
	if obs.GetPosition() == nil {
		return nil
	}

	obsET := sensorTypeToEntityType(obs.GetSensorType())
	cfg := g.GatingConfigFor(obsET)
	obsTime := obs.GetObservationTime().AsTime()
	obsLat := obs.GetPosition().GetLatitude()
	obsLon := obs.GetPosition().GetLongitude()

	type candidate struct {
		track *TrackState
		dist  float64
	}

	var result []candidate
	for _, t := range tracks {
		if t.Status == commonv1.TrackStatus_TRACK_STATUS_DROPPED ||
			t.Status == commonv1.TrackStatus_TRACK_STATUS_MERGED {
			continue
		}
		if !entityTypesCompatible(obsET, t.EntityType) {
			continue
		}
		if t.KalmanState == nil {
			continue
		}
		dist := HaversineDistanceNM(obsLat, obsLon, t.KalmanState.Latitude, t.KalmanState.Longitude)
		if dist > cfg.MaxDistanceNM {
			continue
		}
		delta := obsTime.Sub(t.UpdatedAt)
		if delta < 0 {
			delta = -delta
		}
		if delta > cfg.MaxTimeDelta {
			continue
		}
		result = append(result, candidate{track: t, dist: dist})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].dist < result[j].dist
	})

	out := make([]*TrackState, len(result))
	for i, c := range result {
		out[i] = c.track
	}
	return out
}

// HaversineDistanceNM calculates the great-circle distance in nautical miles.
func HaversineDistanceNM(lat1, lon1, lat2, lon2 float64) float64 {
	φ1 := degreesToRadians(lat1)
	φ2 := degreesToRadians(lat2)
	Δφ := degreesToRadians(lat2 - lat1)
	Δλ := degreesToRadians(lon2 - lon1)

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distM := earthRadiusM * c
	return distM / metersPerNM
}

// sensorTypeToEntityType maps sensor type to the primary entity type it detects.
func sensorTypeToEntityType(st commonv1.SensorType) commonv1.EntityType {
	switch st {
	case commonv1.SensorType_SENSOR_TYPE_RADAR:
		return commonv1.EntityType_ENTITY_TYPE_SURFACE
	case commonv1.SensorType_SENSOR_TYPE_EW_SIGINT:
		return commonv1.EntityType_ENTITY_TYPE_AIR
	case commonv1.SensorType_SENSOR_TYPE_ELINT_COMINT:
		return commonv1.EntityType_ENTITY_TYPE_AIR
	case commonv1.SensorType_SENSOR_TYPE_ISR:
		return commonv1.EntityType_ENTITY_TYPE_SURFACE
	case commonv1.SensorType_SENSOR_TYPE_AIS_BFT:
		return commonv1.EntityType_ENTITY_TYPE_SURFACE
	case commonv1.SensorType_SENSOR_TYPE_CYBER:
		return commonv1.EntityType_ENTITY_TYPE_CYBER
	default:
		return commonv1.EntityType_ENTITY_TYPE_SURFACE
	}
}

// entityTypesCompatible returns true when an observation of obsType can correlate with trackType.
func entityTypesCompatible(obsType, trackType commonv1.EntityType) bool {
	if obsType == commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED ||
		trackType == commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED {
		return true
	}
	return obsType == trackType
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}
