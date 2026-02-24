// CLASSIFICATION: UNCLASSIFIED
package testutil

import (
"net"
"testing"

"google.golang.org/grpc"
"google.golang.org/grpc/credentials/insecure"
)

// StartTestGRPCServer starts an in-process gRPC server for testing.
func StartTestGRPCServer(t *testing.T, registerFn func(s *grpc.Server)) (addr string, cleanup func()) {
t.Helper()
lis, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("testutil: listen: %v", err)
}

srv := grpc.NewServer()
registerFn(srv)

go func() {
if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
// Expected on cleanup
}
}()

return lis.Addr().String(), func() {
srv.GracefulStop()
lis.Close()
}
}

// DialTestGRPC creates a client connection to a test gRPC server.
func DialTestGRPC(t *testing.T, addr string) *grpc.ClientConn {
t.Helper()
conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
t.Fatalf("testutil: dial gRPC: %v", err)
}
t.Cleanup(func() { conn.Close() })
return conn
}
