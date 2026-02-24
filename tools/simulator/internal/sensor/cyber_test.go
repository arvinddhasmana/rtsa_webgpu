// CLASSIFICATION: UNCLASSIFIED
package sensor_test

import (
"math/rand"
"regexp"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/sensor"
)

var validIOCTypes = map[string]bool{
"ipv4-addr": true, "domain-name": true, "file:hashes": true, "url": true,
}

var dedupHashRe = regexp.MustCompile(`^[0-9a-f]{64}$`)
var stixIDRe = regexp.MustCompile(`^indicator--`)

func TestGenerateCyberObservation_ValidFields(t *testing.T) {
r := rand.New(rand.NewSource(42))
obs := sensor.GenerateCyberObservation(r)

if obs == nil {
t.Fatal("cyber observation should not be nil")
}
if obs.SensorType != commonv1.SensorType_SENSOR_TYPE_CYBER {
t.Errorf("expected SENSOR_TYPE_CYBER, got %v", obs.SensorType)
}
if obs.Classification != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
t.Errorf("must be UNCLASSIFIED, got %v", obs.Classification)
}

cyberPl, ok := obs.SensorData.(*ingestionv1.SensorObservation_Cyber)
if !ok {
t.Fatal("sensor_data should be CyberIOC")
}
ioc := cyberPl.Cyber

// STIX ID must start with "indicator--".
if !stixIDRe.MatchString(ioc.StixId) {
t.Errorf("stix_id must start with 'indicator--', got %q", ioc.StixId)
}

// IOC type must be valid.
if !validIOCTypes[ioc.IocType] {
t.Errorf("invalid ioc_type %q", ioc.IocType)
}

// IOC value must be non-empty.
if ioc.IocValue == "" {
t.Error("ioc_value must be non-empty")
}

// Confidence must be 0.0-1.0.
if ioc.Confidence < 0 || ioc.Confidence > 1 {
t.Errorf("confidence must be 0-1, got %f", ioc.Confidence)
}

// valid_from must not be nil.
if ioc.ValidFrom == nil {
t.Error("valid_from must be set")
}

// source_feed must be non-empty.
if ioc.SourceFeed == "" {
t.Error("source_feed must be non-empty")
}

// dedup_hash must be 64 hex chars (SHA-256).
if !dedupHashRe.MatchString(ioc.DedupHash) {
t.Errorf("dedup_hash must be 64 hex chars, got %q (len=%d)", ioc.DedupHash, len(ioc.DedupHash))
}

// ATT&CK technique IDs must be non-empty list.
if len(ioc.MitreAttackIds) == 0 {
t.Error("mitre_attack_ids must have at least one entry")
}
}

func TestGenerateCyberObservation_UniqueDedupHash(t *testing.T) {
r := rand.New(rand.NewSource(42))
obs1 := sensor.GenerateCyberObservation(r)
obs2 := sensor.GenerateCyberObservation(r)

h1 := obs1.SensorData.(*ingestionv1.SensorObservation_Cyber).Cyber.DedupHash
h2 := obs2.SensorData.(*ingestionv1.SensorObservation_Cyber).Cyber.DedupHash

if h1 == h2 {
t.Errorf("dedup hashes should differ between observations: both %q", h1)
}
}

func TestGenerateCyberObservation_ValidFromNotFuture(t *testing.T) {
r := rand.New(rand.NewSource(42))
for i := 0; i < 20; i++ {
obs := sensor.GenerateCyberObservation(r)
ioc := obs.SensorData.(*ingestionv1.SensorObservation_Cyber).Cyber
if ioc.ValidFrom.AsTime().After(obs.ObservationTime.AsTime()) {
t.Errorf("valid_from (%v) must not be after observation_time (%v)",
ioc.ValidFrom.AsTime(), obs.ObservationTime.AsTime())
}
}
}

func TestGenerateCyberObservation_SensorID(t *testing.T) {
r := rand.New(rand.NewSource(42))
obs := sensor.GenerateCyberObservation(r)
if obs.SensorId != "CYBER-SIM-001" {
t.Errorf("expected sensor_id CYBER-SIM-001, got %q", obs.SensorId)
}
}
