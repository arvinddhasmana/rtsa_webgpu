// CLASSIFICATION: UNCLASSIFIED
//go:build integration

// Package integration_test validates the alert service handler pipeline for
// svc-alert: alert queue lifecycle (enqueue/dequeue), acknowledgement, details
// retrieval, and assignment operations.
// Tests exercise two or more interacting components without a live Redpanda container.
package integration_test

import (
	"context"
	"log/slog"
	"testing"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/handler"
)

// newAlertStack assembles a complete alert handler composed of all sub-handlers
// sharing a single AlertQueue, using a nop logger.
func newAlertStack(t *testing.T) (*domain.AlertQueue, *handler.AlertServer) {
	t.Helper()
	logger := slog.Default()
	q := domain.NewAlertQueue(100)

	acknowledger := domain.NewAcknowledger(q, nil, logger)
	assigner := domain.NewAssigner(q, logger)

	streamH := handler.NewStreamHandler(q, nil, logger)
	ackH := handler.NewAcknowledgeHandler(acknowledger, logger)
	detailsH := handler.NewDetailsHandler(q, logger)
	assignH := handler.NewAssignHandler(assigner, nil, logger)

	srv := handler.NewAlertServer(streamH, ackH, detailsH, assignH)
	return q, srv
}

// sampleAlert returns a minimal CRITICAL anomaly alert for testing.
func sampleAlert(id string) *inferencev1.AnomalyAlert {
	return &inferencev1.AnomalyAlert{
		AlertId:         id,
		TrackId:         "track-alert-test-001",
		AnomalyType:     commonv1.AnomalyType_ANOMALY_TYPE_SPEED,
		Severity:        commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL,
		ConfidenceScore: 0.93,
		Explanation:     "synthetic speed anomaly for integration test",
		Classification:  commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED,
		ModelVersion:    "rules-v1.0.0",
	}
}

// TestAlertQueue_EnqueueAndDetails_AlertRetrievable validates that the full
// enqueue → GetAlertDetails pipeline returns the expected alert data.
func TestAlertQueue_EnqueueAndDetails_AlertRetrievable(t *testing.T) {
	q, srv := newAlertStack(t)

	alert := sampleAlert("alert-int-001")
	q.Enqueue(alert)

	resp, err := srv.GetAlertDetails(context.Background(), &inferencev1.GetAlertDetailsRequest{
		AlertId: "alert-int-001",
	})
	if err != nil {
		t.Fatalf("TestAlertQueue_EnqueueAndDetails_AlertRetrievable: GetAlertDetails error: %v", err)
	}
	if resp.GetAlertId() != "alert-int-001" {
		t.Errorf("TestAlertQueue_EnqueueAndDetails_AlertRetrievable: alert_id=%q, want %q",
			resp.GetAlertId(), "alert-int-001")
	}
	if resp.GetSeverity() != commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL {
		t.Errorf("TestAlertQueue_EnqueueAndDetails_AlertRetrievable: severity=%v, want CRITICAL", resp.GetSeverity())
	}
	if resp.GetClassification() != commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED {
		t.Errorf("TestAlertQueue_EnqueueAndDetails_AlertRetrievable: classification=%v, want UNCLASSIFIED",
			resp.GetClassification())
	}
}

// TestAlertAcknowledge_ExistingAlert_AcknowledgedSuccessfully validates that the
// full enqueue → AcknowledgeAlert pipeline acknowledges the alert without error.
func TestAlertAcknowledge_ExistingAlert_AcknowledgedSuccessfully(t *testing.T) {
	q, srv := newAlertStack(t)

	alert := sampleAlert("alert-int-ack-001")
	q.Enqueue(alert)

	resp, err := srv.AcknowledgeAlert(context.Background(), &inferencev1.AcknowledgeAlertRequest{
		AlertId:    "alert-int-ack-001",
		OperatorId: "operator-test-01",
		Comment:    "acknowledged in integration test",
	})
	if err != nil {
		t.Fatalf("TestAlertAcknowledge_ExistingAlert_AcknowledgedSuccessfully: AcknowledgeAlert error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("TestAlertAcknowledge_ExistingAlert_AcknowledgedSuccessfully: success=%v, want true", resp.GetSuccess())
	}
}

// TestAlertAssign_ExistingAlert_AssignedSuccessfully validates that the
// full enqueue → AssignAlert pipeline assigns the alert to an operator.
func TestAlertAssign_ExistingAlert_AssignedSuccessfully(t *testing.T) {
	q, srv := newAlertStack(t)

	alert := sampleAlert("alert-int-assign-001")
	q.Enqueue(alert)

	resp, err := srv.AssignAlert(context.Background(), &inferencev1.AssignAlertRequest{
		AlertId:            "alert-int-assign-001",
		AssignerOperatorId: "operator-test-01",
		AssigneeOperatorId: "operator-test-02",
	})
	if err != nil {
		t.Fatalf("TestAlertAssign_ExistingAlert_AssignedSuccessfully: AssignAlert error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("TestAlertAssign_ExistingAlert_AssignedSuccessfully: success=%v, want true", resp.GetSuccess())
	}
}

// TestAlertQueue_MultipleAlerts_QueueSizeTracked validates that enqueueing
// multiple alerts is reflected in the queue depth.
func TestAlertQueue_MultipleAlerts_QueueSizeTracked(t *testing.T) {
	q, _ := newAlertStack(t)

	for i := 0; i < 5; i++ {
		id := "alert-int-multi-" + string(rune('a'+i))
		alert := sampleAlert(id)
		q.Enqueue(alert)
	}

	size := q.Size()
	if size != 5 {
		t.Errorf("TestAlertQueue_MultipleAlerts_QueueSizeTracked: queue size=%d, want 5", size)
	}
}
