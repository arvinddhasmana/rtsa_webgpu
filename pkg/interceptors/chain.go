// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
"go.opentelemetry.io/otel/metric"
"go.uber.org/zap"
"google.golang.org/grpc"

"github.com/arvinddhasmana/rtsa_webgpu/pkg/classification"
)

// ChainConfig configures the standard interceptor chain.
type ChainConfig struct {
Logger              *zap.Logger
Meter               metric.Meter
ClassificationGuard *classification.Guard
ServiceName         string
}

// BuildUnaryServerInterceptors returns the standard unary interceptor chain.
// Order: Recovery → Tracing → Metrics → Classification → Logging
func BuildUnaryServerInterceptors(cfg ChainConfig) []grpc.UnaryServerInterceptor {
interceptors := []grpc.UnaryServerInterceptor{
UnaryRecoveryInterceptor(cfg.Logger),
UnaryTracingInterceptor(),
UnaryLoggingInterceptor(cfg.Logger),
}
if cfg.Meter != nil {
interceptors = append(interceptors, UnaryMetricsInterceptor(cfg.Meter, cfg.ServiceName))
}
if cfg.ClassificationGuard != nil {
interceptors = append(interceptors, UnaryClassificationInterceptor(cfg.ClassificationGuard))
}
return interceptors
}

// BuildStreamServerInterceptors returns the standard stream interceptor chain.
// Order: Recovery → Tracing → Metrics → Classification → Logging
func BuildStreamServerInterceptors(cfg ChainConfig) []grpc.StreamServerInterceptor {
interceptors := []grpc.StreamServerInterceptor{
StreamRecoveryInterceptor(cfg.Logger),
StreamTracingInterceptor(),
StreamLoggingInterceptor(cfg.Logger),
}
if cfg.Meter != nil {
interceptors = append(interceptors, StreamMetricsInterceptor(cfg.Meter, cfg.ServiceName))
}
if cfg.ClassificationGuard != nil {
interceptors = append(interceptors, StreamClassificationInterceptor(cfg.ClassificationGuard))
}
return interceptors
}

// BuildDialOptions returns standard client dial options.
func BuildDialOptions(cfg ChainConfig) []grpc.DialOption {
// Client-side options (no interceptors for now, mTLS handled separately)
return []grpc.DialOption{}
}
