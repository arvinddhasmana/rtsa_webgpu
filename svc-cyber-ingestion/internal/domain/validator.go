// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"net"
"net/url"
"regexp"
"sync"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

var (
validIOCTypes = map[string]bool{
"ipv4-addr":   true,
"domain-name": true,
"file:hashes": true,
"url":         true,
}
sha256Re = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
domainRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z]{2,})+$`)
)

const dedupCacheSize = 1000

// Validator validates Cyber IOC observations with deduplication.
type Validator struct {
mu         sync.Mutex
dedupCache []string // ring buffer of recent dedup hashes
dedupIdx   int
dedupSet   map[string]bool
}

// NewValidator creates a new Cyber Validator.
func NewValidator() *Validator {
return &Validator{
dedupCache: make([]string, dedupCacheSize),
dedupSet:   make(map[string]bool),
}
}

// Validate checks Cyber-specific rules.
func (v *Validator) Validate(obs *ingestionv1.SensorObservation) ingestion.ValidationResult {
var errs []ingestion.ValidationError

if obs.GetSensorId() == "" {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "required", Message: "sensor_id is required"})
} else if len(obs.GetSensorId()) > 128 {
errs = append(errs, ingestion.ValidationError{Field: "sensor_id", Rule: "max_length", Message: "sensor_id exceeds 128 chars"})
}

if obs.GetSensorType() != commonv1.SensorType_SENSOR_TYPE_CYBER {
errs = append(errs, ingestion.ValidationError{Field: "sensor_type", Rule: "enum", Message: fmt.Sprintf("sensor_type must be SENSOR_TYPE_CYBER, got %s", obs.GetSensorType())})
}

if obs.GetObservationTime() == nil {
errs = append(errs, ingestion.ValidationError{Field: "observation_time", Rule: "required", Message: "observation_time is required"})
} else {
t := obs.GetObservationTime().AsTime()
now := time.Now()
if t.After(now.Add(5 * time.Minute)) {
errs = append(errs, ingestion.ValidationError{Field: "observation_time", Rule: "future", Message: "observation_time is more than 5 minutes in the future"})
}
if t.Before(now.Add(-24 * time.Hour)) {
errs = append(errs, ingestion.ValidationError{Field: "observation_time", Rule: "past", Message: "observation_time is more than 24 hours in the past"})
}
}

if obs.GetClassification() == commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNSPECIFIED {
errs = append(errs, ingestion.ValidationError{Field: "classification", Rule: "required", Message: "classification must not be UNSPECIFIED"})
}

cyber := obs.GetCyber()
if cyber == nil {
errs = append(errs, ingestion.ValidationError{Field: "cyber", Rule: "required", Message: "cyber payload is required"})
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

// STIX ID: must start with "indicator--"
stixID := cyber.GetStixId()
if len(stixID) < 11 || stixID[:11] != "indicator--" {
errs = append(errs, ingestion.ValidationError{Field: "cyber.stix_id", Rule: "format", Message: `cyber.stix_id must start with "indicator--"`})
}

// IOC type
iocType := cyber.GetIocType()
if !validIOCTypes[iocType] {
errs = append(errs, ingestion.ValidationError{Field: "cyber.ioc_type", Rule: "enum", Message: fmt.Sprintf("cyber.ioc_type must be one of ipv4-addr, domain-name, file:hashes, url; got %s", iocType)})
}

// IOC value
if cyber.GetIocValue() == "" {
errs = append(errs, ingestion.ValidationError{Field: "cyber.ioc_value", Rule: "required", Message: "cyber.ioc_value is required"})
} else {
switch iocType {
case "ipv4-addr":
ip := net.ParseIP(cyber.GetIocValue())
if ip == nil || ip.To4() == nil {
errs = append(errs, ingestion.ValidationError{Field: "cyber.ioc_value", Rule: "ipv4_format", Message: fmt.Sprintf("invalid IPv4 format: %s", cyber.GetIocValue())})
}
case "domain-name":
if !domainRe.MatchString(cyber.GetIocValue()) {
errs = append(errs, ingestion.ValidationError{Field: "cyber.ioc_value", Rule: "domain_format", Message: fmt.Sprintf("invalid domain format: %s", cyber.GetIocValue())})
}
case "url":
if _, parseErr := url.ParseRequestURI(cyber.GetIocValue()); parseErr != nil {
errs = append(errs, ingestion.ValidationError{Field: "cyber.ioc_value", Rule: "url_format", Message: fmt.Sprintf("invalid URL format: %s", cyber.GetIocValue())})
}
case "file:hashes":
if cyber.GetIocValue() == "" {
errs = append(errs, ingestion.ValidationError{Field: "cyber.ioc_value", Rule: "required", Message: "file hash value must not be empty"})
}
}
}

// Confidence
conf := cyber.GetConfidence()
if conf < 0.0 || conf > 1.0 {
errs = append(errs, ingestion.ValidationError{Field: "cyber.confidence", Rule: "range", Message: fmt.Sprintf("cyber.confidence out of range [0.0, 1.0]: %.2f", conf)})
}

// valid_from must not be in the future
if cyber.GetValidFrom() == nil {
errs = append(errs, ingestion.ValidationError{Field: "cyber.valid_from", Rule: "required", Message: "cyber.valid_from is required"})
} else if cyber.GetValidFrom().AsTime().After(time.Now()) {
errs = append(errs, ingestion.ValidationError{Field: "cyber.valid_from", Rule: "future", Message: "cyber.valid_from must not be in the future"})
}

// source_feed
if cyber.GetSourceFeed() == "" {
errs = append(errs, ingestion.ValidationError{Field: "cyber.source_feed", Rule: "required", Message: "cyber.source_feed is required"})
}

// dedup_hash: must be 64 hex chars (SHA-256)
if !sha256Re.MatchString(cyber.GetDedupHash()) {
errs = append(errs, ingestion.ValidationError{Field: "cyber.dedup_hash", Rule: "sha256_format", Message: "cyber.dedup_hash must be 64 hex characters (SHA-256)"})
}

if len(errs) > 0 {
return ingestion.ValidationResult{Valid: false, Errors: errs}
}

// Deduplication check (only if all other fields are valid)
if v.isDuplicate(cyber.GetDedupHash()) {
return ingestion.ValidationResult{Valid: false, Errors: []ingestion.ValidationError{
{Field: "cyber.dedup_hash", Rule: "duplicate", Message: "duplicate IOC observation"},
}}
}

return ingestion.ValidationResult{Valid: true}
}

// isDuplicate checks if a hash was recently seen, and adds it to the cache if not.
func (v *Validator) isDuplicate(hash string) bool {
v.mu.Lock()
defer v.mu.Unlock()

if v.dedupSet[hash] {
return true
}

// Evict oldest entry if slot is occupied
if old := v.dedupCache[v.dedupIdx]; old != "" {
delete(v.dedupSet, old)
}
v.dedupCache[v.dedupIdx] = hash
v.dedupSet[hash] = true
v.dedupIdx = (v.dedupIdx + 1) % dedupCacheSize
return false
}
