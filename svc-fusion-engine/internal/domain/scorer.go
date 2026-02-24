// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"math"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
)

// CorrelationAction describes the outcome decision for a scored observation-track pair.
type CorrelationAction int

const (
	ActionAutoCorrelate CorrelationAction = iota // Score ≥ 0.85
	ActionTentative                               // Score 0.60–0.84
	ActionNewTrack                                // Score < 0.60
)

// CorrelationResult holds the weighted score and per-component breakdown.
type CorrelationResult struct {
	Score         float64
	PositionScore float64
	VelocityScore float64
	TypeScore     float64
	TemporalScore float64
	Action        CorrelationAction
}

// CorrelationScorer computes a weighted correlation score between an observation and a track.
type CorrelationScorer struct {
	weightPosition float64
	weightVelocity float64
	weightType     float64
	weightTemporal float64

	autoThreshold      float64
	tentativeThreshold float64
}

// NewCorrelationScorer creates a scorer with the specified weights and thresholds.
func NewCorrelationScorer(wPos, wVel, wType, wTemporal, autoThresh, tentThresh float64) *CorrelationScorer {
	return &CorrelationScorer{
		weightPosition:     wPos,
		weightVelocity:     wVel,
		weightType:         wType,
		weightTemporal:     wTemporal,
		autoThreshold:      autoThresh,
		tentativeThreshold: tentThresh,
	}
}

// Score calculates the weighted correlation score between an observation and a track.
func (s *CorrelationScorer) Score(
	obs *ingestionv1.SensorObservation,
	track *TrackState,
	gate GatingConfig,
) *CorrelationResult {
	obsTime := obs.GetObservationTime().AsTime()
	obsLat := obs.GetPosition().GetLatitude()
	obsLon := obs.GetPosition().GetLongitude()

	distNM := HaversineDistanceNM(obsLat, obsLon, track.KalmanState.Latitude, track.KalmanState.Longitude)

	timeDelta := obsTime.Sub(track.UpdatedAt)
	if timeDelta < 0 {
		timeDelta = -timeDelta
	}

	posScore := clamp01(1.0 - distNM/gate.MaxDistanceNM)
	velScore := s.scoreVelocity(obs, track)
	typeScore := scoreEntityType(sensorTypeToEntityType(obs.GetSensorType()), track.EntityType)
	tempScore := s.scoreTemporal(timeDelta, gate.MaxTimeDelta)

	total := s.weightPosition*posScore +
		s.weightVelocity*velScore +
		s.weightType*typeScore +
		s.weightTemporal*tempScore

	result := &CorrelationResult{
		Score:         total,
		PositionScore: posScore,
		VelocityScore: velScore,
		TypeScore:     typeScore,
		TemporalScore: tempScore,
	}
	switch {
	case total >= s.autoThreshold:
		result.Action = ActionAutoCorrelate
	case total >= s.tentativeThreshold:
		result.Action = ActionTentative
	default:
		result.Action = ActionNewTrack
	}
	return result
}

// scoreVelocity returns the velocity component score.
// Returns 0.5 (neutral) when either source lacks velocity.
func (s *CorrelationScorer) scoreVelocity(obs *ingestionv1.SensorObservation, track *TrackState) float64 {
	pos := obs.GetPosition()
	if pos == nil || pos.SpeedKnots == nil {
		return 0.5
	}
	if track.KalmanState == nil {
		return 0.5
	}
	obsSpeedMS := pos.GetSpeedKnots() * 0.514444 // knots → m/s

	trackSpeedMS := math.Sqrt(
		track.KalmanState.VelocityN*track.KalmanState.VelocityN +
			track.KalmanState.VelocityE*track.KalmanState.VelocityE,
	)

	const maxSpeedMS = 900.0 // ~1750 knots, covers aircraft
	diff := math.Abs(obsSpeedMS - trackSpeedMS)
	return clamp01(1.0 - diff/maxSpeedMS)
}

// scoreEntityType returns 1.0 for same type, 0.5 if either is unspecified, 0.0 for mismatch.
func scoreEntityType(obsType, trackType commonv1.EntityType) float64 {
	if obsType == commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED ||
		trackType == commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED {
		return 0.5
	}
	if obsType == trackType {
		return 1.0
	}
	return 0.0
}

// scoreTemporal returns 1 - (delta / maxDelta), clamped to [0,1].
func (s *CorrelationScorer) scoreTemporal(delta, maxDelta time.Duration) float64 {
	if maxDelta <= 0 {
		return 0.0
	}
	ratio := float64(delta) / float64(maxDelta)
	return clamp01(1.0 - ratio)
}

// String returns a Prometheus-friendly label for the action.
func (a CorrelationAction) String() string {
	switch a {
	case ActionAutoCorrelate:
		return "auto_correlate"
	case ActionTentative:
		return "tentative"
	default:
		return "new_track"
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
