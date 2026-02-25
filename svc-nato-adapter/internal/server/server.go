// CLASSIFICATION: UNCLASSIFIED
// Package server wires the gRPC server for svc-nato-adapter.
//
// Feature: FEAT-15 NATO Interoperability
// Requirements: CR-NATO-001, CR-SEC-001
package server

import (
"crypto/tls"
"crypto/x509"
"fmt"
"net"
"os"

natov1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/nato/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/config"
"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/handler"
"go.uber.org/zap"
"google.golang.org/grpc"
"google.golang.org/grpc/credentials"
"google.golang.org/grpc/credentials/insecure"
"google.golang.org/grpc/reflection"
)

// New creates and returns a configured gRPC server with the NATO adapter handler registered.
func New(cfg *config.Config, logger *zap.Logger) (*grpc.Server, net.Listener, error) {
var serverOpt grpc.ServerOption
if cfg.TLSEnabled {
creds, err := loadTLSCredentials(cfg)
if err != nil {
return nil, nil, fmt.Errorf("server.New: load TLS: %w", err)
}
serverOpt = creds
} else {
logger.Warn("TLS disabled — mTLS required for production (RTSA_TLS_ENABLED=true)")
serverOpt = grpc.Creds(insecure.NewCredentials())
}

srv := grpc.NewServer(serverOpt)
h := handler.New(logger)
natov1.RegisterNatoAdapterServiceServer(srv, h)
reflection.Register(srv)

lis, err := net.Listen("tcp", cfg.GRPCAddr)
if err != nil {
return nil, nil, fmt.Errorf("server.New: listen %s: %w", cfg.GRPCAddr, err)
}

return srv, lis, nil
}

// loadTLSCredentials loads mTLS credentials from the configured certificate paths.
func loadTLSCredentials(cfg *config.Config) (grpc.ServerOption, error) {
cert, err := tls.LoadX509KeyPair(cfg.TLSServerCert, cfg.TLSServerKey)
if err != nil {
return nil, fmt.Errorf("load server key pair: %w", err)
}

caPEM, err := os.ReadFile(cfg.TLSCACert)
if err != nil {
return nil, fmt.Errorf("read CA cert %q: %w", cfg.TLSCACert, err)
}

certPool := x509.NewCertPool()
if !certPool.AppendCertsFromPEM(caPEM) {
return nil, fmt.Errorf("failed to parse CA cert %q", cfg.TLSCACert)
}

tlsCfg := &tls.Config{
Certificates: []tls.Certificate{cert},
ClientAuth:   tls.RequireAndVerifyClientCert,
ClientCAs:    certPool,
MinVersion:   tls.VersionTLS13,
}

return grpc.Creds(credentials.NewTLS(tlsCfg)), nil
}
