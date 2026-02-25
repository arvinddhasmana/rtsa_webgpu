// CLASSIFICATION: UNCLASSIFIED
// Package domain provides the in-memory track cache for svc-track.
//
// The TrackCache maintains current state of all tracks indexed by track_id for
// O(1) lookup.  It also maintains a bounded position history (up to maxHistory
// points) per track, and invokes an onChange callback on every state transition
// to enable fan-out streaming to connected clients.
//
// Feature: FEAT-13 Situational Awareness UI
// UC: UC012 Situational Awareness UI
// Requirements: CR-UI-001, CR-UI-002, CR-SEC-001
package domain

import (
	"fmt"
	"sync"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// DefaultHistoryMaxPoints is the default maximum position history points per track.
	DefaultHistoryMaxPoints = 100
)

// HistoryPoint is a historical position snapshot for a track.
type HistoryPoint struct {
	Position   *commonv1.Position
	Timestamp  time.Time
	Confidence float64
	Status     commonv1.TrackStatus
}

// CachedTrack wraps a FusedTrack with cache metadata.
type CachedTrack struct {
	Track    *entityv1.FusedTrack
	CachedAt time.Time
}

// TrackCache maintains current state of all tracks in memory.
// It is safe for concurrent use by multiple goroutines.
type TrackCache struct {
	mu            sync.RWMutex
	tracks        map[string]*CachedTrack
	history       map[string][]*HistoryPoint
	onChange      func(update *entityv1.TrackUpdate)
	historyMaxPts int
}

// NewTrackCache creates a new TrackCache.
// historyMax controls maximum history points retained per track (default 100 if ≤0).
func NewTrackCache(historyMax int) *TrackCache {
	if historyMax <= 0 {
		historyMax = DefaultHistoryMaxPoints
	}
	return &TrackCache{
		tracks:        make(map[string]*CachedTrack),
		history:       make(map[string][]*HistoryPoint),
		historyMaxPts: historyMax,
	}
}

// SetOnChange registers the callback invoked after every state transition.
// Only one callback is supported; subsequent calls replace the previous.
// The callback is called with no locks held.
func (c *TrackCache) SetOnChange(fn func(update *entityv1.TrackUpdate)) {
	c.mu.Lock()
	c.onChange = fn
	c.mu.Unlock()
}

// Put inserts or updates a track in the cache and triggers onChange.
// Update type semantics:
//   - Track did not exist             → UPDATE_TYPE_CREATED
//   - Track exists, status == DROPPED → UPDATE_TYPE_DROPPED
//   - Track exists, status == MERGED  → UPDATE_TYPE_MERGED
//   - Track exists, otherwise         → UPDATE_TYPE_UPDATED
func (c *TrackCache) Put(track *entityv1.FusedTrack) {
	if track == nil || track.TrackId == "" {
		return
	}

	c.mu.Lock()
	update, onChange := c.putLocked(track)
	c.mu.Unlock()

	if onChange != nil {
		onChange(update)
	}
}

// putLocked performs the actual cache mutation. Must be called with c.mu held (write lock).
// Returns the TrackUpdate to broadcast and the current onChange callback.
func (c *TrackCache) putLocked(track *entityv1.FusedTrack) (*entityv1.TrackUpdate, func(*entityv1.TrackUpdate)) {
	id := track.TrackId
	_, exists := c.tracks[id]

	var updateType entityv1.TrackUpdate_UpdateType
	switch {
	case !exists:
		updateType = entityv1.TrackUpdate_UPDATE_TYPE_CREATED
	case track.Status == commonv1.TrackStatus_TRACK_STATUS_DROPPED:
		updateType = entityv1.TrackUpdate_UPDATE_TYPE_DROPPED
	case track.Status == commonv1.TrackStatus_TRACK_STATUS_MERGED:
		updateType = entityv1.TrackUpdate_UPDATE_TYPE_MERGED
	default:
		updateType = entityv1.TrackUpdate_UPDATE_TYPE_UPDATED
	}

	// Update the track map.
	c.tracks[id] = &CachedTrack{
		Track:    track,
		CachedAt: time.Now(),
	}

	// Append position to history if present.
	if pos := track.EstimatedPosition; pos != nil {
		ts := time.Now()
		if track.UpdatedAt != nil {
			ts = track.UpdatedAt.AsTime()
		}
		hp := &HistoryPoint{
			Position:   pos,
			Timestamp:  ts,
			Confidence: track.ConfidenceScore,
			Status:     track.Status,
		}
		hist := c.history[id]
		hist = append(hist, hp)
		if len(hist) > c.historyMaxPts {
			// Trim oldest entries; keep newest historyMaxPts.
			hist = hist[len(hist)-c.historyMaxPts:]
		}
		c.history[id] = hist
	}

	update := &entityv1.TrackUpdate{
		UpdateType: updateType,
		Track:      track,
	}

	return update, c.onChange
}

// Get returns a track by ID. Returns nil if not found.
func (c *TrackCache) Get(trackID string) *entityv1.FusedTrack {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if ct, ok := c.tracks[trackID]; ok {
		return ct.Track
	}
	return nil
}

// GetAll returns all tracks whose status is ACTIVE, STALE, or NEW.
// DROPPED and MERGED tracks are excluded.
func (c *TrackCache) GetAll() []*entityv1.FusedTrack {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*entityv1.FusedTrack, 0, len(c.tracks))
	for _, ct := range c.tracks {
		if isActiveStatus(ct.Track.Status) {
			result = append(result, ct.Track)
		}
	}
	return result
}

// GetFiltered returns active tracks matching all criteria in the filter.
func (c *TrackCache) GetFiltered(filter *TrackFilter) []*entityv1.FusedTrack {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fe := &FilterEngine{}
	active := make([]*entityv1.FusedTrack, 0, len(c.tracks))
	for _, ct := range c.tracks {
		if isActiveStatus(ct.Track.Status) {
			active = append(active, ct.Track)
		}
	}
	return fe.Apply(active, filter)
}

// GetHistory returns up to maxPoints history entries for a track, oldest first.
// If maxPoints ≤ 0, DefaultHistoryMaxPoints is used.
func (c *TrackCache) GetHistory(trackID string, maxPoints int) []*HistoryPoint {
	if maxPoints <= 0 {
		maxPoints = DefaultHistoryMaxPoints
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	hist := c.history[trackID]
	if len(hist) == 0 {
		return nil
	}
	if len(hist) > maxPoints {
		hist = hist[len(hist)-maxPoints:]
	}
	// Return a copy to prevent external mutation of internal state.
	out := make([]*HistoryPoint, len(hist))
	copy(out, hist)
	return out
}

// Count returns the number of active (non-DROPPED, non-MERGED) tracks.
func (c *TrackCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	n := 0
	for _, ct := range c.tracks {
		if isActiveStatus(ct.Track.Status) {
			n++
		}
	}
	return n
}

// Snapshot sends all currently active filtered tracks as SNAPSHOT updates.
// This is used during StreamTracks initialisation to send the initial state.
func (c *TrackCache) Snapshot(filter *TrackFilter, send func(*entityv1.TrackUpdate) error) error {
	tracks := c.GetFiltered(filter)
	for _, t := range tracks {
		update := &entityv1.TrackUpdate{
			UpdateType: entityv1.TrackUpdate_UPDATE_TYPE_SNAPSHOT,
			Track:      t,
		}
		if err := send(update); err != nil {
			return fmt.Errorf("domain.TrackCache.Snapshot(trackID=%s): %w", t.TrackId, err)
		}
	}
	return nil
}

// ToProtoHistoryPoint converts a domain HistoryPoint to its protobuf representation.
func ToProtoHistoryPoint(hp *HistoryPoint) *entityv1.TrackHistoryPoint {
	return &entityv1.TrackHistoryPoint{
		Position:   hp.Position,
		Timestamp:  timestamppb.New(hp.Timestamp),
		Confidence: hp.Confidence,
		Status:     hp.Status,
	}
}

// isActiveStatus reports whether a TrackStatus should be included in active track queries.
func isActiveStatus(s commonv1.TrackStatus) bool {
	return s == commonv1.TrackStatus_TRACK_STATUS_ACTIVE ||
		s == commonv1.TrackStatus_TRACK_STATUS_STALE ||
		s == commonv1.TrackStatus_TRACK_STATUS_NEW
}
