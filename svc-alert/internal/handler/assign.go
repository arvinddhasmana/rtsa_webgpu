// CLASSIFICATION: UNCLASSIFIED
package handler

import (
	"context"
	"log/slog"

	auditv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/audit/v1"
	inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/audit"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AssignHandler serves the unary AssignAlert RPC.
type AssignHandler struct {
	assigner *domain.Assigner
	emitter  *audit.Emitter
	logger   *slog.Logger
}

// NewAssignHandler creates a new AssignHandler.
func NewAssignHandler(assigner *domain.Assigner, emitter *audit.Emitter, logger *slog.Logger) *AssignHandler {
	return &AssignHandler{
		assigner: assigner,
		emitter:  emitter,
		logger:   logger,
	}
}

// AssignAlert processes an alert assignment request.
func (h *AssignHandler) AssignAlert(ctx context.Context, req *inferencev1.AssignAlertRequest) (*inferencev1.AssignAlertResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	assignedAt, err := h.assigner.Assign(ctx, req)
	if err != nil {
		h.logger.WarnContext(ctx, "assignment failed",
			"alert_id", req.GetAlertId(),
			"error", err.Error(),
		)

		if err == domain.ErrAlertNotFound {
			return nil, status.Errorf(codes.NotFound, "alert %s not found", req.GetAlertId())
		}

		// Map domain validation errors back to InvalidArgument if they mention missing parameters
		if req.GetAlertId() == "" || req.GetAssignerOperatorId() == "" || req.GetAssigneeOperatorId() == "" {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}

		return nil, status.Errorf(codes.Internal, "failed to assign alert: %v", err)
	}

	if h.emitter != nil {
		h.emitter.Emit(ctx, audit.AuditParams{
			EventType:    audit.EventAlertAcknowledged, // Reusing event or using custom "alert.assigned"
			ActorID:      req.GetAssignerOperatorId(),
			ActorType:    auditv1.ActorType_ACTOR_TYPE_OPERATOR,
			ResourceType: "alert",
			ResourceID:   req.GetAlertId(),
			Action:       auditv1.AuditAction_AUDIT_ACTION_UPDATE, // Updating assignment
		})
	}

	return &inferencev1.AssignAlertResponse{
		Success:    true,
		AssignedAt: timestamppb.New(*assignedAt),
	}, nil
}
