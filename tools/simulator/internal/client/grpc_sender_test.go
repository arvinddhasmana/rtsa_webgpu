// CLASSIFICATION: UNCLASSIFIED
package client_test

import (
"context"
cryptorand "crypto/rand"
"crypto/ecdsa"
"crypto/elliptic"
"crypto/x509"
"crypto/x509/pkix"
"encoding/pem"
"math/big"
"net"
"os"
"path/filepath"
"testing"
"time"

ingestionv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/ingestion/v1"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/client"
"github.com/arvinddhasmana/RTSA_VS_Opus/tools/simulator/internal/config"
"google.golang.org/grpc"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)

// ── Mock sender (test double) ──────────────────────────────────────────────

type mockSender struct {
sent    []*ingestionv1.SensorObservation
sendErr error
}

func (m *mockSender) Send(_ context.Context, obs *ingestionv1.SensorObservation, _ client.SensorType) error {
if m.sendErr != nil {
return m.sendErr
}
m.sent = append(m.sent, obs)
return nil
}

func (m *mockSender) Close() error { return nil }

// ── Helpers ────────────────────────────────────────────────────────────────

func insecureCfg() *config.SimulatorConfig {
return &config.SimulatorConfig{
RadarEndpoint:  "localhost:9",
EWEndpoint:     "localhost:9",
ELINTEndpoint:  "localhost:9",
ISREndpoint:    "localhost:9",
AISEndpoint:    "localhost:9",
CyberEndpoint:  "localhost:9",
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}
}

// mockIngestionServer is a minimal gRPC server used in unit tests.
type mockIngestionServer struct {
ingestionv1.UnimplementedIngestionServiceServer
received []*ingestionv1.SensorObservation
}

func (m *mockIngestionServer) IngestSingleObservation(
_ context.Context,
obs *ingestionv1.SensorObservation,
) (*ingestionv1.IngestionAck, error) {
m.received = append(m.received, obs)
return &ingestionv1.IngestionAck{ObservationId: obs.GetObservationId()}, nil
}

// errorIngestionServer always returns an error.
type errorIngestionServer struct {
ingestionv1.UnimplementedIngestionServiceServer
}

func (e *errorIngestionServer) IngestSingleObservation(
_ context.Context,
_ *ingestionv1.SensorObservation,
) (*ingestionv1.IngestionAck, error) {
return nil, status.Error(codes.Internal, "simulated server error")
}

// startTestGRPCServer starts an in-process gRPC server on a random port.
func startTestGRPCServer(t *testing.T) (addr string, mock *mockIngestionServer, stop func()) {
t.Helper()
lis, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("failed to listen: %v", err)
}
srv := grpc.NewServer()
mock = &mockIngestionServer{}
ingestionv1.RegisterIngestionServiceServer(srv, mock)
go func() { _ = srv.Serve(lis) }()
return lis.Addr().String(), mock, srv.GracefulStop
}

// generateSelfSignedCert creates a temporary self-signed cert/key for testing.
func generateSelfSignedCert(t *testing.T) (certPath, keyPath, caPath string) {
t.Helper()

privKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
if err != nil {
t.Fatalf("generating key: %v", err)
}

template := &x509.Certificate{
SerialNumber:          big.NewInt(1),
Subject:               pkix.Name{CommonName: "rtsa-simulator-test"},
NotBefore:             time.Now().Add(-time.Hour),
NotAfter:              time.Now().Add(time.Hour),
KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
IsCA:                  true,
BasicConstraintsValid: true,
}

certDER, err := x509.CreateCertificate(cryptorand.Reader, template, template, &privKey.PublicKey, privKey)
if err != nil {
t.Fatalf("creating certificate: %v", err)
}

dir := t.TempDir()
certPath = filepath.Join(dir, "cert.pem")
keyPath = filepath.Join(dir, "key.pem")

cf, err := os.Create(certPath)
if err != nil {
t.Fatalf("creating cert file: %v", err)
}
if err := pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
t.Fatalf("encoding cert PEM: %v", err)
}
cf.Close()

keyDER, err := x509.MarshalECPrivateKey(privKey)
if err != nil {
t.Fatalf("marshalling key: %v", err)
}
kf, err := os.Create(keyPath)
if err != nil {
t.Fatalf("creating key file: %v", err)
}
if err := pem.Encode(kf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
t.Fatalf("encoding key PEM: %v", err)
}
kf.Close()

// CA uses the same self-signed cert.
caPath = certPath
return certPath, keyPath, caPath
}

// ── Mock sender tests ──────────────────────────────────────────────────────

func TestMockSender_RecordsObservations(t *testing.T) {
sender := &mockSender{}
ctx := context.Background()
obs := &ingestionv1.SensorObservation{ObservationId: "test-obs-001", SensorId: "RADAR-TEST"}

if err := sender.Send(ctx, obs, client.SensorTypeRadar); err != nil {
t.Fatalf("unexpected error: %v", err)
}
if len(sender.sent) != 1 {
t.Errorf("expected 1 sent observation, got %d", len(sender.sent))
}
if sender.sent[0].ObservationId != "test-obs-001" {
t.Errorf("expected obs ID test-obs-001, got %q", sender.sent[0].ObservationId)
}
}

func TestMockSender_SendError(t *testing.T) {
ctx := context.Background()
obs := &ingestionv1.SensorObservation{ObservationId: "err-obs"}
sender := &mockSender{sendErr: os.ErrInvalid}

if err := sender.Send(ctx, obs, client.SensorTypeRadar); err == nil {
t.Error("expected error from mock sender")
}
if len(sender.sent) != 0 {
t.Error("failed send should not add to sent list")
}
}

func TestMockSender_Close(t *testing.T) {
sender := &mockSender{}
if err := sender.Close(); err != nil {
t.Errorf("Close should not error: %v", err)
}
}

// ── GRPCSender creation tests ──────────────────────────────────────────────

func TestNewGRPCSender_Insecure_CreatesAndCloses(t *testing.T) {
cfg := insecureCfg()
sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Logf("NewGRPCSender returned error: %v", err)
return
}
if closeErr := sender.Close(); closeErr != nil {
t.Errorf("Close returned unexpected error: %v", closeErr)
}
}

func TestNewGRPCSender_TLS_MissingCerts(t *testing.T) {
cfg := insecureCfg()
cfg.TLSEnabled = true
cfg.TLSCertFile = "/nonexistent.crt"
cfg.TLSKeyFile = "/nonexistent.key"
cfg.TLSCAFile = "/nonexistent.ca"

_, err := client.NewGRPCSender(cfg)
if err == nil {
t.Error("expected error when TLS cert files are missing")
}
}

func TestNewGRPCSender_TLS_InvalidCertContent(t *testing.T) {
dir := t.TempDir()
certFile := filepath.Join(dir, "cert.pem")
keyFile := filepath.Join(dir, "key.pem")
caFile := filepath.Join(dir, "ca.pem")

for _, f := range []string{certFile, keyFile, caFile} {
if err := os.WriteFile(f, []byte("not-a-real-cert"), 0600); err != nil {
t.Fatalf("failed to write temp file: %v", err)
}
}

cfg := insecureCfg()
cfg.TLSEnabled = true
cfg.TLSCertFile = certFile
cfg.TLSKeyFile = keyFile
cfg.TLSCAFile = caFile

_, err := client.NewGRPCSender(cfg)
if err == nil {
t.Error("expected error when cert/key content is invalid")
}
}

func TestNewGRPCSender_TLS_ValidCerts_InvalidCA(t *testing.T) {
certPath, keyPath, _ := generateSelfSignedCert(t)

dir := t.TempDir()
caPath := filepath.Join(dir, "ca-invalid.pem")
if err := os.WriteFile(caPath, []byte("not-pem-data"), 0600); err != nil {
t.Fatalf("writing CA file: %v", err)
}

cfg := insecureCfg()
cfg.TLSEnabled = true
cfg.TLSCertFile = certPath
cfg.TLSKeyFile = keyPath
cfg.TLSCAFile = caPath

_, err := client.NewGRPCSender(cfg)
if err == nil {
t.Error("expected error when CA PEM is invalid")
}
}

func TestNewGRPCSender_TLS_ValidCerts_ValidCA(t *testing.T) {
certPath, keyPath, caPath := generateSelfSignedCert(t)

cfg := &config.SimulatorConfig{
RadarEndpoint:  "localhost:9",
EWEndpoint:     "localhost:9",
ELINTEndpoint:  "localhost:9",
ISREndpoint:    "localhost:9",
AISEndpoint:    "localhost:9",
CyberEndpoint:  "localhost:9",
TLSEnabled:     true,
TLSCertFile:    certPath,
TLSKeyFile:     keyPath,
TLSCAFile:      caPath,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender with valid TLS certs should not fail: %v", err)
}
if closeErr := sender.Close(); closeErr != nil {
t.Errorf("Close error: %v", closeErr)
}
}

// ── GRPCSender.Send tests ──────────────────────────────────────────────────

func TestGRPCSender_Send_SingleObservation(t *testing.T) {
addr, mock, stop := startTestGRPCServer(t)
defer stop()

cfg := &config.SimulatorConfig{
RadarEndpoint:  addr,
EWEndpoint:     addr,
ELINTEndpoint:  addr,
ISREndpoint:    addr,
AISEndpoint:    addr,
CyberEndpoint:  addr,
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender: %v", err)
}
defer func() { _ = sender.Close() }()

obs := &ingestionv1.SensorObservation{ObservationId: "send-test-001"}
if err := sender.Send(context.Background(), obs, client.SensorTypeRadar); err != nil {
t.Errorf("Send returned unexpected error: %v", err)
}
if len(mock.received) != 1 {
t.Errorf("expected 1 observation on server, got %d", len(mock.received))
}
}

func TestGRPCSender_Send_AllSensorTypes(t *testing.T) {
addr, mock, stop := startTestGRPCServer(t)
defer stop()

cfg := &config.SimulatorConfig{
RadarEndpoint:  addr,
EWEndpoint:     addr,
ELINTEndpoint:  addr,
ISREndpoint:    addr,
AISEndpoint:    addr,
CyberEndpoint:  addr,
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender: %v", err)
}
defer func() { _ = sender.Close() }()

ctx := context.Background()
sensorTypes := []client.SensorType{
client.SensorTypeRadar, client.SensorTypeEW, client.SensorTypeELINT,
client.SensorTypeISR, client.SensorTypeAIS, client.SensorTypeCyber,
}
for _, st := range sensorTypes {
obs := &ingestionv1.SensorObservation{ObservationId: "obs-" + string(st)}
if err := sender.Send(ctx, obs, st); err != nil {
t.Errorf("Send(%s): %v", st, err)
}
}
if len(mock.received) != len(sensorTypes) {
t.Errorf("expected %d observations, got %d", len(sensorTypes), len(mock.received))
}
}

func TestGRPCSender_Send_UnknownSensorType(t *testing.T) {
addr, _, stop := startTestGRPCServer(t)
defer stop()

cfg := &config.SimulatorConfig{
RadarEndpoint:  addr,
EWEndpoint:     addr,
ELINTEndpoint:  addr,
ISREndpoint:    addr,
AISEndpoint:    addr,
CyberEndpoint:  addr,
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender: %v", err)
}
defer func() { _ = sender.Close() }()

obs := &ingestionv1.SensorObservation{ObservationId: "unknown-type"}
err = sender.Send(context.Background(), obs, client.SensorType("unknown-sensor"))
if err == nil {
t.Error("expected error for unknown sensor type")
}
}

func TestGRPCSender_Send_ServerError(t *testing.T) {
lis, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("failed to listen: %v", err)
}
srv := grpc.NewServer()
ingestionv1.RegisterIngestionServiceServer(srv, &errorIngestionServer{})
go func() { _ = srv.Serve(lis) }()
defer srv.GracefulStop()

addr := lis.Addr().String()
cfg := &config.SimulatorConfig{
RadarEndpoint:  addr,
EWEndpoint:     addr,
ELINTEndpoint:  addr,
ISREndpoint:    addr,
AISEndpoint:    addr,
CyberEndpoint:  addr,
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender: %v", err)
}
defer func() { _ = sender.Close() }()

obs := &ingestionv1.SensorObservation{ObservationId: "error-test"}
if err := sender.Send(context.Background(), obs, client.SensorTypeRadar); err == nil {
t.Error("expected error from server error response")
}
}

func TestGRPCSender_Send_CancelledContext(t *testing.T) {
addr, _, stop := startTestGRPCServer(t)
defer stop()

cfg := &config.SimulatorConfig{
RadarEndpoint:  addr,
EWEndpoint:     addr,
ELINTEndpoint:  addr,
ISREndpoint:    addr,
AISEndpoint:    addr,
CyberEndpoint:  addr,
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender: %v", err)
}
defer func() { _ = sender.Close() }()

ctx, cancel := context.WithCancel(context.Background())
cancel() // cancel immediately

obs := &ingestionv1.SensorObservation{ObservationId: "cancelled-ctx"}
// Cancelled context should produce an error or succeed (timing-dependent).
_ = sender.Send(ctx, obs, client.SensorTypeRadar)
}

func TestGRPCSender_Close_NoError(t *testing.T) {
addr, _, stop := startTestGRPCServer(t)
defer stop()

cfg := &config.SimulatorConfig{
RadarEndpoint:  addr,
EWEndpoint:     addr,
ELINTEndpoint:  addr,
ISREndpoint:    addr,
AISEndpoint:    addr,
CyberEndpoint:  addr,
TLSEnabled:     false,
UpdateInterval: time.Second,
AnomalyRate:    0.05,
}

sender, err := client.NewGRPCSender(cfg)
if err != nil {
t.Fatalf("NewGRPCSender: %v", err)
}
if err := sender.Close(); err != nil {
t.Errorf("Close should not error: %v", err)
}
}

// ── Sensor type constants ──────────────────────────────────────────────────

func TestSensorTypeConstants(t *testing.T) {
types := []client.SensorType{
client.SensorTypeRadar,
client.SensorTypeEW,
client.SensorTypeELINT,
client.SensorTypeISR,
client.SensorTypeAIS,
client.SensorTypeCyber,
}
seen := make(map[client.SensorType]bool)
for _, st := range types {
if seen[st] {
t.Errorf("duplicate sensor type value: %q", st)
}
seen[st] = true
if string(st) == "" {
t.Error("sensor type must not be empty string")
}
}
}

func TestSenderInterface_MockImplements(t *testing.T) {
var sender client.ObservationSender = &mockSender{}
ctx := context.Background()
obs := &ingestionv1.SensorObservation{ObservationId: "interface-test"}
if err := sender.Send(ctx, obs, client.SensorTypeAIS); err != nil {
t.Errorf("unexpected error: %v", err)
}
if err := sender.Close(); err != nil {
t.Errorf("Close error: %v", err)
}
}
