// CLASSIFICATION: UNCLASSIFIED
package health_test

import (
"context"
"testing"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/health"
grpcmetadata "google.golang.org/grpc/metadata"
)

func TestChecker_AllServing(t *testing.T) {
c := health.NewChecker()
c.Register("a")
c.Register("b")
c.SetStatus("a", health.StatusServing)
c.SetStatus("b", health.StatusServing)
if c.Overall() != health.StatusServing {
t.Error("expected StatusServing")
}
}

func TestChecker_OneNotServing(t *testing.T) {
c := health.NewChecker()
c.Register("a")
c.Register("b")
c.SetStatus("a", health.StatusServing)
c.SetStatus("b", health.StatusNotServing)
if c.Overall() != health.StatusNotServing {
t.Error("expected StatusNotServing")
}
}

func TestChecker_AllUnknown(t *testing.T) {
c := health.NewChecker()
c.Register("a")
if c.Overall() != health.StatusUnknown {
t.Error("expected StatusUnknown")
}
}

func TestServer_Check_Overall(t *testing.T) {
c := health.NewChecker()
c.Register("svc")
c.SetStatus("svc", health.StatusServing)
srv := health.NewServer(c)

resp, err := srv.Check(context.Background(), &commonv1.HealthCheckRequest{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.Status != commonv1.HealthCheckResponse_SERVING_STATUS_SERVING {
t.Errorf("expected SERVING, got %v", resp.Status)
}
}

func TestServer_Check_SpecificService(t *testing.T) {
c := health.NewChecker()
c.Register("grpc")
c.Register("db")
c.SetStatus("grpc", health.StatusServing)
c.SetStatus("db", health.StatusNotServing)
srv := health.NewServer(c)

resp, err := srv.Check(context.Background(), &commonv1.HealthCheckRequest{Service: "grpc"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.Status != commonv1.HealthCheckResponse_SERVING_STATUS_SERVING {
t.Errorf("expected SERVING for grpc, got %v", resp.Status)
}
}

func TestChecker_ComponentStatus_Unknown(t *testing.T) {
c := health.NewChecker()
status := c.ComponentStatus("unregistered")
if status != health.StatusUnknown {
t.Errorf("expected StatusUnknown for unregistered, got %v", status)
}
}

func TestChecker_EmptyChecker(t *testing.T) {
c := health.NewChecker()
if c.Overall() != health.StatusUnknown {
t.Error("expected StatusUnknown for empty checker")
}
}

func TestChecker_SetStatus_NotServing(t *testing.T) {
c := health.NewChecker()
c.Register("a")
c.SetStatus("a", health.StatusNotServing)
if c.ComponentStatus("a") != health.StatusNotServing {
t.Error("expected StatusNotServing after SetStatus")
}
}

func TestServer_Check_NotServing(t *testing.T) {
c := health.NewChecker()
c.Register("svc")
c.SetStatus("svc", health.StatusNotServing)
srv := health.NewServer(c)

resp, err := srv.Check(context.Background(), &commonv1.HealthCheckRequest{})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.Status != commonv1.HealthCheckResponse_SERVING_STATUS_NOT_SERVING {
t.Errorf("expected NOT_SERVING, got %v", resp.Status)
}
}

func TestServer_Check_UnknownComponent(t *testing.T) {
c := health.NewChecker()
srv := health.NewServer(c)
resp, err := srv.Check(context.Background(), &commonv1.HealthCheckRequest{Service: "unknown"})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if resp.Status != commonv1.HealthCheckResponse_SERVING_STATUS_SERVICE_UNKNOWN {
t.Errorf("expected SERVICE_UNKNOWN, got %v", resp.Status)
}
}

// mockWatchStream implements grpc.ServerStreamingServer[commonv1.HealthCheckResponse]
type mockWatchStream struct {
ctx      context.Context
received []*commonv1.HealthCheckResponse
}

func (m *mockWatchStream) Send(resp *commonv1.HealthCheckResponse) error {
m.received = append(m.received, resp)
return nil
}

func (m *mockWatchStream) SetHeader(md grpcmetadata.MD) error  { return nil }
func (m *mockWatchStream) SendHeader(md grpcmetadata.MD) error { return nil }
func (m *mockWatchStream) SetTrailer(md grpcmetadata.MD)       {}
func (m *mockWatchStream) Context() context.Context            { return m.ctx }
func (m *mockWatchStream) SendMsg(v interface{}) error         { return nil }
func (m *mockWatchStream) RecvMsg(v interface{}) error         { return nil }

func TestServer_Watch_ContextCancellation(t *testing.T) {
ctx, cancel := context.WithCancel(context.Background())

c := health.NewChecker()
c.Register("svc")
c.SetStatus("svc", health.StatusServing)
srv := health.NewServer(c)

stream := &mockWatchStream{ctx: ctx}

// Cancel context immediately so Watch returns
go func() {
cancel()
}()

req := &commonv1.HealthCheckRequest{}
err := srv.Watch(req, stream)
// Should return due to context cancellation
if err != nil && err != context.Canceled {
t.Errorf("expected nil or context.Canceled, got: %v", err)
}
}
