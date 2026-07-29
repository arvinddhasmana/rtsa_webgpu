// CLASSIFICATION: UNCLASSIFIED
package scenario_test

import (
"os"
"path/filepath"
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/tools/simulator/internal/scenario"
)

const validYAML = `# CLASSIFICATION: UNCLASSIFIED
name: "Test Scenario"
description: "Test"
seed: 42
duration_minutes: 10
entities:
  surface:
    count: 5
  air:
    count: 3
  subsurface:
    count: 1
anomalies:
  injection_rate: 0.05
`

func TestParse_ValidScenario(t *testing.T) {
sc, err := scenario.Parse([]byte(validYAML))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if sc.Name != "Test Scenario" {
t.Errorf("expected name 'Test Scenario', got %q", sc.Name)
}
if sc.Seed != 42 {
t.Errorf("expected seed 42, got %d", sc.Seed)
}
if sc.Entities.Surface.Count != 5 {
t.Errorf("expected 5 surface entities, got %d", sc.Entities.Surface.Count)
}
if sc.Anomalies.InjectionRate != 0.05 {
t.Errorf("expected injection rate 0.05, got %f", sc.Anomalies.InjectionRate)
}
}

func TestParse_MissingName(t *testing.T) {
yaml := `description: "No name"
entities:
  surface:
    count: 1
anomalies:
  injection_rate: 0.0
`
_, err := scenario.Parse([]byte(yaml))
if err == nil {
t.Error("expected error for missing scenario name")
}
}

func TestParse_InvalidAnomalyRate(t *testing.T) {
yaml := `name: "Bad Rate"
entities:
  surface:
    count: 1
anomalies:
  injection_rate: 1.5
`
_, err := scenario.Parse([]byte(yaml))
if err == nil {
t.Error("expected error for anomaly rate > 1.0")
}
}

func TestParse_NegativeEntityCount(t *testing.T) {
yaml := `name: "Negative"
entities:
  surface:
    count: -1
anomalies:
  injection_rate: 0.0
`
_, err := scenario.Parse([]byte(yaml))
if err == nil {
t.Error("expected error for negative entity count")
}
}

func TestLoadFromFile_ValidFile(t *testing.T) {
// Write a temp scenario file.
dir := t.TempDir()
path := filepath.Join(dir, "test.yaml")
if err := os.WriteFile(path, []byte(validYAML), 0600); err != nil {
t.Fatalf("failed to write temp file: %v", err)
}

sc, err := scenario.LoadFromFile(path)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if sc.Name != "Test Scenario" {
t.Errorf("expected 'Test Scenario', got %q", sc.Name)
}
}

func TestLoadFromFile_MissingFile(t *testing.T) {
_, err := scenario.LoadFromFile("/nonexistent/path/scenario.yaml")
if err == nil {
t.Error("expected error for non-existent file")
}
}

func TestLoadFromFile_DefaultScenario(t *testing.T) {
// Test loading the actual default scenario file from the repo.
sc, err := scenario.LoadFromFile("../../scenarios/default.yaml")
if err != nil {
t.Fatalf("failed to load default scenario: %v", err)
}
if sc.Name == "" {
t.Error("default scenario name should not be empty")
}
if sc.Entities.Surface.Count < 1 {
t.Error("default scenario should have surface entities")
}
}

func TestLoadFromFile_StressScenario(t *testing.T) {
sc, err := scenario.LoadFromFile("../../scenarios/stress.yaml")
if err != nil {
t.Fatalf("failed to load stress scenario: %v", err)
}
if sc.Entities.Surface.Count < 100 {
t.Errorf("stress scenario should have ≥100 surface entities, got %d", sc.Entities.Surface.Count)
}
}

func TestLoadFromFile_AnomalyDemoScenario(t *testing.T) {
sc, err := scenario.LoadFromFile("../../scenarios/anomaly-demo.yaml")
if err != nil {
t.Fatalf("failed to load anomaly-demo scenario: %v", err)
}
if sc.Anomalies.InjectionRate <= 0.1 {
t.Errorf("anomaly-demo should have high injection rate, got %f", sc.Anomalies.InjectionRate)
}
}

func TestScenario_Validate_Valid(t *testing.T) {
sc := &scenario.Scenario{
Name: "Valid",
Entities: scenario.EntityConfig{
Surface:    scenario.DomainEntityConfig{Count: 10},
Air:        scenario.DomainEntityConfig{Count: 5},
Subsurface: scenario.DomainEntityConfig{Count: 2},
},
Anomalies: scenario.AnomalyConfig{InjectionRate: 0.1},
}
if err := sc.Validate(); err != nil {
t.Errorf("valid scenario should not fail validation: %v", err)
}
}
