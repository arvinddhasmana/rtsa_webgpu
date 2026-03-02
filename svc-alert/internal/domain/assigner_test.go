// CLASSIFICATION: UNCLASSIFIED
package domain_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
)

func newTestAssignerLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestAssigner_Assign(t *testing.T) {
	q := domain.NewAlertQueue(10)
	assigner := domain.NewAssigner(q, newTestAssignerLogger())

	// Enqueue a test alert to assign
	q.Enqueue(makeAlert("alert-1", commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED, time.Now()))

	tests := []struct {
		name          string
		req           *inferencev1.AssignAlertRequest
		expectedError error
	}{
		{
			name: "successful assignment",
			req: &inferencev1.AssignAlertRequest{
				AlertId:            "alert-1",
				AssignerOperatorId: "op-admin",
				AssigneeOperatorId: "op-analyst",
				Comment:            "Please investigate",
			},
			expectedError: nil,
		},
		{
			name: "missing alert id",
			req: &inferencev1.AssignAlertRequest{
				AlertId:            "",
				AssignerOperatorId: "op-admin",
				AssigneeOperatorId: "op-analyst",
			},
			expectedError: errors.New("alert_id is required"), // Contains check
		},
		{
			name: "missing assigner id",
			req: &inferencev1.AssignAlertRequest{
				AlertId:            "alert-1",
				AssignerOperatorId: "",
				AssigneeOperatorId: "op-analyst",
			},
			expectedError: errors.New("assigner_operator_id is required"),
		},
		{
			name: "missing assignee id",
			req: &inferencev1.AssignAlertRequest{
				AlertId:            "alert-1",
				AssignerOperatorId: "op-admin",
				AssigneeOperatorId: "",
			},
			expectedError: errors.New("assignee_operator_id is required"),
		},
		{
			name: "alert not found",
			req: &inferencev1.AssignAlertRequest{
				AlertId:            "non-existent-alert",
				AssignerOperatorId: "op-admin",
				AssigneeOperatorId: "op-analyst",
			},
			expectedError: domain.ErrAlertNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assignedAt, err := assigner.Assign(context.Background(), tt.req)

			if tt.expectedError != nil {
				if err == nil {
					t.Fatalf("expected error '%v', got nil", tt.expectedError)
				}
				if tt.expectedError == domain.ErrAlertNotFound {
					if !errors.Is(err, tt.expectedError) {
						t.Errorf("expected error %v, got %v", tt.expectedError, err)
					}
				} else {
					if err.Error() != "[domain].[Assigner.Assign]: "+tt.expectedError.Error() {
						// Less strict substring match just in case
						if !errors.Is(err, tt.expectedError) && err.Error() != "[domain].[Assigner.Assign]: "+tt.expectedError.Error() {
						   errStr := err.Error()
						   expStr := tt.expectedError.Error()
						   found := false
						   // substring check
						   for i := 0; i < len(errStr)-len(expStr)+1; i++ {
						       if errStr[i:i+len(expStr)] == expStr {
						           found = true
						           break
						       }
						   }
						   if !found {
							t.Errorf("expected error containing '%v', got '%v'", tt.expectedError, err)
						   }
						}
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if assignedAt == nil {
					t.Error("expected valid assignment timestamp, got nil")
				}
			}
		})
	}
}
