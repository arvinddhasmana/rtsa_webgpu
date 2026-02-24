// CLASSIFICATION: UNCLASSIFIED
package interceptors

import (
"context"

"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/classification"
"google.golang.org/grpc"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/metadata"
"google.golang.org/grpc/status"
)

const metadataClassificationKey = "rtsa-classification"

// UnaryClassificationInterceptor checks that the incoming request's
// classification level does not exceed the service's ceiling.
func UnaryClassificationInterceptor(guard *classification.Guard) grpc.UnaryServerInterceptor {
return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler) (interface{}, error) {
if err := checkClassificationFromMD(ctx, guard); err != nil {
return nil, err
}
return handler(ctx, req)
}
}

// StreamClassificationInterceptor checks classification for streaming RPCs.
func StreamClassificationInterceptor(guard *classification.Guard) grpc.StreamServerInterceptor {
return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) error {
if err := checkClassificationFromMD(ss.Context(), guard); err != nil {
return err
}
return handler(srv, ss)
}
}

func checkClassificationFromMD(ctx context.Context, guard *classification.Guard) error {
md, ok := metadata.FromIncomingContext(ctx)
if !ok {
return nil // No metadata — allow (classification enforced at data level)
}
vals := md.Get(metadataClassificationKey)
if len(vals) == 0 {
return nil
}
level := classification.StringToLevel(vals[0])
if err := guard.Check(level); err != nil {
return status.Errorf(codes.PermissionDenied, "%s", err.Error())
}
return nil
}
