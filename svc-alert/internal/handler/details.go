// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"
"log/slog"

inferencev1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/domain"
"github.com/arvinddhasmana/rtsa_webgpu/svc-alert/internal/mapper"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// DetailsHandler implements the GetAlertDetails gRPC unary RPC.
type DetailsHandler struct {
queue  *domain.AlertQueue
logger *slog.Logger
}

// NewDetailsHandler creates a new DetailsHandler.
func NewDetailsHandler(q *domain.AlertQueue, logger *slog.Logger) *DetailsHandler {
return &DetailsHandler{queue: q, logger: logger}
}

// GetAlertDetails implements AlertService.GetAlertDetails.
//
// Returns:
//   - NOT_FOUND if the alert_id does not exist.
//   - PERMISSION_DENIED if the alert classification exceeds the clearance level.
//   - INVALID_ARGUMENT if required fields are missing.
func (h *DetailsHandler) GetAlertDetails(
ctx context.Context,
req *inferencev1.GetAlertDetailsRequest,
) (*inferencev1.AnomalyAlert, error) {
if req == nil {
return nil, status.Error(codes.InvalidArgument, "request must not be nil")
}
if req.GetAlertId() == "" {
return nil, status.Error(codes.InvalidArgument, "alert_id is required")
}

qa, found := h.queue.Get(req.GetAlertId())
if !found {
return nil, status.Errorf(codes.NotFound, "alert %q not found", req.GetAlertId())
}

// MANDATORY classification filter.
if !mapper.PassesClassificationFilter(qa.Alert.GetClassification(), req.GetClearanceLevel()) {
h.logger.WarnContext(ctx, "GetAlertDetails: classification denied",
"alert_id", req.GetAlertId(),
"alert_classification", qa.Alert.GetClassification().String(),
"clearance_level", req.GetClearanceLevel().String(),
)
return nil, status.Errorf(codes.PermissionDenied,
"alert classification exceeds your clearance level")
}

return mapper.QueuedAlertToProto(qa), nil
}
