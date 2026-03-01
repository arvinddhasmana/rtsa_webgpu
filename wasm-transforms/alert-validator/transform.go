// CLASSIFICATION: UNCLASSIFIED
package main

import (
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/redpanda-data/redpanda/src/transform-sdk/go/transform"
	"google.golang.org/protobuf/encoding/protojson"
)

// requiredAlertHeaders lists all mandatory message headers for alert topics.
var requiredAlertHeaders = []string{
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

// validateAlertMessage runs on every record written to alerts.anomaly.* topics.
// On validation pass it adds rtsa-validated=true and writes to the output topic.
// On validation failure it adds rtsa-validation-error and routes to DLQ.
func validateAlertMessage(event transform.WriteEvent, writer transform.RecordWriter) error {
	record := event.Record()

	// Validate required headers.
	for _, h := range requiredAlertHeaders {
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

	// Validate protobuf can be deserialized as AnomalyAlert.
	var alert inferencev1.AnomalyAlert
	if err := protojson.Unmarshal(record.Value, &alert); err != nil {
		return writeToDLQ(writer, record, "invalid protobuf: "+err.Error())
	}

	// Validate required domain fields.
	if alert.GetAlertId() == "" {
		return writeToDLQ(writer, record, "missing alert_id")
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
	return writer.Write(record, transform.ToTopic("dlq.alerts"))
}
