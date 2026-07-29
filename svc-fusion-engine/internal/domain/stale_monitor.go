// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"context"
	"time"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
)

// StaleMonitor runs a background goroutine that periodically scans active tracks
// for staleness and transitions their status accordingly.
type StaleMonitor struct {
	manager        *TrackManager
	staleTimeout   time.Duration
	dropTimeout    time.Duration
	checkInterval  time.Duration
	onStatusChange func(track *TrackState, oldStatus, newStatus commonv1.TrackStatus)
}

// NewStaleMonitor creates a StaleMonitor with the provided timeouts and callback.
func NewStaleMonitor(
	manager *TrackManager,
	staleTimeout, dropTimeout, checkInterval time.Duration,
	onStatusChange func(*TrackState, commonv1.TrackStatus, commonv1.TrackStatus),
) *StaleMonitor {
	return &StaleMonitor{
		manager:        manager,
		staleTimeout:   staleTimeout,
		dropTimeout:    dropTimeout,
		checkInterval:  checkInterval,
		onStatusChange: onStatusChange,
	}
}

// Start begins the stale monitoring loop. Blocks until ctx is cancelled.
func (sm *StaleMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(sm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sm.runCheck()
		}
	}
}

// runCheck inspects all tracks and applies stale/drop transitions.
func (sm *StaleMonitor) runCheck() {
	now := time.Now()
	tracks := sm.manager.GetActiveTracks()

	for _, track := range tracks {
		age := now.Sub(track.UpdatedAt)
		oldStatus := track.Status

		switch oldStatus {
		case commonv1.TrackStatus_TRACK_STATUS_ACTIVE:
			if age > sm.staleTimeout {
				sm.manager.MarkStale(track.TrackID)
				sm.notifyChange(track, oldStatus, commonv1.TrackStatus_TRACK_STATUS_STALE)
			}
		case commonv1.TrackStatus_TRACK_STATUS_STALE:
			if age > sm.dropTimeout {
				sm.manager.MarkDropped(track.TrackID)
				sm.notifyChange(track, oldStatus, commonv1.TrackStatus_TRACK_STATUS_DROPPED)
			}
		}
	}
}

func (sm *StaleMonitor) notifyChange(track *TrackState, oldStatus, newStatus commonv1.TrackStatus) {
	if sm.onStatusChange != nil {
		sm.onStatusChange(track, oldStatus, newStatus)
	}
}
