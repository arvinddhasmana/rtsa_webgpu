// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"context"
	"flag"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/client"
	"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/generator"
	"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/scenario"
	"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/sensor"
)

// consecutiveFailures tracks how many gRPC sends have failed without a single
// success in between. After consecutiveFailureThreshold the simulator escalates
// from Warn to Error so operators notice immediately.
var consecutiveFailures atomic.Int64

const consecutiveFailureThreshold = 10

func main() {
// ── CLI flags ────────────────────────────────────────────────────────────
scenarioFile := flag.String("scenario", "", "Path to YAML scenario file")
surfaceEntities := flag.Int("surface-entities", -1, "Number of surface entities")
airEntities := flag.Int("air-entities", -1, "Number of air entities")
updateInterval := flag.Duration("update-interval", 0, "Update interval (e.g. 500ms, 1s)")
anomalyRate := flag.Float64("anomaly-rate", -1, "Anomaly injection rate (0.0-1.0)")
seed := flag.Int64("seed", -1, "Random seed (0 = random, positive = deterministic)")
duration := flag.Duration("duration", 0, "Run duration (0 = infinite)")
dryRun := flag.Bool("dry-run", false, "Generate observations but do not send via gRPC")
flag.Parse()

// ── Logger ───────────────────────────────────────────────────────────────
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// ── Base config from environment variables ───────────────────────────────
cfg, err := config.Load()
if err != nil {
slog.Error("failed to load configuration", "error", err)
os.Exit(1)
}

// ── Scenario file overrides ───────────────────────────────────────────────
scenarioPath := cfg.ScenarioFile
if *scenarioFile != "" {
scenarioPath = *scenarioFile
}
if scenarioPath != "" {
sc, loadErr := scenario.LoadFromFile(scenarioPath)
if loadErr != nil {
slog.Error("failed to load scenario", "path", scenarioPath, "error", loadErr)
os.Exit(1)
}
slog.Info("loaded scenario", "name", sc.Name, "description", sc.Description)
applyScenarioToConfig(cfg, sc)
}

// ── CLI flag overrides (highest priority) ─────────────────────────────────
if *surfaceEntities >= 0 {
cfg.SurfaceEntityCount = *surfaceEntities
}
if *airEntities >= 0 {
cfg.AirEntityCount = *airEntities
}
if *updateInterval > 0 {
cfg.UpdateInterval = *updateInterval
}
if *anomalyRate >= 0 {
cfg.AnomalyRate = *anomalyRate
}
if *seed >= 0 {
cfg.RandomSeed = *seed
}
if *duration > 0 {
cfg.DurationMinutes = int(duration.Minutes())
}

// ── RNG ───────────────────────────────────────────────────────────────────
var simRNG *rand.Rand
if cfg.RandomSeed == 0 {
simRNG = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
} else {
simRNG = rand.New(rand.NewSource(cfg.RandomSeed))
}
sensor.SetRNG(simRNG)

// ── Entity manager ────────────────────────────────────────────────────────
mgr := generator.NewEntityManager(cfg, simRNG)
slog.Info("entity manager initialised",
"surface", cfg.SurfaceEntityCount,
"air", cfg.AirEntityCount,
"subsurface", cfg.SubEntityCount,
"land", cfg.LandEntityCount,
"cyber", cfg.CyberEntityCount,
"anomaly_rate", cfg.AnomalyRate,
)

// ── Sender setup ──────────────────────────────────────────────────────────
var sender client.ObservationSender
if *dryRun {
sender = &noopSender{}
slog.Info("dry-run mode: observations will not be sent over gRPC")
} else {
grpcSender, senderErr := client.NewGRPCSender(cfg)
if senderErr != nil {
slog.Error("failed to create gRPC sender", "error", senderErr)
os.Exit(1)
}
sender = grpcSender
defer func() {
if closeErr := grpcSender.Close(); closeErr != nil {
slog.Warn("error closing gRPC sender", "error", closeErr)
}
}()
}

// ── Signal handling and duration deadline ─────────────────────────────────
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

if cfg.DurationMinutes > 0 {
deadline := time.Now().Add(time.Duration(cfg.DurationMinutes) * time.Minute)
var deadlineCancel context.CancelFunc
ctx, deadlineCancel = context.WithDeadline(ctx, deadline)
defer deadlineCancel()
slog.Info("simulation duration set", "minutes", cfg.DurationMinutes)
}

slog.Info("simulator started", "update_interval", cfg.UpdateInterval)

// ── Main simulation loop ──────────────────────────────────────────────────
ticker := time.NewTicker(cfg.UpdateInterval)
defer ticker.Stop()

for {
select {
case <-ctx.Done():
slog.Info("simulator stopping", "reason", ctx.Err())
return
case <-ticker.C:
mgr.Tick(cfg.UpdateInterval)
runTick(ctx, mgr, sender, simRNG)
}
}
}

// runTick generates and dispatches sensor observations for one simulation tick.
func runTick(ctx context.Context, mgr *generator.EntityManager, sender client.ObservationSender, r *rand.Rand) {
	entities := mgr.Entities()
	sent := 0
	failed := 0

	radarIDs := []string{"RADAR-SIM-001", "RADAR-SIM-002", "RADAR-SIM-003"}
	ewIDs := []string{"EW-SIM-001", "EW-SIM-002"}

	for _, e := range entities {
		// Radar: surface + air entities.
		if e.EntityType == commonv1.EntityType_ENTITY_TYPE_SURFACE ||
			e.EntityType == commonv1.EntityType_ENTITY_TYPE_AIR {
			rid := radarIDs[r.Intn(len(radarIDs))]
			sendObs(ctx, sensor.GenerateRadarObservation(e, rid), client.SensorTypeRadar, sender, &sent, &failed)
		}

		// AIS: surface entities only. AIS manipulation generates both radar AND ais.
		if e.EntityType == commonv1.EntityType_ENTITY_TYPE_SURFACE {
			var manipPos *generator.Position
			if e.IsAnomalous && e.AnomalyType == commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION {
				p := e.AISOffset
				manipPos = &p
			}
			sendObs(ctx, sensor.GenerateAISObservation(e, manipPos), client.SensorTypeAIS, sender, &sent, &failed)
		}

		// EW: 50% probability per entity per tick.
		if r.Float64() < 0.5 {
			ewID := ewIDs[r.Intn(len(ewIDs))]
			sendObs(ctx, sensor.GenerateEWObservation(e, ewID), client.SensorTypeEW, sender, &sent, &failed)
		}

		// ELINT: 30% probability.
		if r.Float64() < 0.3 {
			sendObs(ctx, sensor.GenerateELINTObservation(e, "ELINT-SIM-001"), client.SensorTypeELINT, sender, &sent, &failed)
		}

		// ISR: 20% probability for all; always for LAND entities.
		if r.Float64() < 0.2 || e.EntityType == commonv1.EntityType_ENTITY_TYPE_LAND {
			sendObs(ctx, sensor.GenerateISRObservation(e, "ISR-SIM-001"), client.SensorTypeISR, sender, &sent, &failed)
		}

		// EW: always for CYBER entities (RF signature of compromised nodes).
		if e.EntityType == commonv1.EntityType_ENTITY_TYPE_CYBER && r.Float64() < 0.7 {
			ewID := ewIDs[r.Intn(len(ewIDs))]
			sendObs(ctx, sensor.GenerateEWObservation(e, ewID), client.SensorTypeEW, sender, &sent, &failed)
		}
	}

	// Cyber: 2-5 IOCs per tick (increased for better CYBER domain visibility).
	for i := 0; i < 2+r.Intn(4); i++ {
		sendObs(ctx, sensor.GenerateCyberObservation(r), client.SensorTypeCyber, sender, &sent, &failed)
	}

	if failed > 0 {
		slog.Warn("tick completed with failures",
			"entities", len(entities),
			"sent", sent,
			"failed", failed,
		)
	} else {
		slog.Debug("tick complete", "entities", len(entities), "observations", sent)
	}
}

// sendObs dispatches one observation, increments count on success, and tracks
// consecutive failures. After consecutiveFailureThreshold failures without a
// single success the log level escalates from Warn → Error so the problem is
// impossible to miss.
func sendObs(ctx context.Context, obs *ingestionv1.SensorObservation, st client.SensorType, sender client.ObservationSender, count *int, failed *int) {
	if err := sender.Send(ctx, obs, st); err != nil {
		*failed++
		n := consecutiveFailures.Add(1)
		if n >= consecutiveFailureThreshold {
			slog.Error("gRPC sends failing — are ingestion services running?",
				"sensor_type", string(st),
				"consecutive_failures", n,
				"error", err,
			)
		} else {
			slog.Warn("send failed", "sensor_type", string(st), "consecutive_failures", n, "error", err)
		}
		return
	}
	consecutiveFailures.Store(0)
	*count++
}

// applyScenarioToConfig copies relevant scenario fields into the config.
func applyScenarioToConfig(cfg *config.SimulatorConfig, sc *scenario.Scenario) {
if sc.Entities.Surface.Count > 0 {
cfg.SurfaceEntityCount = sc.Entities.Surface.Count
}
if sc.Entities.Air.Count > 0 {
cfg.AirEntityCount = sc.Entities.Air.Count
}
if sc.Entities.Subsurface.Count > 0 {
cfg.SubEntityCount = sc.Entities.Subsurface.Count
}
// Land and Cyber counts may be set via env var; scenario YAML can override
// them once the scenario format is extended. Defaulting to 0 override here.
if sc.Anomalies.InjectionRate > 0 {
cfg.AnomalyRate = sc.Anomalies.InjectionRate
}
if sc.Seed != 0 {
cfg.RandomSeed = sc.Seed
}
if sc.DurationMin > 0 {
cfg.DurationMinutes = sc.DurationMin
}
}

// noopSender discards observations (dry-run mode).
type noopSender struct{}

func (n *noopSender) Send(_ context.Context, obs *ingestionv1.SensorObservation, st client.SensorType) error {
slog.Debug("dry-run: observation generated",
"sensor_type", string(st),
"observation_id", obs.GetObservationId(),
"sensor_id", obs.GetSensorId(),
)
return nil
}

func (n *noopSender) Close() error { return nil }
