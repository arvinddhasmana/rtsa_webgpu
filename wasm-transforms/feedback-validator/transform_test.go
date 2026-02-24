// CLASSIFICATION: UNCLASSIFIED
package main

import (
	"testing"

	feedbackv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/feedback/v1"
	"github.com/redpanda-data/redpanda/src/transform-sdk/go/transform"
	"google.golang.org/protobuf/proto"
)

// ─── Test helpers ─────────────────────────────────────────────────────────────

type mockWriteEvent struct {
	record transform.Record
}

func (m *mockWriteEvent) Record() transform.Record { return m.record }

type recordingWriter struct {
	record   transform.Record
	routedTo string
}

func (r *recordingWriter) Write(record transform.Record, opts ...transform.WriteOpt) error {
	r.record = record
	if len(opts) > 0 {
		r.routedTo = "dlq"
	}
	return nil
}

func hasHeader(record transform.Record, key, value string) bool {
	for _, h := range record.Headers {
		if string(h.Key) == key && string(h.Value) == value {
			return true
		}
	}
	return false
}

func hasHeaderKey(record transform.Record, key string) bool {
	for _, h := range record.Headers {
		if string(h.Key) == key {
			return true
		}
	}
	return false
}

func validHeaders() []transform.RecordHeader {
	return []transform.RecordHeader{
		{Key: []byte("rtsa-classification"), Value: []byte("UNCLASSIFIED")},
		{Key: []byte("rtsa-source-service"), Value: []byte("svc-feedback")},
		{Key: []byte("rtsa-trace-id"), Value: []byte("trace-def-789")},
		{Key: []byte("rtsa-timestamp"), Value: []byte("2026-02-24T07:15:00Z")},
		{Key: []byte("rtsa-schema-version"), Value: []byte("1")},
	}
}

func validFeedbackProto(t *testing.T) []byte {
	t.Helper()
	fb := &feedbackv1.OperatorFeedback{FeedbackId: "FB-001"}
	b, err := proto.Marshal(fb)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	return b
}

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

func TestValidateFeedbackMessage_ValidMessage_PassesThrough(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: validHeaders(),
		Value:   validFeedbackProto(t),
	}}
	if err := validateFeedbackMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo != "" {
		t.Error("expected default output, not DLQ")
	}
	if !hasHeader(w.record, "rtsa-validated", "true") {
		t.Error("expected rtsa-validated=true header")
	}
}

func TestValidateFeedbackMessage_MissingClassification_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: headersWithout("rtsa-classification"),
		Value:   validFeedbackProto(t),
	}}
	if err := validateFeedbackMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing")
	}
	if !hasHeaderKey(w.record, "rtsa-validation-error") {
		t.Error("expected rtsa-validation-error header")
	}
}

func TestValidateFeedbackMessage_InvalidClassification_RoutedToDLQ(t *testing.T) {
	headers := validHeaders()
	for i, h := range headers {
		if string(h.Key) == "rtsa-classification" {
			headers[i].Value = []byte("FOR_OFFICIAL_USE_ONLY")
		}
	}
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: headers, Value: validFeedbackProto(t)}}
	if err := validateFeedbackMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for invalid classification")
	}
}

func TestValidateFeedbackMessage_EmptyValue_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: validHeaders(), Value: []byte{}}}
	if err := validateFeedbackMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for empty value")
	}
}

func TestValidateFeedbackMessage_InvalidProtobuf_RoutedToDLQ(t *testing.T) {
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{
		Headers: validHeaders(),
		Value:   []byte{0xFF, 0xFE, 0x00, 0x01},
	}}
	if err := validateFeedbackMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for invalid protobuf")
	}
}

func TestValidateFeedbackMessage_EmptyFeedbackID_RoutedToDLQ(t *testing.T) {
	fb := &feedbackv1.OperatorFeedback{FeedbackId: ""}
	b, _ := proto.Marshal(fb)
	w := &recordingWriter{}
	e := &mockWriteEvent{record: transform.Record{Headers: validHeaders(), Value: b}}
	if err := validateFeedbackMessage(e, w); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.routedTo == "" {
		t.Error("expected DLQ routing for empty feedback_id")
	}
}

func TestValidateFeedbackMessage_EachRequiredHeaderMissing_RoutedToDLQ(t *testing.T) {
	for _, missing := range requiredFeedbackHeaders {
		t.Run("missing_"+missing, func(t *testing.T) {
			w := &recordingWriter{}
			e := &mockWriteEvent{record: transform.Record{
				Headers: headersWithout(missing),
				Value:   validFeedbackProto(t),
			}}
			if err := validateFeedbackMessage(e, w); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if w.routedTo == "" {
				t.Errorf("expected DLQ routing when %q header is missing", missing)
			}
		})
	}
}
