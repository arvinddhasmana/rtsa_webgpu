// CLASSIFICATION: UNCLASSIFIED
// Package domain — unit tests for TrackCache.
//
// Test coverage for T01–T04 per Module 10 specification.
package domain

import (
	"sync"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	entityv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/entity/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// makeTrack is a helper that builds a minimal FusedTrack for testing.
func makeTrack(id string, status commonv1.TrackStatus, et commonv1.EntityType, hc commonv1.HostileClassification, cls commonv1.ClassificationLevel, confidence float64) *entityv1.FusedTrack {
	return &entityv1.FusedTrack{
		TrackId:      id,
		Status:       status,
		EntityType:   et,
		HostileClass: hc,
		Classification: cls,
		ConfidenceScore: confidence,
		EstimatedPosition: &commonv1.Position{
			Latitude:  45.4215,
			Longitude: -75.6972,
		},
		UpdatedAt: timestamppb.New(time.Now()),
	}
}

// T01: Put new track → onChange called with UPDATE_TYPE_CREATED.
func TestTrackCache_Put_NewTrack_CREATED(t *testing.T) {
	cache := NewTrackCache(10)
	var (
		mu         sync.Mutex
		gotUpdates []*entityv1.TrackUpdate
	)
	cache.SetOnChange(func(u *entityv1.TrackUpdate) {
		mu.Lock()
		gotUpdates = append(gotUpdates, u)
		mu.Unlock()
	})

	track := makeTrack("trk-001", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9)
	cache.Put(track)

	mu.Lock()
	defer mu.Unlock()
	if len(gotUpdates) != 1 {
		t.Fatalf("expected 1 onChange call, got %d", len(gotUpdates))
	}
	if gotUpdates[0].UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_CREATED {
		t.Errorf("expected UPDATE_TYPE_CREATED, got %v", gotUpdates[0].UpdateType)
	}
	if gotUpdates[0].Track.TrackId != "trk-001" {
		t.Errorf("expected track_id=trk-001, got %q", gotUpdates[0].Track.TrackId)
	}
}

// T02: Put existing track → onChange called with UPDATE_TYPE_UPDATED.
func TestTrackCache_Put_ExistingTrack_UPDATED(t *testing.T) {
	cache := NewTrackCache(10)
	track := makeTrack("trk-002", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.7)
	cache.Put(track) // first put — creates

	var (
		mu         sync.Mutex
		gotUpdates []*entityv1.TrackUpdate
	)
	cache.SetOnChange(func(u *entityv1.TrackUpdate) {
		mu.Lock()
		gotUpdates = append(gotUpdates, u)
		mu.Unlock()
	})

	// Second put — same ID, still active.
	track2 := makeTrack("trk-002", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.85)
	cache.Put(track2)

	mu.Lock()
	defer mu.Unlock()
	if len(gotUpdates) != 1 {
		t.Fatalf("expected 1 onChange call, got %d", len(gotUpdates))
	}
	if gotUpdates[0].UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_UPDATED {
		t.Errorf("expected UPDATE_TYPE_UPDATED, got %v", gotUpdates[0].UpdateType)
	}
}

// T03: Put track with DROPPED status → onChange called with UPDATE_TYPE_DROPPED.
func TestTrackCache_Put_DroppedStatus_DROPPED(t *testing.T) {
	cache := NewTrackCache(10)
	// Insert track first.
	track := makeTrack("trk-003", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.6)
	cache.Put(track)

	var (
		mu         sync.Mutex
		gotUpdates []*entityv1.TrackUpdate
	)
	cache.SetOnChange(func(u *entityv1.TrackUpdate) {
		mu.Lock()
		gotUpdates = append(gotUpdates, u)
		mu.Unlock()
	})

	// Now drop it.
	dropped := makeTrack("trk-003", commonv1.TrackStatus_TRACK_STATUS_DROPPED, commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.6)
	cache.Put(dropped)

	mu.Lock()
	defer mu.Unlock()
	if len(gotUpdates) != 1 {
		t.Fatalf("expected 1 onChange call, got %d", len(gotUpdates))
	}
	if gotUpdates[0].UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_DROPPED {
		t.Errorf("expected UPDATE_TYPE_DROPPED, got %v", gotUpdates[0].UpdateType)
	}
}

// T03b: Put track with MERGED status → onChange called with UPDATE_TYPE_MERGED.
func TestTrackCache_Put_MergedStatus_MERGED(t *testing.T) {
	cache := NewTrackCache(10)
	track := makeTrack("trk-004", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.5)
	cache.Put(track)

	var (
		mu         sync.Mutex
		gotUpdates []*entityv1.TrackUpdate
	)
	cache.SetOnChange(func(u *entityv1.TrackUpdate) {
		mu.Lock()
		gotUpdates = append(gotUpdates, u)
		mu.Unlock()
	})

	merged := makeTrack("trk-004", commonv1.TrackStatus_TRACK_STATUS_MERGED, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.5)
	cache.Put(merged)

	mu.Lock()
	defer mu.Unlock()
	if len(gotUpdates) != 1 {
		t.Fatalf("expected 1 onChange call, got %d", len(gotUpdates))
	}
	if gotUpdates[0].UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_MERGED {
		t.Errorf("expected UPDATE_TYPE_MERGED, got %v", gotUpdates[0].UpdateType)
	}
}

// T04: GetAll excludes DROPPED and MERGED tracks.
func TestTrackCache_GetAll_ExcludesInactive(t *testing.T) {
	cache := NewTrackCache(10)

	active := makeTrack("a1", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9)
	stale := makeTrack("a2", commonv1.TrackStatus_TRACK_STATUS_STALE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.5)
	newTrack := makeTrack("a3", commonv1.TrackStatus_TRACK_STATUS_NEW, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.3)
	dropped := makeTrack("a4", commonv1.TrackStatus_TRACK_STATUS_DROPPED, commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.0)
	merged := makeTrack("a5", commonv1.TrackStatus_TRACK_STATUS_MERGED, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.0)

	for _, tr := range []*entityv1.FusedTrack{active, stale, newTrack, dropped, merged} {
		cache.Put(tr)
	}

	all := cache.GetAll()
	if len(all) != 3 {
		t.Errorf("expected 3 active tracks, got %d", len(all))
	}
	for _, tr := range all {
		if tr.Status == commonv1.TrackStatus_TRACK_STATUS_DROPPED || tr.Status == commonv1.TrackStatus_TRACK_STATUS_MERGED {
			t.Errorf("GetAll returned inactive track %q with status %v", tr.TrackId, tr.Status)
		}
	}
}

// TestTrackCache_Get_NotFound: Get returns nil for unknown track.
func TestTrackCache_Get_NotFound(t *testing.T) {
	cache := NewTrackCache(10)
	if got := cache.Get("nonexistent"); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// TestTrackCache_Get_Found: Get returns the correct track.
func TestTrackCache_Get_Found(t *testing.T) {
	cache := NewTrackCache(10)
	tr := makeTrack("trk-get", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9)
	cache.Put(tr)
	got := cache.Get("trk-get")
	if got == nil {
		t.Fatal("expected track, got nil")
	}
	if got.TrackId != "trk-get" {
		t.Errorf("expected track_id=trk-get, got %q", got.TrackId)
	}
}

// TestTrackCache_Count: Count reflects active tracks only.
func TestTrackCache_Count(t *testing.T) {
	cache := NewTrackCache(10)
	cache.Put(makeTrack("c1", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9))
	cache.Put(makeTrack("c2", commonv1.TrackStatus_TRACK_STATUS_STALE, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.5))
	cache.Put(makeTrack("c3", commonv1.TrackStatus_TRACK_STATUS_DROPPED, commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.1))
	if got := cache.Count(); got != 2 {
		t.Errorf("expected Count=2, got %d", got)
	}
}

// TestTrackCache_GetHistory: History is populated and bounded.
func TestTrackCache_GetHistory(t *testing.T) {
	maxPts := 5
	cache := NewTrackCache(maxPts)

	// Insert 7 updates for the same track.
	for i := 0; i < 7; i++ {
		tr := makeTrack("hist-trk", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, float64(i)*0.1)
		cache.Put(tr)
	}

	hist := cache.GetHistory("hist-trk", 10)
	if len(hist) != maxPts {
		t.Errorf("expected %d history points (bounded), got %d", maxPts, len(hist))
	}
}

// TestTrackCache_GetHistory_NotFound: No history for unknown track.
func TestTrackCache_GetHistory_NotFound(t *testing.T) {
	cache := NewTrackCache(10)
	hist := cache.GetHistory("unknown", 10)
	if hist != nil {
		t.Errorf("expected nil history, got %v", hist)
	}
}

// TestTrackCache_Put_NilTrack: Nil track is ignored without panic.
func TestTrackCache_Put_NilTrack(t *testing.T) {
	cache := NewTrackCache(10)
	called := false
	cache.SetOnChange(func(u *entityv1.TrackUpdate) { called = true })
	cache.Put(nil)
	if called {
		t.Error("onChange must not be called for nil track")
	}
}

// TestTrackCache_Put_EmptyID: Track with empty ID is ignored without panic.
func TestTrackCache_Put_EmptyID(t *testing.T) {
	cache := NewTrackCache(10)
	called := false
	cache.SetOnChange(func(u *entityv1.TrackUpdate) { called = true })
	tr := makeTrack("", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.5)
	cache.Put(tr)
	if called {
		t.Error("onChange must not be called for empty-ID track")
	}
}

// TestTrackCache_Snapshot: Snapshot sends all active tracks.
func TestTrackCache_Snapshot(t *testing.T) {
	cache := NewTrackCache(10)
	cache.Put(makeTrack("s1", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.9))
	cache.Put(makeTrack("s2", commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_SURFACE, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.8))
	cache.Put(makeTrack("s3", commonv1.TrackStatus_TRACK_STATUS_DROPPED, commonv1.EntityType_ENTITY_TYPE_LAND, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.1))

	filter := &TrackFilter{
		ClearanceLevel: commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET,
	}

	var got []*entityv1.TrackUpdate
	err := cache.Snapshot(filter, func(u *entityv1.TrackUpdate) error {
		got = append(got, u)
		return nil
	})
	if err != nil {
		t.Fatalf("Snapshot error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 snapshot updates (active only), got %d", len(got))
	}
	for _, u := range got {
		if u.UpdateType != entityv1.TrackUpdate_UPDATE_TYPE_SNAPSHOT {
			t.Errorf("expected UPDATE_TYPE_SNAPSHOT, got %v", u.UpdateType)
		}
	}
}

// TestTrackCache_Concurrent: Verify no races under concurrent Put/Get.
func TestTrackCache_Concurrent(t *testing.T) {
	cache := NewTrackCache(20)
	cache.SetOnChange(func(u *entityv1.TrackUpdate) {}) // no-op

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "concurrent-" + string(rune('a'+i%26))
			cache.Put(makeTrack(id, commonv1.TrackStatus_TRACK_STATUS_ACTIVE, commonv1.EntityType_ENTITY_TYPE_AIR, commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED, 0.5))
			cache.Get(id)
			cache.Count()
		}(i)
	}
	wg.Wait()
}

// TestToProtoHistoryPoint: verifies proto conversion.
func TestToProtoHistoryPoint(t *testing.T) {
hp := &HistoryPoint{
Position:   &commonv1.Position{Latitude: 45.0, Longitude: -75.0},
Timestamp:  time.Now(),
Confidence: 0.88,
Status:     commonv1.TrackStatus_TRACK_STATUS_ACTIVE,
}
proto := ToProtoHistoryPoint(hp)
if proto == nil {
t.Fatal("expected non-nil proto")
}
if proto.Confidence != 0.88 {
t.Errorf("confidence mismatch: %v", proto.Confidence)
}
if proto.Status != commonv1.TrackStatus_TRACK_STATUS_ACTIVE {
t.Errorf("status mismatch: %v", proto.Status)
}
}

// TestNewTrackCache_DefaultHistoryMax: zero/negative historyMax uses default.
func TestNewTrackCache_DefaultHistoryMax(t *testing.T) {
c := NewTrackCache(0)
if c == nil {
t.Fatal("expected non-nil cache")
}
c2 := NewTrackCache(-5)
if c2 == nil {
t.Fatal("expected non-nil cache")
}
}
