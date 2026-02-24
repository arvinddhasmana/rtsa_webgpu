// CLASSIFICATION: UNCLASSIFIED
package interceptors_test

import (
	"context"
	"errors"
	"testing"

	commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/interceptors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestLoggingInterceptor_LogsMethod(t *testing.T) {
logger := zap.NewNop()
interceptor := interceptors.UnaryLoggingInterceptor(logger)

handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return "response", nil
}

resp, err := interceptor(context.Background(), "request",
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
if resp != "response" {
t.Errorf("unexpected response: %v", resp)
}
}

func TestLoggingInterceptor_ErrorCase(t *testing.T) {
logger := zap.NewNop()
interceptor := interceptors.UnaryLoggingInterceptor(logger)

handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return nil, errors.New("test error")
}

_, err := interceptor(context.Background(), nil,
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err == nil {
t.Error("expected error to be propagated")
}
}

func TestRecoveryInterceptor_CatchesPanic(t *testing.T) {
logger := zap.NewNop()
interceptor := interceptors.UnaryRecoveryInterceptor(logger)

handler := func(ctx context.Context, req interface{}) (interface{}, error) {
panic("test panic")
}

_, err := interceptor(context.Background(), nil,
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err == nil {
t.Fatal("expected error from panic recovery")
}
st, ok := status.FromError(err)
if !ok || st.Code() != codes.Internal {
t.Errorf("expected INTERNAL status, got: %v", err)
}
}

func TestClassificationInterceptor_AllowsValid(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
interceptor := interceptors.UnaryClassificationInterceptor(guard)

md := metadata.New(map[string]string{"rtsa-classification": "UNCLASSIFIED"})
ctx := metadata.NewIncomingContext(context.Background(), md)

handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return "ok", nil
}

resp, err := interceptor(ctx, nil,
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
if resp != "ok" {
t.Errorf("unexpected response: %v", resp)
}
}

func TestClassificationInterceptor_RejectsHigh(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_UNCLASSIFIED)
interceptor := interceptors.UnaryClassificationInterceptor(guard)

md := metadata.New(map[string]string{"rtsa-classification": "SECRET"})
ctx := metadata.NewIncomingContext(context.Background(), md)

handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return "ok", nil
}

_, err := interceptor(ctx, nil,
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err == nil {
t.Fatal("expected PERMISSION_DENIED error")
}
st, ok := status.FromError(err)
if !ok || st.Code() != codes.PermissionDenied {
t.Errorf("expected PERMISSION_DENIED, got: %v", st.Code())
}
}

func TestBuildUnaryChain(t *testing.T) {
logger := zap.NewNop()
chain := interceptors.BuildUnaryServerInterceptors(interceptors.ChainConfig{
Logger:      logger,
ServiceName: "test-svc",
})
if len(chain) == 0 {
t.Error("expected at least one interceptor in chain")
}
}

func TestTracingInterceptor_NoError(t *testing.T) {
interceptor := interceptors.UnaryTracingInterceptor()
handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return "ok", nil
}
resp, err := interceptor(context.Background(), nil,
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
if resp != "ok" {
t.Errorf("expected ok, got %v", resp)
}
}

func TestStreamRecoveryInterceptor_CatchesPanic(t *testing.T) {
logger := zap.NewNop()
interceptor := interceptors.StreamRecoveryInterceptor(logger)

handler := func(srv interface{}, stream grpc.ServerStream) error {
panic("stream panic")
}

err := interceptor(nil, nil, &grpc.StreamServerInfo{FullMethod: "/test.Service/Stream"}, handler)
if err == nil {
t.Fatal("expected error from panic recovery")
}
}

func TestClassificationInterceptor_NoMetadata(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
interceptor := interceptors.UnaryClassificationInterceptor(guard)

// No metadata in context — should allow through
handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return "ok", nil
}

resp, err := interceptor(context.Background(), nil,
&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
if err != nil {
t.Errorf("unexpected error with no metadata: %v", err)
}
if resp != "ok" {
t.Errorf("expected ok, got %v", resp)
}
}

func TestBuildStreamChain(t *testing.T) {
logger := zap.NewNop()
chain := interceptors.BuildStreamServerInterceptors(interceptors.ChainConfig{
Logger:      logger,
ServiceName: "test-svc",
})
if len(chain) == 0 {
t.Error("expected at least one stream interceptor")
}
}

// mockServerStream implements grpc.ServerStream for testing
type mockServerStream struct {
ctx context.Context
}

func (m *mockServerStream) SetHeader(md metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(md metadata.MD)       {}
func (m *mockServerStream) Context() context.Context        { return m.ctx }
func (m *mockServerStream) SendMsg(v interface{}) error     { return nil }
func (m *mockServerStream) RecvMsg(v interface{}) error     { return nil }

func TestStreamLoggingInterceptor(t *testing.T) {
logger := zap.NewNop()
interceptor := interceptors.StreamLoggingInterceptor(logger)

stream := &mockServerStream{ctx: context.Background()}
handler := func(srv interface{}, stream grpc.ServerStream) error {
return nil
}

err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
}

func TestStreamTracingInterceptor(t *testing.T) {
interceptor := interceptors.StreamTracingInterceptor()

stream := &mockServerStream{ctx: context.Background()}
handler := func(srv interface{}, s grpc.ServerStream) error {
return nil
}

err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
}

func TestStreamClassificationInterceptor_Allow(t *testing.T) {
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
interceptor := interceptors.StreamClassificationInterceptor(guard)

md := metadata.New(map[string]string{"rtsa-classification": "UNCLASSIFIED"})
ctx := metadata.NewIncomingContext(context.Background(), md)
stream := &mockServerStream{ctx: ctx}

handler := func(srv interface{}, s grpc.ServerStream) error {
return nil
}

err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
}

func TestBuildDialOptions(t *testing.T) {
logger := zap.NewNop()
opts := interceptors.BuildDialOptions(interceptors.ChainConfig{
Logger:      logger,
ServiceName: "test-svc",
})
// Should return a slice (even if empty)
_ = opts
}

func TestBuildUnaryChainWithMeter(t *testing.T) {
logger := zap.NewNop()
chain := interceptors.BuildUnaryServerInterceptors(interceptors.ChainConfig{
Logger:              logger,
ServiceName:         "test-svc",
ClassificationGuard: nil,
Meter:               nil,
})
if len(chain) < 3 {
t.Errorf("expected at least 3 interceptors, got %d", len(chain))
}
}

func TestUnaryMetricsInterceptor_WithGlobalMeter(t *testing.T) {
// Initialize a minimal OTel meter
meter := otel.GetMeterProvider().Meter("test")
interceptor := interceptors.UnaryMetricsInterceptor(meter, "test-svc")

handler := func(ctx context.Context, req interface{}) (interface{}, error) {
return "ok", nil
}

resp, err := interceptor(context.Background(), nil,
&grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
if resp != "ok" {
t.Errorf("expected ok, got %v", resp)
}
}

func TestStreamMetricsInterceptor_WithGlobalMeter(t *testing.T) {
meter := otel.GetMeterProvider().Meter("test")
interceptor := interceptors.StreamMetricsInterceptor(meter, "test-svc")

stream := &mockServerStream{ctx: context.Background()}
handler := func(srv interface{}, s grpc.ServerStream) error {
return nil
}

err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test/Method"}, handler)
if err != nil {
t.Errorf("unexpected error: %v", err)
}
}

func TestBuildUnaryChainWithAllOptions(t *testing.T) {
logger := zap.NewNop()
meter := otel.GetMeterProvider().Meter("test")
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
chain := interceptors.BuildUnaryServerInterceptors(interceptors.ChainConfig{
Logger:              logger,
ServiceName:         "test-svc",
ClassificationGuard: guard,
Meter:               meter,
})
if len(chain) < 5 {
t.Errorf("expected at least 5 interceptors, got %d", len(chain))
}
}

func TestBuildStreamChainWithAllOptions(t *testing.T) {
logger := zap.NewNop()
meter := otel.GetMeterProvider().Meter("test")
guard := classification.NewGuard(commonv1.ClassificationLevel_CLASSIFICATION_LEVEL_SECRET)
chain := interceptors.BuildStreamServerInterceptors(interceptors.ChainConfig{
Logger:              logger,
ServiceName:         "test-svc",
ClassificationGuard: guard,
Meter:               meter,
})
if len(chain) < 5 {
t.Errorf("expected at least 5 stream interceptors, got %d", len(chain))
}
}
