// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"testing"

	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"
	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"
	"github.com/redpanda-data/redpanda/src/transform-sdk/go/transform"
	"google.golang.org/protobuf/proto"
)

// ─── Test helpers ─────────────────────────────────────────────────────────────

// mockWriteEvent implements transform.WriteEvent for tests.
type mockWriteEvent struct {
	record transform.Record
}

func (m *mockWriteEvent) Record() transform.Record { return m.record }

// recordingWriter captures Write calls for assertion.
type recordingWriter struct {
	record   transform.Record
	routedTo string // empty means default output topic
}

func (r *recordingWriter) Write(record transform.Record, opts ...transform.WriteOpt) error {
	r.record = record
	if len(opts) > 0 {
		// Any non-empty opts means a non-default topic was requested (DLQ).
		r.routedTo = "dlq"
	}
	return nil
}

// hasHeader checks whether record carries a header with the given key and value.
func hasHeader(record transform.Record, key, value string) bool {
	for _, h := range record.Headers {
		if string(h.Key) == key && string(h.Value) == value {
			return true
		}
	}
	return false
}

// hasHeaderKey checks whether record carries a header with the given key.
func hasHeaderKey(record transform.Record, key string) bool {
	for _, h := range record.Headers {
		if string(h.Key) == key {
			return true
		}
	}
	return false
}

// validHeaders returns a complete set of required sensor message headers.
func validHeaders() []transform.RecordHeader {
	return []transform.RecordHeader{
		{Key: []byte("rtsa-classification"), Value: []byte("UNCLASSIFIED")},
		{Key: []byte("rtsa-source-service"), Value: []byte("svc-radar-ingestion")},
		{Key: []byte("rtsa-trace-id"), Value: []byte("trace-abc-123")},
		{Key: []byte("rtsa-timestamp"), Value: []byte("2026-02-24T07:15:00Z")},
		{Key: []byte("rtsa-schema-version"), Value: []byte("1")},
	}
}

// validSensorProto returns a marshalled SensorObservation with all required fields.
func validSensorProto(t *testing.T) []byte {
	t.Helper()
	obs := &ingestionv1.SensorObservation{
		ObservationId: "obs-001",
		SensorId:      "RADAR-HALIFAX-01",
		SensorType:    commonv1.SensorType_SENSOR_TYPE_RADAR,
	}
	b, err := proto.Marshal(obs)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	return b
}

// headersWithout returns a copy of validHeaders() with the named key removed.
func headersWithout(name string) []transform.RecordHeader {
	var out []transform.RecordHeader
	for _, h := range validHeaders() {
		if string(h.Key) != name {
			out = append(out, h)
		}
	}
	return out
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// T01/T08 — valid message passes through with rtsa-validated=true.
func TestValidateSensorMessage_ValidMessage_PassesThrough(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: validHeaders(),
		Value:   validSensorProto(t),
	}}

	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo != "" {
		t.Error("expected record to stay on default output, not be routed to DLQ")
	}
	if !hasHeader(w.record, "rtsa-validated", "true") {
		t.Error("expected rtsa-validated=true header")
	}
}

// T02 — missing rtsa-classification header → DLQ.
func TestValidateSensorMessage_MissingClassification_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: headersWithout("rtsa-classification"),
		Value:   validSensorProto(t),
	}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing")
	}
	if !hasHeaderKey(w.record, "rtsa-validation-error") {
		t.Error("expected rtsa-validation-error header")
	}
}

// T03 — missing rtsa-source-service header → DLQ.
func TestValidateSensorMessage_MissingSourceService_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: headersWithout("rtsa-source-service"),
		Value:   validSensorProto(t),
	}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing")
	}
}

// T04 — invalid classification value → DLQ.
func TestValidateSensorMessage_InvalidClassification_RoutedToDLQ(t *testing.T) {
	headers := validHeaders()
	for i, h := range headers {
		if string(h.Key) == "rtsa-classification" {
			headers[i].Value = []byte("TOP_SECRET")
		}
	}
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: headers, Value: validSensorProto(t)}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for invalid classification")
	}
}

// T05 — empty record value → DLQ.
func TestValidateSensorMessage_EmptyValue_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: validHeaders(), Value: []byte{}}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for empty value")
	}
}

// T06 — garbage bytes (invalid protobuf) → DLQ.
func TestValidateSensorMessage_InvalidProtobuf_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: validHeaders(),
		Value:   []byte{0xFF, 0xFE, 0x00, 0x01, 0x02, 0x03},
	}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for invalid protobuf")
	}
}

// T07 — valid protobuf but empty sensor_id → DLQ.
func TestValidateSensorMessage_EmptySensorID_RoutedToDLQ(t *testing.T) {
	obs := &ingestionv1.SensorObservation{
		SensorId:   "",
		SensorType: commonv1.SensorType_SENSOR_TYPE_RADAR,
	}
	b, _ := proto.Marshal(obs)
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: validHeaders(), Value: b}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for empty sensor_id")
	}
}

// T07b — valid protobuf but UNSPECIFIED sensor_type → DLQ.
func TestValidateSensorMessage_UnspecifiedSensorType_RoutedToDLQ(t *testing.T) {
	obs := &ingestionv1.SensorObservation{
		SensorId:   "RADAR-01",
		SensorType: commonv1.SensorType_SENSOR_TYPE_UNSPECIFIED,
	}
	b, _ := proto.Marshal(obs)
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: validHeaders(), Value: b}}
	if err := validateSensorMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for unspecified sensor_type")
	}
}

// AllHeaders — each required header individually missing causes DLQ routing.
func TestValidateSensorMessage_EachRequiredHeaderMissing_RoutedToDLQ(t *testing.T) {
	for _, missing := range requiredSensorHeaders {
		t.Run("missing_"+missing, func(t *testing.T) {
			w := &recordingWriter{}
			e := &mockWriteEvent{record: transform.Record{
				Headers: headersWithout(missing),
				Value:   validSensorProto(t),
			}}
			if err := validateSensorMessage(e, w); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if w.routedTo == "" {
				t.Errorf("expected DLQ routing when %q header is missing", missing)
			}
		})
	}
}
