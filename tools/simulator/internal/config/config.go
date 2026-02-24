// CLASSIFICATION: UNCLASSIFIED
package config

import (
"fmt"
"os"
"strconv"
"strings"
"time"
)

// SimulatorConfig holds all configurable parameters for the simulator.
type SimulatorConfig struct {
// Target ingestion service endpoints
RadarEndpoint string
EWEndpoint    string
ELINTEndpoint string
ISREndpoint   string
AISEndpoint   string
CyberEndpoint string

// Entity configuration
SurfaceEntityCount int
AirEntityCount     int
SubEntityCount     int

// Timing
UpdateInterval  time.Duration
DurationMinutes int // 0 = infinite

// Anomaly injection
AnomalyRate  float64 // fraction of entities that are anomalous
AnomalyTypes []string

// Reproducibility
RandomSeed   int64  // 0 = random
ScenarioFile string // path to YAML scenario file

// TLS
TLSEnabled  bool
TLSCertFile string
TLSKeyFile  string
TLSCAFile   string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*SimulatorConfig, error) {
cfg := &SimulatorConfig{
RadarEndpoint:      getEnvStr("SIM_RADAR_ENDPOINT", "localhost:50051"),
EWEndpoint:         getEnvStr("SIM_EW_ENDPOINT", "localhost:50052"),
ELINTEndpoint:      getEnvStr("SIM_ELINT_ENDPOINT", "localhost:50053"),
ISREndpoint:        getEnvStr("SIM_ISR_ENDPOINT", "localhost:50054"),
AISEndpoint:        getEnvStr("SIM_AIS_ENDPOINT", "localhost:50055"),
CyberEndpoint:      getEnvStr("SIM_CYBER_ENDPOINT", "localhost:50056"),
SurfaceEntityCount: getEnvInt("SIM_SURFACE_ENTITIES", 20),
AirEntityCount:     getEnvInt("SIM_AIR_ENTITIES", 10),
SubEntityCount:     getEnvInt("SIM_SUB_ENTITIES", 5),
UpdateInterval:     getEnvDuration("SIM_UPDATE_INTERVAL_MS", 1000*time.Millisecond),
DurationMinutes:    getEnvInt("SIM_DURATION_MINUTES", 0),
AnomalyRate:        getEnvFloat("SIM_ANOMALY_RATE", 0.05),
AnomalyTypes:       strings.Split(getEnvStr("SIM_ANOMALY_TYPES", "speed,route_deviation,ais_manipulation,behavioral,proximity"), ","),
RandomSeed:         getEnvInt64("SIM_RANDOM_SEED", 0),
ScenarioFile:       getEnvStr("SIM_SCENARIO_FILE", ""),
TLSEnabled:         getEnvBool("SIM_TLS_ENABLED", false),
TLSCertFile:        getEnvStr("SIM_TLS_CERT_FILE", ""),
TLSKeyFile:         getEnvStr("SIM_TLS_KEY_FILE", ""),
TLSCAFile:          getEnvStr("SIM_TLS_CA_FILE", ""),
}

if err := cfg.validate(); err != nil {
return nil, err
}
return cfg, nil
}

// validate checks that the configuration is coherent.
func (c *SimulatorConfig) validate() error {
if c.SurfaceEntityCount < 0 {
return fmt.Errorf("surface entity count must be non-negative")
}
if c.AirEntityCount < 0 {
return fmt.Errorf("air entity count must be non-negative")
}
if c.SubEntityCount < 0 {
return fmt.Errorf("sub entity count must be non-negative")
}
if c.AnomalyRate < 0.0 || c.AnomalyRate > 1.0 {
return fmt.Errorf("anomaly rate must be between 0.0 and 1.0, got %f", c.AnomalyRate)
}
if c.UpdateInterval <= 0 {
return fmt.Errorf("update interval must be positive")
}
if c.TLSEnabled && (c.TLSCertFile == "" || c.TLSKeyFile == "" || c.TLSCAFile == "") {
return fmt.Errorf("TLS enabled but cert/key/CA files not provided")
}
return nil
}

// getEnvStr returns the value of the named env var, or the default.
func getEnvStr(key, def string) string {
if v := os.Getenv(key); v != "" {
return v
}
return def
}

// getEnvInt returns the integer value of the named env var, or the default.
func getEnvInt(key string, def int) int {
if v := os.Getenv(key); v != "" {
i, err := strconv.Atoi(v)
if err == nil {
return i
}
}
return def
}

// getEnvInt64 returns the int64 value of the named env var, or the default.
func getEnvInt64(key string, def int64) int64 {
if v := os.Getenv(key); v != "" {
i, err := strconv.ParseInt(v, 10, 64)
if err == nil {
return i
}
}
return def
}

// getEnvFloat returns the float64 value of the named env var, or the default.
func getEnvFloat(key string, def float64) float64 {
if v := os.Getenv(key); v != "" {
f, err := strconv.ParseFloat(v, 64)
if err == nil {
return f
}
}
return def
}

// getEnvBool returns the boolean value of the named env var, or the default.
func getEnvBool(key string, def bool) bool {
if v := os.Getenv(key); v != "" {
b, err := strconv.ParseBool(v)
if err == nil {
return b
}
}
return def
}

// getEnvDuration returns a duration derived from an env var holding milliseconds.
func getEnvDuration(key string, def time.Duration) time.Duration {
if v := os.Getenv(key); v != "" {
i, err := strconv.ParseInt(v, 10, 64)
if err == nil {
return time.Duration(i) * time.Millisecond
}
}
return def
}
