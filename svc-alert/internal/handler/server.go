// CLASSIFICATION: UNCLASSIFIED
package handler

import (
"context"

inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
)

// AlertServer composes all handlers and implements inferencev1.AlertServiceServer.
type AlertServer struct {
inferencev1.UnimplementedAlertServiceServer
	stream      *StreamHandler
	acknowledge *AcknowledgeHandler
	details     *DetailsHandler
	assign      *AssignHandler
}

// NewAlertServer creates an AlertServer that satisfies the AlertServiceServer interface.
func NewAlertServer(
	stream *StreamHandler,
	acknowledge *AcknowledgeHandler,
	details *DetailsHandler,
	assign *AssignHandler,
) *AlertServer {
	return &AlertServer{
		stream:      stream,
		acknowledge: acknowledge,
		details:     details,
		assign:      assign,
	}
}

// StreamAlerts implements AlertService.StreamAlerts.
func (s *AlertServer) StreamAlerts(
req *inferencev1.StreamAlertsRequest,
stream inferencev1.AlertService_StreamAlertsServer,
) error {
return s.stream.StreamAlerts(req, stream)
}

// AcknowledgeAlert implements AlertService.AcknowledgeAlert.
func (s *AlertServer) AcknowledgeAlert(
ctx context.Context,
req *inferencev1.AcknowledgeAlertRequest,
) (*inferencev1.AcknowledgeAlertResponse, error) {
return s.acknowledge.AcknowledgeAlert(ctx, req)
}

// GetAlertDetails implements AlertService.GetAlertDetails.
func (s *AlertServer) GetAlertDetails(
ctx context.Context,
req *inferencev1.GetAlertDetailsRequest,
) (*inferencev1.AnomalyAlert, error) {
	return s.details.GetAlertDetails(ctx, req)
}

// AssignAlert implements AlertService.AssignAlert.
func (s *AlertServer) AssignAlert(
	ctx context.Context,
	req *inferencev1.AssignAlertRequest,
) (*inferencev1.AssignAlertResponse, error) {
	return s.assign.AssignAlert(ctx, req)
}
