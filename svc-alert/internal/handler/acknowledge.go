// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"
"errors"
"log/slog"

inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-alert/internal/domain"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
"google.golang.org/protobuf/types/known/timestamppb"
)

// AcknowledgeHandler implements the AcknowledgeAlert gRPC unary RPC.
type AcknowledgeHandler struct {
acknowledger *domain.Acknowledger
logger       *slog.Logger
}

// NewAcknowledgeHandler creates a new AcknowledgeHandler.
func NewAcknowledgeHandler(ack *domain.Acknowledger, logger *slog.Logger) *AcknowledgeHandler {
return &AcknowledgeHandler{acknowledger: ack, logger: logger}
}

// AcknowledgeAlert implements AlertService.AcknowledgeAlert.
// Returns NOT_FOUND if the alert_id does not exist.
// Returns INVALID_ARGUMENT if required fields are missing.
func (h *AcknowledgeHandler) AcknowledgeAlert(
ctx context.Context,
req *inferencev1.AcknowledgeAlertRequest,
) (*inferencev1.AcknowledgeAlertResponse, error) {
if req == nil {
return nil, status.Error(codes.InvalidArgument, "request must not be nil")
}
if req.GetAlertId() == "" {
return nil, status.Error(codes.InvalidArgument, "alert_id is required")
}
if req.GetOperatorId() == "" {
return nil, status.Error(codes.InvalidArgument, "operator_id is required")
}

ackedAt, err := h.acknowledger.Acknowledge(ctx, req)
if err != nil {
if errors.Is(err, domain.ErrAlertNotFound) {
return nil, status.Errorf(codes.NotFound, "alert %q not found", req.GetAlertId())
}
h.logger.ErrorContext(ctx, "AcknowledgeAlert failed",
"alert_id", req.GetAlertId(),
"error", err.Error(),
)
return nil, status.Errorf(codes.Internal, "acknowledgment failed: %v", err)
}

return &inferencev1.AcknowledgeAlertResponse{
Success:        true,
AcknowledgedAt: timestamppb.New(*ackedAt),
}, nil
}
