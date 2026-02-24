// CLASSIFICATION: UNCLASSIFIED

package repository

import (
"testing"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"google.golang.org/protobuf/types/known/timestamppb"
)

// TestParseEntityType verifies all entity type string <-> proto mappings.
func TestParseEntityType(t *testing.T) {
tests := []struct {
input string
want  commonv1.EntityType
}{
{"SURFACE", commonv1.EntityType_ENTITY_TYPE_SURFACE},
{"AIR", commonv1.EntityType_ENTITY_TYPE_AIR},
{"SUBSURFACE", commonv1.EntityType_ENTITY_TYPE_SUBSURFACE},
{"LAND", commonv1.EntityType_ENTITY_TYPE_LAND},
{"CYBER", commonv1.EntityType_ENTITY_TYPE_CYBER},
{"UNKNOWN_TYPE", commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED},
{"", commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED},
}
for _, tc := range tests {
t.Run(tc.input, func(t *testing.T) {
got := parseEntityType(tc.input)
if got != tc.want {
t.Errorf("parseEntityType(%q) = %v, want %v", tc.input, got, tc.want)
}
})
}
}

// TestParseHostileClass verifies hostile classification mappings.
func TestParseHostileClass(t *testing.T) {
tests := []struct {
input string
want  commonv1.HostileClassification
}{
{"HOSTILE", commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE},
{"FRIENDLY", commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY},
{"NEUTRAL", commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL},
{"UNKNOWN", commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN},
{"BAD", commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNSPECIFIED},
}
for _, tc := range tests {
t.Run(tc.input, func(t *testing.T) {
got := parseHostileClass(tc.input)
if got != tc.want {
t.Errorf("parseHostileClass(%q) = %v, want %v", tc.input, got, tc.want)
}
})
}
}

// TestParseClassificationLevel verifies classification level mappings.
func TestParseClassificationLevel(t *testing.T) {
tests := []struct {
input string
want  commonv1.ClassificationLevel
}{
{"UNCLASSIFIED", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
{"PROTECTED_A", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_A},
{"PROTECTED_B", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_B},
{"PROTECTED_C", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_PROTECTED_C},
{"SECRET", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET},
{"TOP_SECRET", commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED},
}
for _, tc := range tests {
t.Run(tc.input, func(t *testing.T) {
got := parseClassificationLevel(tc.input)
if got != tc.want {
t.Errorf("parseClassificationLevel(%q) = %v, want %v", tc.input, got, tc.want)
}
})
}
}

// TestParseTrackStatus verifies track status string mappings.
func TestParseTrackStatus(t *testing.T) {
tests := []struct {
input string
want  commonv1.TrackStatus
}{
{"ACTIVE", commonv1.TrackStatus_TRACK_STATUS_ACTIVE},
{"STALE", commonv1.TrackStatus_TRACK_STATUS_STALE},
{"DROPPED", commonv1.TrackStatus_TRACK_STATUS_DROPPED},
{"MERGED", commonv1.TrackStatus_TRACK_STATUS_MERGED},
{"", commonv1.TrackStatus_TRACK_STATUS_ACTIVE},
}
for _, tc := range tests {
t.Run(tc.input, func(t *testing.T) {
got := parseTrackStatus(tc.input)
if got != tc.want {
t.Errorf("parseTrackStatus(%q) = %v, want %v", tc.input, got, tc.want)
}
})
}
}

// TestEntityTypeToString verifies enum -> string conversions for SQL parameters.
func TestEntityTypeToString(t *testing.T) {
tests := []struct {
input commonv1.EntityType
want  string
}{
{commonv1.EntityType_ENTITY_TYPE_SURFACE, "SURFACE"},
{commonv1.EntityType_ENTITY_TYPE_AIR, "AIR"},
{commonv1.EntityType_ENTITY_TYPE_SUBSURFACE, "SUBSURFACE"},
{commonv1.EntityType_ENTITY_TYPE_LAND, "LAND"},
{commonv1.EntityType_ENTITY_TYPE_CYBER, "CYBER"},
{commonv1.EntityType_ENTITY_TYPE_UNSPECIFIED, "UNSPECIFIED"},
}
for _, tc := range tests {
got := entityTypeToString(tc.input)
if got != tc.want {
t.Errorf("entityTypeToString(%v) = %q, want %q", tc.input, got, tc.want)
}
}
}

// TestHostileClassToString verifies hostile class -> string conversions.
func TestHostileClassToString(t *testing.T) {
tests := []struct {
input commonv1.HostileClassification
want  string
}{
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_HOSTILE, "HOSTILE"},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_FRIENDLY, "FRIENDLY"},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_NEUTRAL, "NEUTRAL"},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNKNOWN, "UNKNOWN"},
{commonv1.HostileClassification_HOSTILE_CLASSIFICATION_UNSPECIFIED, "UNSPECIFIED"},
}
for _, tc := range tests {
got := hostileClassToString(tc.input)
if got != tc.want {
t.Errorf("hostileClassToString(%v) = %q, want %q", tc.input, got, tc.want)
}
}
}

// TestAnomalyTypeToString verifies anomaly type -> string conversions.
func TestAnomalyTypeToString(t *testing.T) {
tests := []struct {
input commonv1.AnomalyType
want  string
}{
{commonv1.AnomalyType_ANOMALY_TYPE_SPEED, "SPEED"},
{commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION, "ROUTE_DEVIATION"},
{commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION, "AIS_MANIPULATION"},
{commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL, "BEHAVIORAL"},
{commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL, "TEMPORAL"},
{commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY, "PROXIMITY"},
{commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED, "UNSPECIFIED"},
}
for _, tc := range tests {
got := anomalyTypeToString(tc.input)
if got != tc.want {
t.Errorf("anomalyTypeToString(%v) = %q, want %q", tc.input, got, tc.want)
}
}
}

// TestSeverityToString verifies alert severity -> string conversions.
func TestSeverityToString(t *testing.T) {
tests := []struct {
input commonv1.AlertSeverity
want  string
}{
{commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL, "NORMAL"},
{commonv1.AlertSeverity_ALERT_SEVERITY_WATCH, "WATCH"},
{commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, "ELEVATED"},
{commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL, "CRITICAL"},
{commonv1.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED, "NORMAL"},
}
for _, tc := range tests {
got := severityToString(tc.input)
if got != tc.want {
t.Errorf("severityToString(%v) = %q, want %q", tc.input, got, tc.want)
}
}
}

// TestParseAnomalyType verifies string -> anomaly type proto conversions.
func TestParseAnomalyType(t *testing.T) {
tests := []struct {
input string
want  commonv1.AnomalyType
}{
{"SPEED", commonv1.AnomalyType_ANOMALY_TYPE_SPEED},
{"ROUTE_DEVIATION", commonv1.AnomalyType_ANOMALY_TYPE_ROUTE_DEVIATION},
{"AIS_MANIPULATION", commonv1.AnomalyType_ANOMALY_TYPE_AIS_MANIPULATION},
{"BEHAVIORAL", commonv1.AnomalyType_ANOMALY_TYPE_BEHAVIORAL},
{"TEMPORAL", commonv1.AnomalyType_ANOMALY_TYPE_TEMPORAL},
{"PROXIMITY", commonv1.AnomalyType_ANOMALY_TYPE_PROXIMITY},
{"UNKNOWN", commonv1.AnomalyType_ANOMALY_TYPE_UNSPECIFIED},
}
for _, tc := range tests {
got := parseAnomalyType(tc.input)
if got != tc.want {
t.Errorf("parseAnomalyType(%q) = %v, want %v", tc.input, got, tc.want)
}
}
}

// TestParseSeverity verifies string -> alert severity proto conversions.
func TestParseSeverity(t *testing.T) {
tests := []struct {
input string
want  commonv1.AlertSeverity
}{
{"NORMAL", commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL},
{"WATCH", commonv1.AlertSeverity_ALERT_SEVERITY_WATCH},
{"ELEVATED", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED},
{"CRITICAL", commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL},
{"BAD", commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL},
}
for _, tc := range tests {
got := parseSeverity(tc.input)
if got != tc.want {
t.Errorf("parseSeverity(%q) = %v, want %v", tc.input, got, tc.want)
}
}
}

// TestTrackRowToProto verifies the ClickHouse row -> FusedTrack proto conversion.
func TestTrackRowToProto_nilNullables(t *testing.T) {
eventTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
row := &trackRow{
TrackID:               "track-001",
EntityType:            "SURFACE",
HostileClassification: "HOSTILE",
Latitude:              45.0,
Longitude:             -75.0,
AltitudeMeters:        nil,
SpeedKnots:            nil,
HeadingDegrees:        nil,
ConfidenceScore:       0.92,
SourceCount:           2,
SourceSensors:         []string{"radar-01", "ew-01"},
ClassificationLevel:   "UNCLASSIFIED",
TrackStatus:           "ACTIVE",
EventTime:             eventTime,
}

proto := trackRowToProto(row)

if proto.TrackId != "track-001" {
t.Errorf("TrackId = %q, want %q", proto.TrackId, "track-001")
}
if proto.EstimatedPosition == nil {
t.Fatal("EstimatedPosition is nil")
}
if proto.EstimatedPosition.Latitude != 45.0 {
t.Errorf("Latitude = %f, want 45.0", proto.EstimatedPosition.Latitude)
}
if proto.EstimatedPosition.AltitudeMeters != nil {
t.Errorf("AltitudeMeters should be nil, got %v", proto.EstimatedPosition.AltitudeMeters)
}
if len(proto.Sources) != 2 {
t.Errorf("Sources len = %d, want 2", len(proto.Sources))
}
if proto.Sources[0].SensorId != "radar-01" {
t.Errorf("Sources[0].SensorId = %q, want %q", proto.Sources[0].SensorId, "radar-01")
}
if proto.UpdatedAt == nil {
t.Fatal("UpdatedAt is nil")
}
if !proto.UpdatedAt.AsTime().Equal(eventTime) {
t.Errorf("UpdatedAt = %v, want %v", proto.UpdatedAt.AsTime(), eventTime)
}
}

func TestTrackRowToProto_withNullables(t *testing.T) {
alt := 10000.0
speed := 250.5
heading := 090.0
row := &trackRow{
TrackID:               "track-002",
EntityType:            "AIR",
HostileClassification: "UNKNOWN",
Latitude:              48.0,
Longitude:             -73.0,
AltitudeMeters:        &alt,
SpeedKnots:            &speed,
HeadingDegrees:        &heading,
ConfidenceScore:       0.75,
SourceCount:           1,
SourceSensors:         []string{"radar-02"},
ClassificationLevel:   "SECRET",
TrackStatus:           "STALE",
EventTime:             time.Now(),
}

proto := trackRowToProto(row)

if proto.EstimatedPosition.AltitudeMeters == nil {
t.Fatal("AltitudeMeters should not be nil")
}
if *proto.EstimatedPosition.AltitudeMeters != alt {
t.Errorf("AltitudeMeters = %f, want %f", *proto.EstimatedPosition.AltitudeMeters, alt)
}
if *proto.EstimatedPosition.SpeedKnots != speed {
t.Errorf("SpeedKnots = %f, want %f", *proto.EstimatedPosition.SpeedKnots, speed)
}
}

// TestAnomalyRowToProto verifies anomaly row -> AnomalyAlert proto conversion.
func TestAnomalyRowToProto(t *testing.T) {
eventTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
row := &anomalyRow{
AlertID:             "alert-001",
TrackID:             "track-001",
AnomalyType:         "SPEED",
Severity:            "CRITICAL",
ConfidenceScore:     0.95,
Explanation:         "Speed anomaly detected",
ModelVersion:        "v1.0.0",
ClassificationLevel: "PROTECTED_B",
EventTime:           eventTime,
}

proto := anomalyRowToProto(row)

if proto.AlertId != "alert-001" {
t.Errorf("AlertId = %q, want %q", proto.AlertId, "alert-001")
}
if proto.ConfidenceScore != 0.95 {
t.Errorf("ConfidenceScore = %f, want 0.95", proto.ConfidenceScore)
}
if proto.DetectedAt == nil {
t.Fatal("DetectedAt is nil")
}
if !proto.DetectedAt.AsTime().Equal(eventTime) {
t.Errorf("DetectedAt = %v, want %v", proto.DetectedAt.AsTime(), eventTime)
}
}

// TestAuditRowToProto verifies audit row -> AuditLogEntry proto conversion.
func TestAuditRowToProto(t *testing.T) {
eventTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
row := &auditRow{
AuditID:             "audit-001",
ServiceID:           "svc-query",
EventType:           "query_executed",
ActorID:             "op-001",
ActorType:           "OPERATOR",
ResourceType:        "tracks",
ResourceID:          "",
Action:              "QUERY",
DetailJSON:          `{"query_type":"QueryTracks"}`,
ClassificationLevel: "UNCLASSIFIED",
EventTime:           eventTime,
}

proto := auditRowToProto(row)

if proto.AuditId != "audit-001" {
t.Errorf("AuditId = %q, want %q", proto.AuditId, "audit-001")
}
if proto.ServiceId != "svc-query" {
t.Errorf("ServiceId = %q, want %q", proto.ServiceId, "svc-query")
}
if proto.EventTime == "" {
t.Error("EventTime is empty")
}
if proto.DetailJson != `{"query_type":"QueryTracks"}` {
t.Errorf("DetailJson = %q, unexpected", proto.DetailJson)
}
}

// TestParseDSN verifies DSN parsing for ClickHouse connection.
func TestParseDSN(t *testing.T) {
tests := []struct {
dsn      string
wantAddr string
wantUser string
wantDB   string
wantErr  bool
}{
{
dsn:      "clickhouse://default:secret@localhost:9000/rtsa",
wantAddr: "localhost:9000",
wantUser: "default",
wantDB:   "rtsa",
},
{
dsn:      "clickhouse://localhost/mydb",
wantAddr: "localhost:9000",
wantDB:   "mydb",
},
{
dsn:     ":::invalid:::",
wantErr: true,
},
}
for _, tc := range tests {
t.Run(tc.dsn, func(t *testing.T) {
addr, auth, err := parseDSN(tc.dsn)
if tc.wantErr {
if err == nil {
t.Fatal("expected error, got nil")
}
return
}
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if addr != tc.wantAddr {
t.Errorf("addr = %q, want %q", addr, tc.wantAddr)
}
if tc.wantUser != "" && auth.Username != tc.wantUser {
t.Errorf("user = %q, want %q", auth.Username, tc.wantUser)
}
if auth.Database != tc.wantDB {
t.Errorf("db = %q, want %q", auth.Database, tc.wantDB)
}
})
}
}

// TestNewTracksRepository verifies construction doesn't panic.
func TestNewTracksRepository(t *testing.T) {
repo := NewTracksRepository(nil, nil, nil)
if repo == nil {
t.Fatal("expected non-nil repository")
}
}

// TestNewAnomalyRepository verifies construction doesn't panic.
func TestNewAnomalyRepository(t *testing.T) {
repo := NewAnomalyRepository(nil, nil, nil)
if repo == nil {
t.Fatal("expected non-nil repository")
}
}

// TestNewAuditRepository verifies construction doesn't panic.
func TestNewAuditRepository(t *testing.T) {
repo := NewAuditRepository(nil, nil, nil)
if repo == nil {
t.Fatal("expected non-nil repository")
}
}

// Compile-time interface conformance checks
var (
_ TracksQuerier   = (*TracksRepository)(nil)
_ AnomaliesQuerier = (*AnomalyRepository)(nil)
_ AuditQuerier    = (*AuditRepository)(nil)
)

// Silence unused import
var _ = timestamppb.Now
