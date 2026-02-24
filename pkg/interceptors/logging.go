// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
"context"
"time"

"go.uber.org/zap"
"google.golang.org/grpc"
"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor logs gRPC unary call details.
// NEVER logs request/response payloads (may contain classified data).
func UnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler) (interface{}, error) {
start := time.Now()
resp, err := handler(ctx, req)
duration := time.Since(start)

code := status.Code(err)
if err != nil {
logger.Error("grpc unary call failed",
zap.String("method", info.FullMethod),
zap.Duration("duration", duration),
zap.String("status_code", code.String()),
zap.Error(err))
} else {
logger.Info("grpc unary call",
zap.String("method", info.FullMethod),
zap.Duration("duration", duration),
zap.String("status_code", code.String()))
}
return resp, err
}
}

// StreamLoggingInterceptor logs gRPC stream lifecycle.
func StreamLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) error {
start := time.Now()
err := handler(srv, ss)
duration := time.Since(start)

code := status.Code(err)
if err != nil {
logger.Error("grpc stream failed",
zap.String("method", info.FullMethod),
zap.Duration("duration", duration),
zap.String("status_code", code.String()),
zap.Error(err))
} else {
logger.Info("grpc stream completed",
zap.String("method", info.FullMethod),
zap.Duration("duration", duration),
zap.String("status_code", code.String()))
}
return err
}
}
