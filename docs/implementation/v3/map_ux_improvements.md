<!-- CLASSIFICATION: UNCLASSIFIED -->
# Fusion Dashboard — Map UX Improvements Plan

## Current Problems

1. **All track circles are identical amber dots** — simulator generates only `hostileClass: UNKNOWN`
2. **No domain differentiation** — air, surface, subsurface, land all render as same circle
3. **No labels** — can't identify tracks without clicking
4. **Red rectangle** — hardcoded geo-fence with bad opacity
5. **Layer toggles are stubs** — Track Labels, Track Trails, Sensor Coverage, MGRS Grid do nothing

## Improvement Roadmap (Priority Order)

### A. Domain Shape Differentiation ← IMPLEMENTING NOW
Different icon/shape per domain, color still driven by hostile classification:

| Domain | Shape | Color (hostile) | Color (friendly) | Color (unknown) |
|---|---|---|---|---|
| SURFACE (ships) | Ship icon | Red | Blue | Amber |
| AIR (aircraft) | Aircraft icon | Red | Blue | Amber |
| SUBSURFACE | Submarine icon | Red | Blue | Amber |
| LAND (vehicles) | Vehicle icon | Red | Blue | Amber |
| UNKNOWN | Circle | Red | Blue | Amber |

### B. Track Labels (Toggle)
- Abbreviated info next to marker: `AIS-7c5 | SURFACE | 90%`

### C. Track Trail Breadcrumbs (Toggle)
- Fading polyline showing last N positions for each moving track

### D. Size by Confidence
- Higher confidence = larger marker (6–10px)

### E. Simulator Data Diversity ← IMPLEMENTING NOW
- Generate HOSTILE, FRIENDLY, NEUTRAL tracks (not just UNKNOWN)
- Generate AIR, SUBSURFACE, LAND tracks (not just SURFACE)

### F. Sensor Coverage Arcs (Toggle)
- Fan-shaped sectors from sensor positions showing detection range
- Color by sensor type (RADAR=green, EW=purple, AIS=blue)

### G. Fix Geo-fence
- Proper semi-transparent styling, working layer toggle
