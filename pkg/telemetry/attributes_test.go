// CLASSIFICATION: UNCLASSIFIED
package telemetry_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/telemetry"
)

func TestAttributes(t *testing.T) {
	if telemetry.AttrServiceName == "" {
		t.Error("expected AttrServiceName to be populated")
	}
	if string(telemetry.AttrServiceName) != "service.name" {
		t.Errorf("expected Key('service.name'), got %v", telemetry.AttrServiceName)
	}

	// Just a sample check for others to ensure they are initialized
	keys := []string{
		string(telemetry.AttrSensorType),
		string(telemetry.AttrEntityType),
		string(telemetry.AttrClassification),
		string(telemetry.AttrAnomalyType),
		string(telemetry.AttrAlertSeverity),
		string(telemetry.AttrFeedbackType),
		string(telemetry.AttrOperatorID),
		string(telemetry.AttrTrackStatus),
		string(telemetry.AttrHostileClass),
		string(telemetry.AttrTopicName),
		string(telemetry.AttrConsumerGroup),
		string(telemetry.AttrGRPCMethod),
		string(telemetry.AttrGRPCStatusCode),
	}

	for _, k := range keys {
		if k == "" {
			t.Error("expected all attribute keys to be populated")
		}
	}
}
