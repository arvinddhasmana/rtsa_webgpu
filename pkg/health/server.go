// CLASSIFICATION: UNCLASSIFIED
package health

import (
	"context"
	"time"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	grpchealthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

// Server implements the gRPC HealthService from common.v1.
type Server struct {
	commonv1.UnimplementedHealthServiceServer
	checker *Checker
}

// NewServer creates a health gRPC server backed by the given Checker.
func NewServer(checker *Checker) *Server {
	return &Server{checker: checker}
}

// RegisterAll registers both the RTSA custom health service (commonv1.HealthService)
// and the standard gRPC health protocol (grpc.health.v1) on the given server.
// The standard registration enables grpc_health_probe to work in distroless containers.
func RegisterAll(srv *grpc.Server, checker *Checker) {
	// RTSA custom health (used by svc-query, web UI, inter-service checks)
	commonv1.RegisterHealthServiceServer(srv, NewServer(checker))
	// Standard gRPC health protocol — required for grpc_health_probe health checks
	hs := grpchealth.NewServer()
	hs.SetServingStatus("", grpchealthv1.HealthCheckResponse_SERVING)
	grpchealthv1.RegisterHealthServer(srv, hs)
}

// Check implements HealthService.Check.
func (s *Server) Check(ctx context.Context, req *commonv1.HealthCheckRequest) (*commonv1.HealthCheckResponse, error) {
	var status Status
	if req.GetService() != "" {
		status = s.checker.ComponentStatus(req.GetService())
	} else {
		status = s.checker.Overall()
	}
	return &commonv1.HealthCheckResponse{
		Status: toProtoStatus(status),
	}, nil
}

// Watch implements HealthService.Watch (server-streaming).
// Sends status updates every 5 seconds until the context is cancelled.
func (s *Server) Watch(req *commonv1.HealthCheckRequest, stream commonv1.HealthService_WatchServer) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			var status Status
			if req.GetService() != "" {
				status = s.checker.ComponentStatus(req.GetService())
			} else {
				status = s.checker.Overall()
			}
			if err := stream.Send(&commonv1.HealthCheckResponse{
				Status: toProtoStatus(status),
			}); err != nil {
				return err
			}
		}
	}
}

func toProtoStatus(s Status) commonv1.HealthCheckResponse_ServingStatus {
	switch s {
	case StatusServing:
		return commonv1.HealthCheckResponse_SERVING_STATUS_SERVING
	case StatusNotServing:
		return commonv1.HealthCheckResponse_SERVING_STATUS_NOT_SERVING
	default:
		return commonv1.HealthCheckResponse_SERVING_STATUS_SERVICE_UNKNOWN
	}
}

