// CLASSIFICATION: UNCLASSIFIED
package client

import (
"context"
"crypto/tls"
"crypto/x509"
"fmt"
"os"
"time"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/config"
"google.golang.org/grpc"
"google.golang.org/grpc/credentials"
"google.golang.org/grpc/credentials/insecure"
)

// SensorType identifies which ingestion service to send to.
type SensorType string

const (
SensorTypeRadar SensorType = "radar"
SensorTypeEW    SensorType = "ew"
SensorTypeELINT SensorType = "elint"
SensorTypeISR   SensorType = "isr"
SensorTypeAIS   SensorType = "ais"
SensorTypeCyber SensorType = "cyber"
)

// ObservationSender defines the interface for sending observations to ingestion services.
type ObservationSender interface {
Send(ctx context.Context, obs *ingestionv1.SensorObservation, sensorType SensorType) error
Close() error
}

// GRPCSender manages gRPC connections to all ingestion services.
type GRPCSender struct {
conns   map[SensorType]*grpc.ClientConn
clients map[SensorType]ingestionv1.IngestionServiceClient
cfg     *config.SimulatorConfig
}

// NewGRPCSender creates a GRPCSender and establishes connections to all endpoints.
// Each connection is established lazily — the first Send call will trigger dialling.
func NewGRPCSender(cfg *config.SimulatorConfig) (*GRPCSender, error) {
s := &GRPCSender{
conns:   make(map[SensorType]*grpc.ClientConn),
clients: make(map[SensorType]ingestionv1.IngestionServiceClient),
cfg:     cfg,
}

endpoints := map[SensorType]string{
SensorTypeRadar: cfg.RadarEndpoint,
SensorTypeEW:    cfg.EWEndpoint,
SensorTypeELINT: cfg.ELINTEndpoint,
SensorTypeISR:   cfg.ISREndpoint,
SensorTypeAIS:   cfg.AISEndpoint,
SensorTypeCyber: cfg.CyberEndpoint,
}

dialOpts, err := buildDialOpts(cfg)
if err != nil {
return nil, fmt.Errorf("building dial options: %w", err)
}

for st, addr := range endpoints {
conn, err := grpc.NewClient(addr, dialOpts...)
if err != nil {
// Close already-opened connections on error.
_ = s.Close()
return nil, fmt.Errorf("dialling %s at %s: %w", st, addr, err)
}
s.conns[st] = conn
s.clients[st] = ingestionv1.NewIngestionServiceClient(conn)
}

return s, nil
}

// Send submits a single observation to the appropriate ingestion service.
func (s *GRPCSender) Send(ctx context.Context, obs *ingestionv1.SensorObservation, sensorType SensorType) error {
client, ok := s.clients[sensorType]
if !ok {
return fmt.Errorf("no client for sensor type %q", sensorType)
}

ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

_, err := client.IngestSingleObservation(ctx, obs)
if err != nil {
return fmt.Errorf("sending %s observation: %w", sensorType, err)
}
return nil
}

// Close tears down all gRPC connections.
func (s *GRPCSender) Close() error {
var lastErr error
for _, conn := range s.conns {
if err := conn.Close(); err != nil {
lastErr = err
}
}
return lastErr
}

// buildDialOpts returns gRPC dial options appropriate for the TLS configuration.
func buildDialOpts(cfg *config.SimulatorConfig) ([]grpc.DialOption, error) {
if !cfg.TLSEnabled {
return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, nil
}

// Load mTLS certificates.
cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
if err != nil {
return nil, fmt.Errorf("loading TLS cert/key: %w", err)
}
caPEM, err := os.ReadFile(cfg.TLSCAFile)
if err != nil {
return nil, fmt.Errorf("reading CA file: %w", err)
}
pool := x509.NewCertPool()
if !pool.AppendCertsFromPEM(caPEM) {
return nil, fmt.Errorf("parsing CA certificate from %s", cfg.TLSCAFile)
}

tlsCfg := &tls.Config{
Certificates: []tls.Certificate{cert},
RootCAs:      pool,
MinVersion:   tls.VersionTLS13,
}
return []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))}, nil
}
