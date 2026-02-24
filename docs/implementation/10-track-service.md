<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 10 — Track Service

> **Module**: 10-track-service
> **Phase**: P3 (Presentation)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries), Module 07 (fusion engine produces tracks)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 3 days

---

## 1. Objective

Implement the Track Service (`svc-track`) that consumes fused tracks from `tracks.fused.*` topics, maintains an in-memory cache of current track state, and exposes gRPC server-streaming to COP Web App clients for real-time track updates with filtering by entity type, hostile classification, bounding box, and classification clearance.

**Acceptance Criteria**:

- Consumes from all 5 `tracks.fused.*` topics (consumer group: `track-service`)
- In-memory cache of all active tracks indexed by `track_id`
- `StreamTracks` — server-streaming with initial snapshot then incremental updates
- `GetTrackDetails` — unary RPC returning full track with source attribution
- `GetTrackHistory` — returns recent position history from cache
- Classification filtering: only send tracks ≤ caller's clearance level
- Spatial filtering via bounding box
- Metrics: active track count, connected stream clients, update latency
- ≥80% line coverage

---

## 2. Service Structure

```
svc-track/
├── cmd/
│   └── track/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── track_cache.go           # In-memory track state cache
│   │   ├── track_cache_test.go
│   │   ├── filter.go                # Track filtering engine
│   │   └── filter_test.go
│   ├── consumer/
│   │   ├── fused_track_consumer.go  # Consumes tracks.fused.*
│   │   └── fused_track_consumer_test.go
│   ├── handler/
│   │   ├── stream.go                # StreamTracks handler
│   │   ├── details.go               # GetTrackDetails handler
│   │   ├── history.go               # GetTrackHistory handler
│   │   └── handler_test.go
│   └── mapper/
│       └── track_mapper.go          # TrackUpdate building
├── go.mod
├── Dockerfile
└── README.md
```

---

## 3. Track Cache

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// TrackCache maintains current state of all tracks in memory.
// Indexed by track_id for O(1) lookup.
// Also maintains spatial index for bounding box queries.
type TrackCache struct {
    mu       sync.RWMutex
    tracks   map[string]*CachedTrack
    history  map[string][]*HistoryPoint // Last 100 positions per track
    onChange func(update *entityv1.TrackUpdate) // Callback for stream notifications
}

// CachedTrack wraps a FusedTrack with cache metadata.
type CachedTrack struct {
    Track     *entityv1.FusedTrack
    CachedAt  time.Time
}

// Put adds or updates a track in the cache.
// Triggers onChange callback with appropriate UpdateType:
//   - Track not exists → UPDATE_TYPE_CREATED
//   - Track exists, status changed to DROPPED → UPDATE_TYPE_DROPPED
//   - Track exists, status changed to MERGED → UPDATE_TYPE_MERGED
//   - Track exists, otherwise → UPDATE_TYPE_UPDATED
func (c *TrackCache) Put(track *entityv1.FusedTrack) { /* implementation */ }

// Get returns a track by ID. Returns nil if not found.
func (c *TrackCache) Get(trackID string) *entityv1.FusedTrack { /* implementation */ }

// GetAll returns all active tracks (status: ACTIVE or STALE).
func (c *TrackCache) GetAll() []*entityv1.FusedTrack { /* implementation */ }

// GetFiltered returns tracks matching the filter criteria.
func (c *TrackCache) GetFiltered(filter *TrackFilter) []*entityv1.FusedTrack { /* implementation */ }

// GetHistory returns position history for a track.
func (c *TrackCache) GetHistory(trackID string, maxPoints int) []*HistoryPoint { /* implementation */ }

// Count returns the number of active tracks.
func (c *TrackCache) Count() int { /* implementation */ }

// HistoryPoint is a historical position entry.
type HistoryPoint struct {
    Position   *commonv1.Position
    Timestamp  time.Time
    Confidence float64
    Status     commonv1.TrackStatus
}
```

---

## 4. Track Filter

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// TrackFilter specifies filtering criteria for track queries.
type TrackFilter struct {
    EntityTypes    []commonv1.EntityType
    HostileClasses []commonv1.HostileClassification
    BoundingBox    *commonv1.BoundingBox
    MinConfidence  float64
    ClearanceLevel commonv1.ClassificationLevel
}

// FilterEngine applies filters to track collections.
type FilterEngine struct{}

// Apply filters a slice of tracks based on the given criteria.
// Filters are AND-combined:
//   1. Classification: track.classification ≤ clearanceLevel
//   2. EntityType: track.entity_type IN entityTypes (if non-empty)
//   3. HostileClass: track.hostile_class IN hostileClasses (if non-empty)
//   4. BoundingBox: track is within bbox (if non-nil)
//   5. MinConfidence: track.confidence_score ≥ minConfidence
func (f *FilterEngine) Apply(tracks []*entityv1.FusedTrack, filter *TrackFilter) []*entityv1.FusedTrack { /* implementation */ }

// InBoundingBox returns true if position is within the bounding box.
func InBoundingBox(lat, lon float64, bbox *commonv1.BoundingBox) bool { /* implementation */ }
```

---

## 5. StreamTracks Handler

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

// StreamTracksHandler implements TrackService.StreamTracks.
// Flow:
//   1. Parse StreamTracksRequest filters
//   2. Send initial SNAPSHOT: all current tracks matching filters
//   3. Subscribe to track cache onChange notifications
//   4. For each update:
//      a. Apply filter
//      b. If passes → send TrackUpdate to client stream
//   5. Continue until client disconnects or context cancelled
//
// Thread safety: uses channel-based fan-out from cache onChange.
// Each connected client gets its own filtered channel.
type StreamTracksHandler struct {
    cache   *domain.TrackCache
    filter  *domain.FilterEngine
    logger  *zap.Logger
    // Track connected clients
    clients sync.Map // clientID → chan *TrackUpdate
}

func (h *StreamTracksHandler) StreamTracks(
    req *entityv1.StreamTracksRequest,
    stream entityv1.TrackService_StreamTracksServer) error {
    // 1. Build filter from request
    // 2. Get all current tracks, filter, send as SNAPSHOT
    // 3. Register client channel
    // 4. defer: unregister client channel
    // 5. Loop: receive from channel, filter, send to stream
}
```

---

## 6. Metrics

| Metric                                             | Type      | Labels                       |
| -------------------------------------------------- | --------- | ---------------------------- |
| `rtsa_track_service_active_tracks`                 | Gauge     | `entity_type`, `status`      |
| `rtsa_track_service_stream_clients`                | Gauge     | -                            |
| `rtsa_track_service_updates_sent_total`            | Counter   | `entity_type`, `update_type` |
| `rtsa_track_service_cache_update_duration_seconds` | Histogram | -                            |

---

## 7. Test Scenarios

| #   | Test                                                       | Expected                             |
| --- | ---------------------------------------------------------- | ------------------------------------ |
| T01 | Put new track → onChange with CREATED                      | Callback invoked                     |
| T02 | Put existing track → onChange with UPDATED                 | Callback invoked                     |
| T03 | Put track with DROPPED status → onChange with DROPPED      | Callback invoked                     |
| T04 | GetAll excludes DROPPED tracks                             | Only ACTIVE/STALE returned           |
| T05 | Filter by entity type                                      | Only matching types returned         |
| T06 | Filter by classification clearance                         | Higher classified excluded           |
| T07 | Filter by bounding box                                     | Only within bbox returned            |
| T08 | Filter by min confidence                                   | Low confidence excluded              |
| T09 | StreamTracks: initial snapshot                             | All matching tracks sent as SNAPSHOT |
| T10 | StreamTracks: incremental update                           | Update sent after cache change       |
| T11 | GetTrackDetails: existing track                            | Full track returned                  |
| T12 | GetTrackDetails: non-existent track                        | NOT_FOUND error                      |
| T13 | GetTrackHistory: returns history points                    | Correct positions                    |
| T14 | Classification filter: SECRET track, PROTECTED_B clearance | Track excluded                       |

---

## 8. Agent Invocation

```
@greatest-ever-developer Implement Module 10 from docs/implementation/10-track-service.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for TrackService proto
- Read docs/architecture/component_design.md §8.1 for track service component diagram
- This is a presentation service: consumes Redpanda, serves gRPC to UI
- In-memory cache only (no persistence)
- StreamTracks uses fan-out pattern: each client gets filtered updates
- Classification filtering is MANDATORY on all responses

Deliverables:
1. Complete svc-track/ with all files
2. In-memory track cache with onChange callback
3. Filter engine with 5 filter criteria
4. StreamTracks with initial snapshot + incremental updates
5. GetTrackDetails and GetTrackHistory handlers
6. Unit tests (≥80% coverage)
7. Integration tests with testcontainers
```
