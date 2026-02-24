// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
"context"

"go.opentelemetry.io/otel"
"go.opentelemetry.io/otel/attribute"
"go.opentelemetry.io/otel/codes"
"google.golang.org/grpc"
)

const tracerName = "rtsa.grpc"

// UnaryTracingInterceptor adds distributed tracing to unary calls.
func UnaryTracingInterceptor() grpc.UnaryServerInterceptor {
return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler) (interface{}, error) {
tracer := otel.GetTracerProvider().Tracer(tracerName)
ctx, span := tracer.Start(ctx, info.FullMethod,
// Do not record attributes from req/resp as they may contain classified data
)
defer span.End()

span.SetAttributes(attribute.String("grpc.method", info.FullMethod))

resp, err := handler(ctx, req)
if err != nil {
span.SetStatus(codes.Error, err.Error())
span.RecordError(err)
}
return resp, err
}
}

// StreamTracingInterceptor adds distributed tracing to stream calls.
func StreamTracingInterceptor() grpc.StreamServerInterceptor {
return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) error {
tracer := otel.GetTracerProvider().Tracer(tracerName)
ctx, span := tracer.Start(ss.Context(), info.FullMethod)
defer span.End()

span.SetAttributes(attribute.String("grpc.method", info.FullMethod))

wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
err := handler(srv, wrapped)
if err != nil {
span.SetStatus(codes.Error, err.Error())
span.RecordError(err)
}
return err
}
}

type wrappedServerStream struct {
grpc.ServerStream
ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
return w.ctx
}
