// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// hostileClassFromMetadata extracts hostile classification from observation metadata.
// Returns UNKNOWN if the key is missing or unrecognised.
func hostileClassFromMetadata(md map[string]string) commonv1.HostileClassification {
	v, ok := md["sim_hostile_class"]
	if !ok || v == "" {
		return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN
	}
	// Accept both full enum name and short suffix (e.g. "HOSTILE" or "HOSTILE_CLASSIFICATION_HOSTILE")
	norm := strings.ToUpper(v)
	for num, name := range commonv1.HostileClassification_name {
		if name == norm || strings.HasSuffix(name, "_"+norm) {
			return commonv1.HostileClassification(num)
		}
	}
	return commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN
}

// entityTypeFromMetadata extracts entity type from observation metadata.
// Returns UNSPECIFIED if the key is missing or unrecognised.
func entityTypeFromMetadata(md map[string]string) commonv1.EntityType {
	v, ok := md["sim_entity_type"]
	if !ok || v == "" {
		return commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED
	}
	norm := strings.ToUpper(v)
	for num, name := range commonv1.EntityType_name {
		if name == norm || strings.HasSuffix(name, "_"+norm) {
			return commonv1.EntityType(num)
		}
	}
	return commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED
}

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
	md := obs.GetMetadata()

	// Prefer entity type from metadata (set by simulator), fall back to sensor type inference.
	et := entityTypeFromMetadata(md)
	if et == commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED {
		et = sensorTypeToEntityType(obs.GetSensorType())
	}

	// Extract hostile class from metadata if available.
	hc := hostileClassFromMetadata(md)

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
		HostileClass:   hc,
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

	// Propagate hostile class from metadata if more specific than current.
	md := obs.GetMetadata()
	obsHC := hostileClassFromMetadata(md)
	if track.HostileClass == commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN &&
		obsHC != commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN {
		track.HostileClass = obsHC
	}

	// Propagate entity type if track is still UNSPECIFIED.
	obsET := entityTypeFromMetadata(md)
	if track.EntityType == commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED &&
		obsET != commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED {
		track.EntityType = obsET
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
		speedMS := math.Sqrt(ts.KalmanState.VelocityN*ts.KalmanState.VelocityN +
			ts.KalmanState.VelocityE*ts.KalmanState.VelocityE)
		speedKnots := speedMS / 0.514444
		headingRad := math.Atan2(ts.KalmanState.VelocityE, ts.KalmanState.VelocityN)
		headingDeg := headingRad * 180.0 / math.Pi
		if headingDeg < 0 {
			headingDeg += 360.0
		}

		ft.EstimatedPosition = &commonv1.Position{
			Latitude:       ts.KalmanState.Latitude,
			Longitude:      ts.KalmanState.Longitude,
			SpeedKnots:     &speedKnots,
			HeadingDegrees: &headingDeg,
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
