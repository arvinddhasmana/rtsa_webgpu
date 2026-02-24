// CLASSIFICATION: UNCLASSIFIED
package scenario

import (
"fmt"
"os"

"gopkg.in/yaml.v3"
)

// Scenario defines a complete simulation configuration loaded from YAML.
type Scenario struct {
Name        string      `yaml:"name"`
Description string      `yaml:"description"`
Seed        int64       `yaml:"seed"`
DurationMin int         `yaml:"duration_minutes"`
Entities    EntityConfig `yaml:"entities"`
Sensors     SensorConfig `yaml:"sensors"`
Anomalies   AnomalyConfig `yaml:"anomalies"`
OperationalArea OperationalArea `yaml:"operational_area"`
}

// EntityConfig groups entity configuration by domain.
type EntityConfig struct {
Surface    DomainEntityConfig `yaml:"surface"`
Air        DomainEntityConfig `yaml:"air"`
Subsurface DomainEntityConfig `yaml:"subsurface"`
}

// DomainEntityConfig defines entity generation parameters for one domain.
type DomainEntityConfig struct {
Count         int                `yaml:"count"`
HostileRatio  float64            `yaml:"hostile_ratio"`
FriendlyRatio float64            `yaml:"friendly_ratio"`
NeutralRatio  float64            `yaml:"neutral_ratio"`
UnknownRatio  float64            `yaml:"unknown_ratio"`
SpeedRange    [2]float64         `yaml:"speed_range"`
AltitudeRange [2]float64         `yaml:"altitude_range"`
DepthRange    [2]float64         `yaml:"depth_range"`
Patterns      map[string]float64 `yaml:"patterns"`
}

// SensorConfig groups all sensor-type configurations.
type SensorConfig struct {
Radar  SensorTypeConfig `yaml:"radar"`
AIS    SensorTypeConfig `yaml:"ais"`
EW     SensorTypeConfig `yaml:"ew"`
ELINT  SensorTypeConfig `yaml:"elint"`
ISR    SensorTypeConfig `yaml:"isr"`
Cyber  CyberSensorConfig `yaml:"cyber"`
}

// SensorTypeConfig defines basic sensor scheduling parameters.
type SensorTypeConfig struct {
Count            int      `yaml:"count"`
SensorIDs        []string `yaml:"sensor_ids"`
UpdateIntervalMs int      `yaml:"update_interval_ms"`
CoverageNM       float64  `yaml:"coverage_nm"`
}

// CyberSensorConfig extends SensorTypeConfig with IOC rate.
type CyberSensorConfig struct {
SensorTypeConfig `yaml:",inline"`
IOCRate          int `yaml:"ioc_rate"`
}

// AnomalyConfig defines anomaly injection configuration.
type AnomalyConfig struct {
InjectionRate float64            `yaml:"injection_rate"`
Types         map[string]float64 `yaml:"types"`
}

// OperationalArea defines the geographic bounds of the simulation.
type OperationalArea struct {
Center         GeoPoint        `yaml:"center"`
RadiusNM       float64         `yaml:"radius_nm"`
ExclusionZones []ExclusionZone `yaml:"exclusion_zones"`
}

// GeoPoint is a lat/lon coordinate.
type GeoPoint struct {
Lat float64 `yaml:"lat"`
Lon float64 `yaml:"lon"`
}

// ExclusionZone defines a named circular exclusion area.
type ExclusionZone struct {
Name     string  `yaml:"name"`
Center   GeoPoint `yaml:"center"`
RadiusNM float64  `yaml:"radius_nm"`
}

// LoadFromFile reads and parses a YAML scenario file.
func LoadFromFile(path string) (*Scenario, error) {
data, err := os.ReadFile(path)
if err != nil {
return nil, fmt.Errorf("reading scenario file %q: %w", path, err)
}
return Parse(data)
}

// Parse decodes YAML bytes into a Scenario.
func Parse(data []byte) (*Scenario, error) {
var s Scenario
if err := yaml.Unmarshal(data, &s); err != nil {
return nil, fmt.Errorf("parsing scenario YAML: %w", err)
}
if err := s.Validate(); err != nil {
return nil, fmt.Errorf("invalid scenario: %w", err)
}
return &s, nil
}

// Validate checks that the scenario configuration is coherent.
func (s *Scenario) Validate() error {
if s.Name == "" {
return fmt.Errorf("scenario name is required")
}
if s.Entities.Surface.Count < 0 || s.Entities.Air.Count < 0 || s.Entities.Subsurface.Count < 0 {
return fmt.Errorf("entity counts must be non-negative")
}
if s.Anomalies.InjectionRate < 0 || s.Anomalies.InjectionRate > 1 {
return fmt.Errorf("anomaly injection_rate must be 0.0-1.0")
}
return nil
}
