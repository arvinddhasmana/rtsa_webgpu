// CLASSIFICATION: UNCLASSIFIED
package main

import (
	feedbackv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/feedback/v1"
	"github.com/redpanda-data/redpanda/src/transform-sdk/go/transform"
	"google.golang.org/protobuf/proto"
)

// requiredFeedbackHeaders lists all mandatory message headers for feedback topics.
var requiredFeedbackHeaders = []string{
	"rtsa-classification",
	"rtsa-source-service",
	"rtsa-trace-id",
	"rtsa-timestamp",
	"rtsa-schema-version",
}

// validClassifications contains the set of accepted classification header values.
var validClassifications = map[string]bool{
	"UNCLASSIFIED": true,
	"PROTECTED_A":  true,
	"PROTECTED_B":  true,
	"PROTECTED_C":  true,
	"SECRET":       true,
}

// validateFeedbackMessage runs on every record written to feedback.operator.* topics.
// On validation pass it adds rtsa-validated=true and writes to the output topic.
// On validation failure it adds rtsa-validation-error and routes to DLQ.
func validateFeedbackMessage(event transform.WriteEvent, writer transform.RecordWriter) error {
	record := event.Record()

	// Validate required headers.
	for _, h := range requiredFeedbackHeaders {
		if getHeader(record, h) == "" {
			return writeToDLQ(writer, record, "missing header: "+h)
		}
	}

	// Validate classification value.
	classification := getHeader(record, "rtsa-classification")
	if !validClassifications[classification] {
		return writeToDLQ(writer, record, "invalid classification: "+classification)
	}

	// Validate non-empty record value.
	if len(record.Value) == 0 {
		return writeToDLQ(writer, record, "empty record value")
	}

	// Validate protobuf can be deserialized as OperatorFeedback.
	var feedback feedbackv1.OperatorFeedback
	if err := proto.Unmarshal(record.Value, &feedback); err != nil {
		return writeToDLQ(writer, record, "invalid protobuf: "+err.Error())
	}

	// Validate required domain fields.
	if feedback.GetFeedbackId() == "" {
		return writeToDLQ(writer, record, "missing feedback_id")
	}

	// All checks passed — stamp the record and forward.
	record.Headers = append(record.Headers, transform.RecordHeader{
		Key:   []byte("rtsa-validated"),
		Value: []byte("true"),
	})
	return writer.Write(record)
}

// getHeader returns the value of the first header whose key matches name,
// or an empty string if the header is absent.
func getHeader(record transform.Record, name string) string {
	for _, h := range record.Headers {
		if string(h.Key) == name {
			return string(h.Value)
		}
	}
	return ""
}

// writeToDLQ stamps the record with a validation-error header and routes it to
// the dead-letter topic so that no message is ever silently dropped.
func writeToDLQ(writer transform.RecordWriter, record transform.Record, reason string) error {
	record.Headers = append(record.Headers, transform.RecordHeader{
		Key:   []byte("rtsa-validation-error"),
		Value: []byte(reason),
	})
	return writer.Write(record, transform.ToTopic("dlq.feedback"))
}
