// CLASSIFICATION: UNCLASSIFIED
package sensor

import (
"crypto/sha256"
"fmt"
"math/rand"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/google/uuid"
"google.golang.org/protobuf/types/known/timestamppb"
)

// cyberIOCTypes lists the valid IOC types accepted by the Cyber validator.
var cyberIOCTypes = []string{"ipv4-addr", "domain-name", "file:hashes", "url"}

// cyberSourceFeeds lists synthetic threat-intel feed names.
var cyberSourceFeeds = []string{
"SIM-THREATFEED-ALPHA",
"SIM-THREATFEED-BRAVO",
"SIM-OSINT-01",
"SIM-DARKWEB-MONITOR",
}

// mitreAttackTechniques lists sample ATT&CK technique IDs.
var mitreAttackTechniques = []string{
"T1190", "T1133", "T1078", "T1059", "T1486",
"T1071", "T1105", "T1082", "T1083", "T1021",
"T1566", "T1203", "T1055", "T1027", "T1036",
}

// cyberMalwareFamilies provides synthetic malware family names.
var cyberMalwareFamilies = []string{
"SimLocker", "TestRat", "DemoStager", "SyntheticC2", "MockWiper",
}

// GenerateCyberObservation creates a SensorObservation with a CyberIOC payload.
// Not tied to any specific entity — generates standalone IOC indicators.
// All generated observations are CLASSIFICATION_LEVEL_UNCLASSIFIED.
func GenerateCyberObservation(r *rand.Rand) *ingestionv1.SensorObservation {
iocType := cyberIOCTypes[r.Intn(len(cyberIOCTypes))]
iocValue := generateIOCValue(r, iocType)
stixID := fmt.Sprintf("indicator--%s", uuid.New().String())
sourceFeed := cyberSourceFeeds[r.Intn(len(cyberSourceFeeds))]
confidence := 0.3 + r.Float64()*0.7 // 0.3-1.0

// Select 1-3 ATT&CK technique IDs.
numTechniques := 1 + r.Intn(3)
techniques := make([]string, 0, numTechniques)
used := make(map[int]bool)
for len(techniques) < numTechniques {
idx := r.Intn(len(mitreAttackTechniques))
if !used[idx] {
used[idx] = true
techniques = append(techniques, mitreAttackTechniques[idx])
}
}

// Malware family (optional — 50% chance).
var malwareFamily *string
if r.Intn(2) == 0 {
mf := cyberMalwareFamilies[r.Intn(len(cyberMalwareFamilies))]
malwareFamily = &mf
}

// valid_from: between 90 days ago and now.
validFrom := time.Now().Add(-time.Duration(r.Intn(90*24)) * time.Hour)

// valid_until: 30-180 days after valid_from.
validUntilTime := validFrom.Add(time.Duration(30+r.Intn(150)) * 24 * time.Hour)

// dedup_hash: SHA-256 of IOC type+value+timestamp (64 hex chars).
hashInput := fmt.Sprintf("%s:%s:%d", iocType, iocValue, time.Now().UnixNano())
hashBytes := sha256.Sum256([]byte(hashInput))
dedupHash := fmt.Sprintf("%x", hashBytes[:]) // 64 hex chars

obs := &ingestionv1.SensorObservation{
ObservationId:   uuid.New().String(),
SensorId:        "CYBER-SIM-001",
SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
ObservationTime: timestamppb.New(time.Now()),
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
Metadata: map[string]string{
"sim_generated": "true",
},
SensorData: &ingestionv1.SensorObservation_Cyber{
Cyber: &ingestionv1.CyberIOC{
StixId:         stixID,
IocType:        iocType,
IocValue:       iocValue,
MitreAttackIds: techniques,
MalwareFamily:  malwareFamily,
Confidence:     confidence,
ValidFrom:      timestamppb.New(validFrom),
ValidUntil:     timestamppb.New(validUntilTime),
SourceFeed:     sourceFeed,
DedupHash:      dedupHash,
},
},
}
return obs
}

// generateIOCValue produces a synthetic IOC value appropriate for the type.
func generateIOCValue(r *rand.Rand, iocType string) string {
switch iocType {
case "ipv4-addr":
// Use TEST-NET ranges (RFC 5737) to avoid real IP concerns.
return fmt.Sprintf("192.0.2.%d", r.Intn(254)+1)
case "domain-name":
tlds := []string{"test", "example", "invalid"}
labels := []string{"sim", "demo", "test", "synthetic", "rtsa"}
return fmt.Sprintf("%s-%d.%s", labels[r.Intn(len(labels))], r.Intn(9999), tlds[r.Intn(len(tlds))])
case "file:hashes":
// Generate a synthetic SHA-256 hash.
data := fmt.Sprintf("synthetic-file-%d", r.Int63())
h := sha256.Sum256([]byte(data))
return fmt.Sprintf("%x", h[:])
case "url":
return fmt.Sprintf("http://sim-%d.example.test/path/%d", r.Intn(9999), r.Intn(9999))
default:
return fmt.Sprintf("sim-ioc-%d", r.Intn(99999))
}
}
