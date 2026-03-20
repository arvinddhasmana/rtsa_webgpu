# Plan: MIL-STD-2525 Symbology + West Asia Demo

## TL;DR

Upgrade the track icon rendering from simplified domain-based shapes to full MIL-STD-2525 affiliation frames with internal domain icons. Add Military/Civilian context distinction. Replace the generic mock data with a 150-track West Asia/Iran demo scenario with sensors at named strategic locations. Default map viewport opens centered on Persian Gulf.

## Current State Summary

- **Symbology**: 3 procedural SDF shapes keyed by `icon_index % 3` (Air=triangle, Surface=diamond, Sub=ellipse). Color from `threat_level` (Blue/Red/Yellow). No affiliation-based frames, no MIL-STD-2525 compliance.
- **icon_index formula**: `entityType * 6 + threatLevel` (Go serializer). Only `icon_index % 3` used in shader — loses entity/threat info.
- **Mock data**: 30 tracks, globally scattered, random threat/domain.
- **Viewport**: center=(0°,0°), zoom=2 (whole world).
- **Simulator**: Hardcoded North Atlantic bounds (43-47°N, -65 to -55°W). OperationalArea parsed but unused.
- **Atlas**: 2048×2048 placeholder (solid white). Not used — shapes are procedural SDF.
- **No TrackContext** (Military/Civilian) in proto or record.

## Decisions

- Full MIL-STD-2525 affiliation frames WITH domain icons inside (user confirmed)
- Include Land domain entities (user confirmed)
- ~150 fused tracks for demo (user confirmed)
- Sensors at named strategic locations (user confirmed)
- Context (Military/Civilian) encoded in `icon_index` field: `context * 36 + entityType * 6 + threatLevel` — no record layout change needed
- Procedural SDF rendering (no atlas texture dependency for phase 1) — keeps current approach but with proper MIL-STD shapes
- Pick shader uses QUAD size not icon_index for hit testing — no pick.wgsl changes needed

## Phase 1: Symbology Types & Constants (no deps)

### Step 1.1: Add TrackSymbol types to frontend

- Create `web-cop-gpu/src/types/symbology.ts` with:
  - `TrackDomain` enum: AIR, LAND, VESSEL, SUBSURFACE
  - `TrackAffiliation` enum: FRIENDLY, HOSTILE, UNKNOWN, NEUTRAL, SUSPECT, PENDING
  - `TrackContext` enum: MILITARY, CIVILIAN
  - `TrackSymbolProps` interface
  - Mapping functions: `entityTypeToTrackDomain()`, `threatLevelToAffiliation()`
  - `encodeIconIndex(context, entityType, threatLevel)` and `decodeIconIndex(iconIndex)` helpers

### Step 1.2: Add TrackContext to proto (optional — for future pipeline use)

- Add `TrackContext` enum to `proto/rtsa/common/v1/types.proto`: MILITARY=0, CIVILIAN=1
- **Skip for now** — not needed for demo. Context derived from mock data only. Can add later.

## Phase 2: WGSL Shader Rewrite — MIL-STD-2525 Symbology

### Step 2.1: Rewrite `track-icons.wgsl` fragment shader

**File**: `web-cop-gpu/src/shaders/track-icons.wgsl`

Replace `get_silhouette()` and `fs_main()` with MIL-STD-2525 rendering:

**Frame shapes** (SDF functions keyed by `threat_level` / affiliation):

- `sdf_rectangle(p)` → Friendly (threat_level=2): Rounded rectangle
- `sdf_diamond(p)` → Hostile (threat_level=5) & Suspect (4): 45° rotated square
- `sdf_quatrefoil(p)` → Unknown (threat_level=0): 4-lobed clover shape
- `sdf_circle(p)` → Neutral (threat_level=3): Circle
- `sdf_pending(p)` → Pending (threat_level=1): Quatrefoil with dashed outline

**Domain icons** (inner silhouettes keyed by `entity_type` extracted from `icon_index`):

- Air (entity_type=2): Upward arrow/bird shape
- Surface/Vessel (entity_type=1): Ship bow silhouette
- Subsurface (entity_type=3): Submarine ellipse
- Land (entity_type=4): Rectangular vehicle shape
- Cyber (entity_type=5): Lightning bolt or hex shape

**Context rendering**:

- Military (`icon_index / 36 == 0`): Solid fill with affiliation color
- Civilian (`icon_index / 36 == 1`): Outline frame only (transparent fill)

**Affiliation colors** (MIL-STD-2525 standard):

- Friendly: Cyan-blue `(0.22, 0.74, 1.0)` — #38BDFF
- Hostile: Red `(0.97, 0.44, 0.44)` — #F87171
- Unknown: Yellow `(0.98, 0.75, 0.14)` — #FBBF24
- Neutral: Green `(0.34, 0.90, 0.53)` — #57E688
- Suspect: Orange `(1.0, 0.6, 0.2)` — #FF9933
- Pending: Light blue `(0.5, 0.8, 1.0)` — dashed frame

**Vertex shader**: Keep current quad-based instanced approach. Increase `ICON_BASE_SIZE_PX` from 14 to 20 (MIL frames need more space for internal icons).

### Step 2.2: Update `halos.wgsl` if needed

- Match new affiliation color palette for alert glow colors
- Currently references `threat_level` — should still work

### Step 2.3: Update `pick.wgsl` if needed

- Pick shader uses fixed QUAD radius (20px) and writes `track_id_hash` — doesn't use icon_index for shape. **No changes needed.**

## Phase 3: Mock Data — West Asia/Iran Demo (_parallel with Phase 2_)

### Step 3.1: Rewrite `mock-data.ts` for West Asia scenario

**File**: `web-cop-gpu/src/gpu/mock-data.ts`

**Geographic bounds** (Persian Gulf / Gulf of Oman / Arabian Sea):

- LAT: 22°N to 32°N → radians: 0.3839 to 0.5585
- LON: 46°E to 62°E → radians: 0.8029 to 1.0821

**~150 fused tracks distributed as**:

- Surface/Vessel: 55 tracks (tankers, warships, fishing vessels in Persian Gulf, Strait of Hormuz, Gulf of Oman)
- Air: 40 tracks (military jets, commercial aircraft, UAVs near Iranian airspace & UAE)
- Land: 30 tracks (military vehicles along Iranian coast, SAM sites, UAE/Oman ground forces)
- Subsurface: 15 tracks (submarines in Gulf of Oman approaches)
- Cyber: 10 tracks (positioned at major urban centers — Tehran, Dubai, Riyadh)

**Affiliation distribution**:

- Friendly (threat_level=2): ~45% — coalition naval/air, UAE/Oman military
- Hostile (threat_level=5): ~10% — Iranian military assets w/ aggressive posture
- Neutral (threat_level=3): ~20% — commercial shipping, civilian flights
- Unknown (threat_level=0): ~12% — unidentified contacts
- Suspect (threat_level=4): ~8% — vessels with AIS inconsistencies
- Pending (threat_level=1): ~5% — newly detected contacts

**Context distribution**:

- Military: ~55% (warships, military aircraft, ground forces)
- Civilian: ~45% (cargo ships, tankers, commercial flights, fishing boats)

**Sensor positions** (7 sensors at named strategic locations):

1. `RADAR-HORMUZ-001` @ Strait of Hormuz (26.6°N, 56.3°E) — 150nm range
2. `RADAR-BAHRAIN-002` @ Bahrain (26.2°N, 50.6°E) — 120nm range
3. `AIS-DUBAI-001` @ Dubai/UAE Coast (25.1°N, 55.1°E) — AIS receiver
4. `EW-MUSCAT-001` @ Muscat, Oman (23.6°N, 58.5°E) — SIGINT
5. `ELINT-CHABAHAR-001` @ Chabahar, Iran (25.3°N, 60.6°E) — ELINT monitoring
6. `ISR-BANDARABBAS-001` @ Bandar Abbas, Iran (27.2°N, 56.3°E) — ISR platform
7. `RADAR-QESHM-003` @ Qeshm Island (26.9°N, 55.9°E) — coastal radar

**Track generation**: Pre-generate named tracks with realistic positions:

- Surface tracks clustered along shipping lanes (Hormuz chokepoint, Gulf approaches)
- Air tracks along known air corridors over the Gulf + Iranian airspace
- Land tracks along Iranian coastline, UAE border, Oman border
- Subsurface tracks in Gulf of Oman deep water

### Step 3.2: Update `source_bitmap` in mock data

- Encode sensor sources per track using the 7 sensors (bit 1=Radar, 2=EW, 3=ELINT, 4=ISR, 5=AIS, 6=Cyber)
- Fused tracks typically have 2-3 contributing sensors

## Phase 4: Default Viewport — West Asia Focus

### Step 4.1: Update `viewport.ts` defaults

**File**: `web-cop-gpu/src/signals/viewport.ts`

- Change default: `centerLat: 27.0, centerLon: 54.0, zoom: 6`
- This centers the map on the Persian Gulf showing Iran, UAE, Oman, Bahrain, Qatar

### Step 4.2: Update `LeafletMap.tsx` initial center

**File**: `web-cop-gpu/src/components/map/LeafletMap.tsx`

- Already reads from `viewport()` signal on mount — should auto-follow the signal change

## Phase 5: Backend — icon_index Encoding Update

### Step 5.1: Update Go serializer `buildIconIndex`

**File**: `pkg/flatbuf/serializer.go`

- Change formula to: `context * 36 + entityType * 6 + threatLevel`
- For now, derive context from entity type or default to MILITARY (0) — existing FusedTrack proto doesn't have context field
- Comment the new encoding clearly

### Step 5.2: Add context field to FusedTrack proto (future)

- Not needed for demo — context only matters in mock data and shader
- When real pipeline data flows, add `TrackContext context = N` to FusedTrack message

## Phase 6: Simulator Scenario — West Asia (_parallel with Phase 5_)

### Step 6.1: Create `west-asia-demo.yaml`

**File**: `tools/simulator/scenarios/west-asia-demo.yaml`

- Center: 27°N, 54°E
- 150 entities: 55 surface, 40 air, 30 land, 15 subsurface, 10 cyber
- Sensors at the 7 named locations
- Anomaly rate: 0.08 (moderate)
- Seed: 20240815 (deterministic)

### Step 6.2: Make generator support configurable bounds

**File**: `tools/simulator/internal/generator/entity.go`

- Read `OperationalArea` from config (currently parsed but ignored)
- If OperationalArea.Center is set, derive MinLat/MaxLat/MinLon/MaxLon from center + radius
- Fallback to current hardcoded North Atlantic bounds if not set

### Step 6.3: Update demo seed script

**File**: `scripts/demo/seed-demo-data.sh`

- Add West Asia variant or make scenario configurable

## Phase 7: Legend & UI Updates

### Step 7.1: Update `SensorLegend.tsx`

**File**: `web-cop-gpu/src/components/dashboard/SensorLegend.tsx`

- Update SVG icons to match MIL-STD-2525 frame shapes
- Add all 4 affiliations × 4 domains (Air, Surface, Sub, Land)
- Add Military vs Civilian distinction
- Use correct affiliation colors

### Step 7.2: Add `TrackDetailPanel` context display

**File**: `web-cop-gpu/src/components/panels/TrackDetailPanel.tsx`

- Show domain, affiliation, context in track detail view

## Phase 8: Tests

### Step 8.1: Unit tests for symbology types

- Test `encodeIconIndex` / `decodeIconIndex` roundtrip
- Test `entityTypeToTrackDomain` mapping

### Step 8.2: Update mock-data tests

- Verify 150 tracks generated
- Verify geographic bounds (all tracks within West Asia region)
- Verify affiliation distribution

### Step 8.3: Update existing track rendering tests

- Adapt any tests that reference old icon shapes or threat colors

---

## Relevant Files

### Must Modify

- `web-cop-gpu/src/shaders/track-icons.wgsl` — Complete MIL-STD-2525 frame/icon rewrite
- `web-cop-gpu/src/gpu/mock-data.ts` — 150-track West Asia demo data
- `web-cop-gpu/src/signals/viewport.ts` — Default viewport to Persian Gulf (27°N, 54°E, zoom 6)
- `web-cop-gpu/src/components/dashboard/SensorLegend.tsx` — Updated legend with MIL-STD-2525 symbols
- `pkg/flatbuf/serializer.go` — Update `buildIconIndex` for context encoding

### Must Create

- `web-cop-gpu/src/types/symbology.ts` — TrackDomain, TrackAffiliation, TrackContext enums
- `tools/simulator/scenarios/west-asia-demo.yaml` — New scenario config

### Should Modify

- `tools/simulator/internal/generator/entity.go` — Support OperationalArea bounds
- `web-cop-gpu/src/components/panels/TrackDetailPanel.tsx` — Show context/domain/affiliation
- `web-cop-gpu/src/shaders/halos.wgsl` — Match new affiliation colors (optional)

### Reference Only (no changes)

- `web-cop-gpu/src/shaders/pick.wgsl` — Uses quad radius, not icon shape
- `web-cop-gpu/src/gpu/atlas.ts` — Placeholder; procedural SDF approach continues
- `web-cop-gpu/src/gpu/buffers.ts` — 128-byte record layout unchanged
- `pkg/flatbuf/layout.go` — Byte offsets unchanged

## Verification

1. **Visual**: Open browser → map auto-centers on West Asia (Persian Gulf visible), ~150 tracks rendered with distinct MIL-STD-2525 frames
2. **Frame shapes**: Friendly=blue rectangle, Hostile=red diamond, Unknown=yellow quatrefoil, Neutral=green circle, Suspect=orange diamond
3. **Domain icons**: Air tracks show upward arrow inside frame, Surface shows ship, Subsurface shows oval, Land shows rectangle
4. **Context**: Military tracks have filled frames, civilian tracks have outline-only frames
5. **Interaction**: Click on track → pick buffer returns correct track_id_hash → detail panel shows domain/affiliation/context
6. **Selection glow**: Selected track shows cyan ring (existing behavior preserved)
7. **Alert pulsing**: Suspect/Hostile tracks with alerts pulse correctly
8. **Simulator**: `west-asia-demo.yaml` runs, generates tracks in correct bounds
9. **Unit tests**: `pnpm test` passes in `web-cop-gpu/`
10. **Go tests**: `go test ./pkg/flatbuf/...` passes with updated icon_index encoding
11. **Performance**: 150 tracks render at 60 FPS with < 20% main thread CPU

## Further Considerations

1. **Atlas texture vs procedural SDF**: Current plan continues using procedural SDF shapes in the fragment shader. For production, baked NATO APP-6 icon atlas (64×64 per icon, pre-rendered) would be higher fidelity. This can be Phase 2 work.
2. **TrackContext in proto**: Adding `context` field to `FusedTrack` proto requires `buf generate`, updating all consumers. Defer to a separate PR.
3. **Suspect vs Unknown frame**: MIL-STD-2525D uses the same diamond frame for Hostile and Suspect (distinguished by fill/dashing). The plan uses distinct colors (red vs orange) which is more visually clear for operators.
