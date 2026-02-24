// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
"context"
"fmt"
"runtime/debug"

"go.uber.org/zap"
"google.golang.org/grpc"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// UnaryRecoveryInterceptor catches panics in handlers and converts them to
// gRPC INTERNAL errors. Logs the panic stack trace at ERROR level.
func UnaryRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler) (resp interface{}, err error) {
defer func() {
if r := recover(); r != nil {
stack := debug.Stack()
logger.Error("grpc unary handler panic recovered",
zap.String("method", info.FullMethod),
zap.String("panic", fmt.Sprintf("%v", r)),
zap.ByteString("stack", stack))
err = status.Errorf(codes.Internal, "internal server error")
}
}()
return handler(ctx, req)
}
}

// StreamRecoveryInterceptor catches panics in stream handlers.
func StreamRecoveryInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) (err error) {
defer func() {
if r := recover(); r != nil {
stack := debug.Stack()
logger.Error("grpc stream handler panic recovered",
zap.String("method", info.FullMethod),
zap.String("panic", fmt.Sprintf("%v", r)),
zap.ByteString("stack", stack))
err = status.Errorf(codes.Internal, "internal server error")
}
}()
return handler(srv, ss)
}
}
