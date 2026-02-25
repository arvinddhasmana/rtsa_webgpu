// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"fmt"
	"math"
	"sync"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SourceInfo tracks per-sensor contribution metadata.
type SourceInfo struct {
	SensorID         string
	SensorType       commonv1.SensorType
	Confidence       float64
	LastContribution time.Time
	ObservationCount uint32
}

// TrackState holds the mutable in-memory state of a fused track.
type TrackState struct {
	TrackID        string
	EntityType     commonv1.EntityType
	HostileClass   commonv1.HostileClassification
	KalmanState    *KalmanState
	Sources        map[string]*SourceInfo // key: sensor_id
	Classification commonv1.ClassificationLevel
	Status         commonv1.TrackStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Label          string
}

// TrackManager manages the in-memory state of all active tracks. Thread-safe.
type TrackManager struct {
	mu     sync.RWMutex
	tracks map[string]*TrackState
	kf     *KalmanFilter
}

// NewTrackManager creates a TrackManager backed by the provided KalmanFilter.
func NewTrackManager(kf *KalmanFilter) *TrackManager {
	return &TrackManager{
		tracks: make(map[string]*TrackState),
		kf:     kf,
	}
}

// CreateTrack creates a new fused track from an initial observation.
// Returns an error if the observation lacks a position.
func (tm *TrackManager) CreateTrack(obs *ingestionv1.SensorObservation) (*TrackState, error) {
	if obs.GetPosition() == nil {
		return nil, fmt.Errorf("track_manager: CreateTrack: observation has no position")
	}

	now := time.Now()
	pos := obs.GetPosition()
	et := sensorTypeToEntityType(obs.GetSensorType())

	var vN, vE float64
	if pos.SpeedKnots != nil && pos.HeadingDegrees != nil {
		vN, vE = speedHeadingToNE(pos.GetSpeedKnots(), pos.GetHeadingDegrees())
	}

	var vNPtr, vEPtr *float64
	if pos.SpeedKnots != nil {
		vN2 := vN
		vE2 := vE
		vNPtr = &vN2
		vEPtr = &vE2
	}

	obsTime := obs.GetObservationTime().AsTime()
	m := &Measurement{
		Latitude:  pos.GetLatitude(),
		Longitude: pos.GetLongitude(),
		VelocityN: vNPtr,
		VelocityE: vEPtr,
		Time:      obsTime,
	}

	ks := tm.kf.InitState(m)

	track := &TrackState{
		TrackID:        uuid.New().String(),
		EntityType:     et,
		HostileClass:   commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN,
		KalmanState:    ks,
		Sources:        make(map[string]*SourceInfo),
		Classification: obs.GetClassification(),
		Status:         commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	sensorID := obs.GetSensorId()
	track.Sources[sensorID] = &SourceInfo{
		SensorID:         sensorID,
		SensorType:       obs.GetSensorType(),
		Confidence:       0.7,
		LastContribution: now,
		ObservationCount: 1,
	}

	tm.mu.Lock()
	tm.tracks[track.TrackID] = track
	tm.mu.Unlock()

	return track, nil
}

// UpdateTrack incorporates a new observation into an existing track.
func (tm *TrackManager) UpdateTrack(trackID string, obs *ingestionv1.SensorObservation) (*TrackState, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	track, ok := tm.tracks[trackID]
	if !ok {
		return nil, fmt.Errorf("track_manager: UpdateTrack: track %q not found", trackID)
	}

	now := time.Now()
	pos := obs.GetPosition()
	if pos == nil {
		return track, nil
	}

	obsTime := obs.GetObservationTime().AsTime()
	dt := obsTime.Sub(track.KalmanState.LastUpdate).Seconds()
	if dt > 0 {
		tm.kf.Predict(track.KalmanState, dt)
	}

	var vN, vE float64
	if pos.SpeedKnots != nil && pos.HeadingDegrees != nil {
		vN, vE = speedHeadingToNE(pos.GetSpeedKnots(), pos.GetHeadingDegrees())
	}

	var vNPtr, vEPtr *float64
	if pos.SpeedKnots != nil {
		vNPtr = &vN
		vEPtr = &vE
	}

	m := &Measurement{
		Latitude:  pos.GetLatitude(),
		Longitude: pos.GetLongitude(),
		VelocityN: vNPtr,
		VelocityE: vEPtr,
		Time:      obsTime,
	}
	tm.kf.Update(track.KalmanState, m)

	// Update source attribution
	sensorID := obs.GetSensorId()
	if src, exists := track.Sources[sensorID]; exists {
		src.ObservationCount++
		src.LastContribution = now
	} else {
		track.Sources[sensorID] = &SourceInfo{
			SensorID:         sensorID,
			SensorType:       obs.GetSensorType(),
			Confidence:       0.7,
			LastContribution: now,
			ObservationCount: 1,
		}
	}

	// Propagate classification: MAX of all contributing sources
	if obs.GetClassification() > track.Classification {
		track.Classification = obs.GetClassification()
	}

	track.Status = commonv1.TrackStatus_TRACK_STATUS_ACTIVE
	track.UpdatedAt = now

	return track, nil
}

// MergeTracks merges trackB into trackA. trackB is marked MERGED.
func (tm *TrackManager) MergeTracks(trackAID, trackBID string) (*TrackState, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	trackA, ok := tm.tracks[trackAID]
	if !ok {
		return nil, fmt.Errorf("track_manager: MergeTracks: track A %q not found", trackAID)
	}
	trackB, ok := tm.tracks[trackBID]
	if !ok {
		return nil, fmt.Errorf("track_manager: MergeTracks: track B %q not found", trackBID)
	}

	// Transfer sources from B to A
	for k, v := range trackB.Sources {
		if existing, ok := trackA.Sources[k]; ok {
			existing.ObservationCount += v.ObservationCount
			if v.LastContribution.After(existing.LastContribution) {
				existing.LastContribution = v.LastContribution
			}
		} else {
			trackA.Sources[k] = v
		}
	}

	// Propagate classification
	if trackB.Classification > trackA.Classification {
		trackA.Classification = trackB.Classification
	}

	trackB.Status = commonv1.TrackStatus_TRACK_STATUS_MERGED
	trackA.UpdatedAt = time.Now()

	return trackA, nil
}

// GetActiveTracks returns all tracks with status ACTIVE or STALE.
func (tm *TrackManager) GetActiveTracks() []*TrackState {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var result []*TrackState
	for _, t := range tm.tracks {
		if t.Status == commonv1.TrackStatus_TRACK_STATUS_ACTIVE ||
			t.Status == commonv1.TrackStatus_TRACK_STATUS_STALE {
			result = append(result, t)
		}
	}
	return result
}

// GetTrack returns a specific track by ID and whether it was found.
func (tm *TrackManager) GetTrack(trackID string) (*TrackState, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.tracks[trackID]
	return t, ok
}

// MarkStale marks a track as STALE.
func (tm *TrackManager) MarkStale(trackID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if t, ok := tm.tracks[trackID]; ok {
		t.Status = commonv1.TrackStatus_TRACK_STATUS_STALE
	}
}

// MarkDropped marks a track as DROPPED.
func (tm *TrackManager) MarkDropped(trackID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if t, ok := tm.tracks[trackID]; ok {
		t.Status = commonv1.TrackStatus_TRACK_STATUS_DROPPED
	}
}

// TrackCount returns the total number of tracks (all statuses).
func (tm *TrackManager) TrackCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tracks)
}

// ToFusedTrack converts the TrackState to a FusedTrack protobuf message.
func (ts *TrackState) ToFusedTrack() *entityv1.FusedTrack {
	now := time.Now()
	ageSeconds := now.Sub(ts.CreatedAt).Seconds()

	ft := &entityv1.FusedTrack{
		TrackId:         ts.TrackID,
		EntityType:      ts.EntityType,
		HostileClass:    ts.HostileClass,
		ConfidenceScore: computeConfidence(ts),
		SourceCount:     uint32(len(ts.Sources)),
		Status:          ts.Status,
		Classification:  ts.Classification,
		CreatedAt:       timestamppb.New(ts.CreatedAt),
		UpdatedAt:       timestamppb.New(ts.UpdatedAt),
		AgeSeconds:      ageSeconds,
	}

	if ts.KalmanState != nil {
		ft.EstimatedPosition = &commonv1.Position{
			Latitude:  ts.KalmanState.Latitude,
			Longitude: ts.KalmanState.Longitude,
		}
		ft.Velocity = &commonv1.Velocity{
			NorthMps: ts.KalmanState.VelocityN,
			EastMps:  ts.KalmanState.VelocityE,
		}
	}

	for _, src := range ts.Sources {
		ft.Sources = append(ft.Sources, &entityv1.SourceAttribution{
			SensorId:         src.SensorID,
			SensorType:       src.SensorType,
			Confidence:       src.Confidence,
			LastContribution: timestamppb.New(src.LastContribution),
			ObservationCount: src.ObservationCount,
		})
	}

	return ft
}

// speedHeadingToNE converts speed (knots) and heading (degrees) to N/E velocity components (m/s).
func speedHeadingToNE(speedKnots, headingDeg float64) (vN, vE float64) {
	speedMS := speedKnots * 0.514444
	rad := degreesToRadians(headingDeg)
	vN = speedMS * math.Cos(rad)
	vE = speedMS * math.Sin(rad)
	return
}

// computeConfidence estimates track confidence from source count and Kalman uncertainty.
func computeConfidence(ts *TrackState) float64 {
	base := 0.5
	if len(ts.Sources) >= 3 {
		base = 0.9
	} else if len(ts.Sources) == 2 {
		base = 0.75
	}
	return base
}
