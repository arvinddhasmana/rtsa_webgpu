// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-cyber-ingestion/internal/domain"
"google.golang.org/protobuf/types/known/timestamppb"
)

const validSHA256 = "a948904f2f0f479b8f936378f543e5a9b5a1c0a3e6b8f9c2d0e4f6a8c0e2d4f6"

func validCyberObs() *ingestionv1.SensorObservation {
return &ingestionv1.SensorObservation{
SensorId:        "CYBER-01",
SensorType:      commonv1.SensorType_SENSOR_TYPE_CYBER,
Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
ObservationTime: timestamppb.New(time.Now()),
SensorData: &ingestionv1.SensorObservation_Cyber{
Cyber: &ingestionv1.CyberIOC{
StixId:     "indicator--12345678-1234-1234-1234-123456789012",
IocType:    "ipv4-addr",
IocValue:   "192.168.1.1",
Confidence: 0.9,
ValidFrom:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
SourceFeed: "CTI-FEED-01",
DedupHash:  validSHA256,
},
},
}
}

func TestCyberValidator_T01_ValidIOC(t *testing.T) {
v := domain.NewValidator()
result := v.Validate(validCyberObs())
if !result.Valid {
t.Errorf("expected valid, got errors: %+v", result.Errors)
}
}

func TestCyberValidator_T02_InvalidSTIXID(t *testing.T) {
obs := validCyberObs()
obs.GetCyber().StixId = "bundle--12345"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid: STIX ID must start with indicator--")
}
}

func TestCyberValidator_T03_InvalidIOCType(t *testing.T) {
obs := validCyberObs()
obs.GetCyber().IocType = "email-addr"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid IOC type")
}
}

func TestCyberValidator_T04_InvalidIPv4(t *testing.T) {
obs := validCyberObs()
obs.GetCyber().IocValue = "999.999.999.999"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid IPv4")
}
}

func TestCyberValidator_T05_DuplicateDedupHash(t *testing.T) {
v := domain.NewValidator()
// First observation — should be accepted
obs1 := validCyberObs()
result1 := v.Validate(obs1)
if !result1.Valid {
t.Fatalf("first should be valid, got: %+v", result1.Errors)
}

// Second observation with same hash — should be rejected as duplicate
obs2 := validCyberObs()
obs2.GetCyber().DedupHash = validSHA256
result2 := v.Validate(obs2)
if result2.Valid {
t.Error("expected rejected as duplicate")
}
}

func TestCyberValidator_T06_InvalidSHA256Hash(t *testing.T) {
obs := validCyberObs()
obs.GetCyber().DedupHash = "not-a-sha256"
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid: bad SHA-256 hash")
}
}

func TestCyberValidator_T07_FutureValidFrom(t *testing.T) {
obs := validCyberObs()
obs.GetCyber().ValidFrom = timestamppb.New(time.Now().Add(1 * time.Hour))
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid: valid_from in future")
}
}

func TestCyberValidator_MissingCyberPayload(t *testing.T) {
obs := validCyberObs()
obs.SensorData = nil
v := domain.NewValidator()
result := v.Validate(obs)
if result.Valid {
t.Error("expected invalid when cyber payload is nil")
}
}
