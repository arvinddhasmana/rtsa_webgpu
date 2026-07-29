// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"math"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FindMergeCandidates returns pairs of track IDs that should be merged.
// A pair qualifies when both tracks score ≥ threshold against each other
// (bidirectional correlation check).
func FindMergeCandidates(
	tracks []*TrackState,
	scorer *CorrelationScorer,
	gate GatingConfig,
	threshold float64,
) [][2]string {
	var pairs [][2]string
	for i := 0; i < len(tracks); i++ {
		for j := i + 1; j < len(tracks); j++ {
			a := tracks[i]
			b := tracks[j]
			if a.KalmanState == nil || b.KalmanState == nil {
				continue
			}
			// Score A against B
			obsFromA := trackToSyntheticObs(a)
			scoreAB := scorer.Score(obsFromA, b, gate)
			// Score B against A
			obsFromB := trackToSyntheticObs(b)
			scoreBA := scorer.Score(obsFromB, a, gate)

			if scoreAB.Score >= threshold && scoreBA.Score >= threshold {
				pairs = append(pairs, [2]string{a.TrackID, b.TrackID})
			}
		}
	}
	return pairs
}

// trackToSyntheticObs converts a TrackState to a synthetic SensorObservation
// suitable for scoring purposes.
func trackToSyntheticObs(ts *TrackState) *ingestionv1.SensorObservation {
	pos := &commonv1.Position{
		Latitude:  ts.KalmanState.Latitude,
		Longitude: ts.KalmanState.Longitude,
	}
	speedKnots := 0.0
	if ts.KalmanState.VelocityN != 0 || ts.KalmanState.VelocityE != 0 {
		speedMS := magnitude(ts.KalmanState.VelocityN, ts.KalmanState.VelocityE)
		speedKnots = speedMS / 0.514444
	}
	pos.SpeedKnots = &speedKnots

	return &ingestionv1.SensorObservation{
		SensorId:        ts.TrackID,
		SensorType:      sensorTypeForEntityType(ts.EntityType),
		ObservationTime: timestamppb.New(ts.UpdatedAt),
		Classification:  ts.Classification,
		Position:        pos,
	}
}

// sensorTypeForEntityType picks a representative sensor type for an entity type.
func sensorTypeForEntityType(et commonv1.EntityType) commonv1.SensorType {
	switch et {
	case commonv1.EntityType_ENTITY_TYPE_AIR:
		return commonv1.SensorType_SENSOR_TYPE_EW_SIGINT
	case commonv1.EntityType_ENTITY_TYPE_SUBSURFACE:
		return commonv1.SensorType_SENSOR_TYPE_RADAR
	case commonv1.EntityType_ENTITY_TYPE_CYBER:
		return commonv1.SensorType_SENSOR_TYPE_CYBER
	default:
		return commonv1.SensorType_SENSOR_TYPE_RADAR
	}
}

func magnitude(a, b float64) float64 {
	return math.Sqrt(a*a + b*b)
}
