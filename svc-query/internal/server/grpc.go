// CLASSIFICATION: UNCLASSIFIED

// Package server wires up and starts the gRPC server for the query service.
package server

import (
"context"
"crypto/tls"
"crypto/x509"
"fmt"
"log/slog"
"net"
"os"

queryv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/query/v1"
"google.golang.org/grpc"
"google.golang.org/grpc/credentials"
"google.golang.org/grpc/reflection"
)

// GRPCServer manages the gRPC server lifecycle.
type GRPCServer struct {
server  *grpc.Server
handler queryv1.QueryServiceServer
port    int
}

// NewGRPCServer creates a mTLS gRPC server with logging and recovery interceptors.
// certFile, keyFile, caFile specify the server TLS certificate, key, and CA for mTLS.
func NewGRPCServer(
handler queryv1.QueryServiceServer,
port int,
certFile, keyFile, caFile string,
) (*GRPCServer, error) {
creds, err := buildServerCredentials(certFile, keyFile, caFile)
if err != nil {
return nil, fmt.Errorf("server.NewGRPCServer: build TLS credentials: %w", err)
}

srv := grpc.NewServer(
grpc.Creds(creds),
grpc.ChainUnaryInterceptor(
loggingInterceptor,
recoveryInterceptor,
),
)

queryv1.RegisterQueryServiceServer(srv, handler)
reflection.Register(srv)

return &GRPCServer{server: srv, handler: handler, port: port}, nil
}

// Start listens on the configured port and blocks until Serve returns.
func (s *GRPCServer) Start() error {
lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
if err != nil {
return fmt.Errorf("server.GRPCServer.Start: listen: %w", err)
}
slog.Info("gRPC server listening", "port", s.port)
if err := s.server.Serve(lis); err != nil {
return fmt.Errorf("server.GRPCServer.Start: serve: %w", err)
}
return nil
}

// GracefulStop initiates a graceful shutdown, waiting for in-flight RPCs to complete.
func (s *GRPCServer) GracefulStop() {
slog.Info("gRPC server initiating graceful shutdown")
s.server.GracefulStop()
}

// buildServerCredentials loads the TLS certificate, key, and CA for mTLS.
func buildServerCredentials(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
cert, err := tls.LoadX509KeyPair(certFile, keyFile)
if err != nil {
return nil, fmt.Errorf("buildServerCredentials: load key pair: %w", err)
}

caCert, err := os.ReadFile(caFile)
if err != nil {
return nil, fmt.Errorf("buildServerCredentials: read CA: %w", err)
}
caPool := x509.NewCertPool()
if !caPool.AppendCertsFromPEM(caCert) {
return nil, fmt.Errorf("buildServerCredentials: parse CA cert")
}

tlsCfg := &tls.Config{
Certificates: []tls.Certificate{cert},
ClientAuth:   tls.RequireAndVerifyClientCert,
ClientCAs:    caPool,
MinVersion:   tls.VersionTLS12,
}

return credentials.NewTLS(tlsCfg), nil
}

// loggingInterceptor logs every incoming RPC at INFO level.
// It MUST NOT log request/response bodies (may contain classified data).
func loggingInterceptor(
ctx context.Context,
req interface{},
info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler,
) (interface{}, error) {
slog.InfoContext(ctx, "rpc received", "method", info.FullMethod)
resp, err := handler(ctx, req)
if err != nil {
slog.ErrorContext(ctx, "rpc failed",
"method", info.FullMethod,
"error", err)
}
return resp, err
}

// recoveryInterceptor catches panics from handlers and converts them to gRPC INTERNAL errors.
// Using panic in production code is prohibited (see coding standards), but this interceptor
// guards against unexpected panics in third-party libraries.
func recoveryInterceptor(
ctx context.Context,
req interface{},
info *grpc.UnaryServerInfo,
handler grpc.UnaryHandler,
) (resp interface{}, err error) {
defer func() {
if r := recover(); r != nil {
slog.ErrorContext(ctx, "rpc panic recovered",
"method", info.FullMethod,
"panic", r)
err = fmt.Errorf("internal server error")
}
}()
return handler(ctx, req)
}
