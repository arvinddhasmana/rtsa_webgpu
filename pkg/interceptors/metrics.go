// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
"context"
"time"

"go.opentelemetry.io/otel/attribute"
"go.opentelemetry.io/otel/metric"
"google.golang.org/grpc"
"google.golang.org/grpc/status"
)

// Standard gRPC metric names.
const (
MetricGRPCRequestTotal    = "rtsa_grpc_requests_total"
MetricGRPCRequestDuration = "rtsa_grpc_request_duration_seconds"
MetricGRPCActiveStreams    = "rtsa_grpc_active_streams"
)

// UnaryMetricsInterceptor records request count and duration histograms.
func UnaryMetricsInterceptor(meter metric.Meter, serviceName string) grpc.UnaryServerInterceptor {
counter, _ := meter.Int64Counter(MetricGRPCRequestTotal,
metric.WithDescription("Total gRPC requests"))
histogram, _ := meter.Float64Histogram(MetricGRPCRequestDuration,
metric.WithDescription("gRPC request duration in seconds"))

return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler) (interface{}, error) {
start := time.Now()
resp, err := handler(ctx, req)
dur := time.Since(start).Seconds()

code := status.Code(err)
attrs := attribute.NewSet(
attribute.String("method", info.FullMethod),
attribute.String("status_code", code.String()),
attribute.String("service", serviceName),
)
counter.Add(ctx, 1, metric.WithAttributeSet(attrs))
histogram.Record(ctx, dur, metric.WithAttributeSet(attrs))
return resp, err
}
}

// StreamMetricsInterceptor records stream lifecycle metrics.
func StreamMetricsInterceptor(meter metric.Meter, serviceName string) grpc.StreamServerInterceptor {
counter, _ := meter.Int64Counter(MetricGRPCRequestTotal,
metric.WithDescription("Total gRPC requests"))
activeGauge, _ := meter.Int64UpDownCounter(MetricGRPCActiveStreams,
metric.WithDescription("Active gRPC streams"))

return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) error {
ctx := ss.Context()
attrs := attribute.NewSet(
attribute.String("method", info.FullMethod),
attribute.String("service", serviceName),
)
activeGauge.Add(ctx, 1, metric.WithAttributeSet(attrs))
defer activeGauge.Add(ctx, -1, metric.WithAttributeSet(attrs))

err := handler(srv, ss)

code := status.Code(err)
finalAttrs := attribute.NewSet(
attribute.String("method", info.FullMethod),
attribute.String("status_code", code.String()),
attribute.String("service", serviceName),
)
counter.Add(ctx, 1, metric.WithAttributeSet(finalAttrs))
return err
}
}
