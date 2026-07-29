// CLASSIFICATION: UNCLASSIFIED
package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
)

// Assigner handles alert assignment business logic.
type Assigner struct {
	queue  *AlertQueue
	logger *slog.Logger
}

// NewAssigner creates a new Assigner.
func NewAssigner(queue *AlertQueue, logger *slog.Logger) *Assigner {
	return &Assigner{
		queue:  queue,
		logger: logger,
	}
}

// Assign validates the request, updates the alert with assignment info in the queue,
// and returns the assignment timestamp.
func (a *Assigner) Assign(ctx context.Context, req *inferencev1.AssignAlertRequest) (*time.Time, error) {
	if req.GetAlertId() == "" {
		return nil, fmt.Errorf("[domain].[Assigner.Assign]: alert_id is required")
	}
	if req.GetAssignerOperatorId() == "" {
		return nil, fmt.Errorf("[domain].[Assigner.Assign]: assigner_operator_id is required")
	}
	if req.GetAssigneeOperatorId() == "" {
		return nil, fmt.Errorf("[domain].[Assigner.Assign]: assignee_operator_id is required")
	}

	assignedAt, err := a.queue.Assign(req.GetAlertId(), req.GetAssignerOperatorId(), req.GetAssigneeOperatorId(), req.GetComment())
	if err != nil {
		return nil, fmt.Errorf("[domain].[Assigner.Assign](%s): %w", req.GetAlertId(), err)
	}

	a.logger.InfoContext(ctx, "alert assigned",
		"alert_id", req.GetAlertId(),
		"assigner_id", req.GetAssignerOperatorId(),
		"assignee_id", req.GetAssigneeOperatorId(),
	)

	return assignedAt, nil
}
