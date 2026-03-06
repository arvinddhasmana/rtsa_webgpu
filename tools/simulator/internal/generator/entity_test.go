// CLASSIFICATION: UNCLASSIFIED
package generator_test

import (
"math/rand"
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/generator"
)

func testConfig() *config.SimulatorConfig {
cfg, _ := config.Load()
cfg.SurfaceEntityCount = 5
cfg.AirEntityCount = 3
cfg.SubEntityCount = 2
cfg.LandEntityCount = 2
cfg.CyberEntityCount = 1
cfg.AnomalyRate = 0.20
cfg.UpdateInterval = time.Second
return cfg
}

func TestNewEntityManager_Deterministic(t *testing.T) {
cfg := testConfig()
rng1 := rand.New(rand.NewSource(42))
rng2 := rand.New(rand.NewSource(42))

mgr1 := generator.NewEntityManager(cfg, rng1)
mgr2 := generator.NewEntityManager(cfg, rng2)

ents1 := mgr1.Entities()
ents2 := mgr2.Entities()

if len(ents1) != len(ents2) {
t.Fatalf("expected same entity count, got %d vs %d", len(ents1), len(ents2))
}

// Verify same IDs exist in both managers.
for id, e1 := range ents1 {
e2, ok := ents2[id]
if !ok {
t.Errorf("entity %q present in mgr1 but not mgr2", id)
continue
}
if e1.EntityType != e2.EntityType {
t.Errorf("entity %q: type mismatch %v vs %v", id, e1.EntityType, e2.EntityType)
}
}
}

func TestNewEntityManager_EntityCounts(t *testing.T) {
cfg := testConfig()
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)
entities := mgr.Entities()

total := cfg.SurfaceEntityCount + cfg.AirEntityCount + cfg.SubEntityCount +
cfg.LandEntityCount + cfg.CyberEntityCount
if len(entities) != total {
t.Errorf("expected %d entities, got %d", total, len(entities))
}

var surface, air, sub, land, cyber int
for _, e := range entities {
switch e.EntityType {
case commonv1.EntityType_ENTITY_TYPE_SURFACE:
surface++
case commonv1.EntityType_ENTITY_TYPE_AIR:
air++
case commonv1.EntityType_ENTITY_TYPE_SUBSURFACE:
sub++
case commonv1.EntityType_ENTITY_TYPE_LAND:
land++
case commonv1.EntityType_ENTITY_TYPE_CYBER:
cyber++
}
}

if surface != cfg.SurfaceEntityCount {
t.Errorf("expected %d surface entities, got %d", cfg.SurfaceEntityCount, surface)
}
if air != cfg.AirEntityCount {
t.Errorf("expected %d air entities, got %d", cfg.AirEntityCount, air)
}
if sub != cfg.SubEntityCount {
t.Errorf("expected %d sub entities, got %d", cfg.SubEntityCount, sub)
}
if land != cfg.LandEntityCount {
t.Errorf("expected %d land entities, got %d", cfg.LandEntityCount, land)
}
if cyber != cfg.CyberEntityCount {
t.Errorf("expected %d cyber entities, got %d", cfg.CyberEntityCount, cyber)
}
}

func TestEntitiesStayInBounds(t *testing.T) {
cfg := testConfig()
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)

// Run 100 ticks and verify all positions stay in bounds.
for tick := 0; tick < 100; tick++ {
mgr.Tick(time.Second)
for _, e := range mgr.Entities() {
if e.Position.Lat < generator.MinLat || e.Position.Lat > generator.MaxLat {
t.Errorf("entity %s lat %f out of bounds [%f, %f] at tick %d",
e.ID, e.Position.Lat, generator.MinLat, generator.MaxLat, tick)
}
if e.Position.Lon < generator.MinLon || e.Position.Lon > generator.MaxLon {
t.Errorf("entity %s lon %f out of bounds [%f, %f] at tick %d",
e.ID, e.Position.Lon, generator.MinLon, generator.MaxLon, tick)
}
}
}
}

func TestEntityManager_AnomalousEntitiesAssigned(t *testing.T) {
cfg := testConfig()
cfg.AnomalyRate = 1.0 // All entities anomalous
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)

for _, e := range mgr.Entities() {
if !e.IsAnomalous {
t.Errorf("entity %s expected to be anomalous but is not", e.ID)
}
if e.AnomalyType == commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED {
t.Errorf("entity %s is anomalous but has UNSPECIFIED anomaly type", e.ID)
}
}
}

func TestEntityManager_ZeroAnomalyRate(t *testing.T) {
cfg := testConfig()
cfg.AnomalyRate = 0.0
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)

for _, e := range mgr.Entities() {
if e.IsAnomalous {
t.Errorf("entity %s should not be anomalous with rate=0", e.ID)
}
}
}

func TestEntityManager_SurfaceSpeedRange(t *testing.T) {
cfg := testConfig()
cfg.AnomalyRate = 0.0
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)

for _, e := range mgr.Entities() {
if e.EntityType != commonv1.EntityType_ENTITY_TYPE_SURFACE {
continue
}
if e.Position.SpeedKn < 8 || e.Position.SpeedKn > 25 {
t.Errorf("surface entity %s speed %f outside [8,25] knots", e.ID, e.Position.SpeedKn)
}
}
}

func TestEntityManager_AirAltitudeRange(t *testing.T) {
cfg := testConfig()
cfg.AnomalyRate = 0.0
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)

for _, e := range mgr.Entities() {
if e.EntityType != commonv1.EntityType_ENTITY_TYPE_AIR {
continue
}
if e.Position.AltM < 1000 || e.Position.AltM > 15000 {
t.Errorf("air entity %s altitude %f outside [1000,15000] m", e.ID, e.Position.AltM)
}
}
}

func TestEntityManager_SubsurfaceDepthRange(t *testing.T) {
cfg := testConfig()
cfg.AnomalyRate = 0.0
rng := rand.New(rand.NewSource(42))
mgr := generator.NewEntityManager(cfg, rng)

for _, e := range mgr.Entities() {
if e.EntityType != commonv1.EntityType_ENTITY_TYPE_SUBSURFACE {
continue
}
if e.Position.AltM > -50 || e.Position.AltM < -400 {
t.Errorf("subsurface entity %s depth %f outside [-400,-50] m", e.ID, e.Position.AltM)
}
}
}
