// CLASSIFICATION: UNCLASSIFIED
package redpanda

import (
"crypto/tls"
"crypto/x509"
"fmt"
"os"

"github.com/twmb/franz-go/pkg/kgo"
"github.com/twmb/franz-go/pkg/sasl/scram"
)

// ConnectionOptions configures the Redpanda client connection.
type ConnectionOptions struct {
Brokers     []string
TLSEnabled  bool
TLSCertFile string
TLSKeyFile  string
TLSCAFile   string
ClientID    string
SASL        *SASLConfig
}

// SASLConfig holds SASL authentication credentials.
type SASLConfig struct {
Mechanism string // "SCRAM-SHA-256" or "SCRAM-SHA-512"
Username  string
Password  string
}

// BuildKgoOpts converts ConnectionOptions into franz-go kgo.Opt slice.
func (o *ConnectionOptions) BuildKgoOpts() ([]kgo.Opt, error) {
opts := []kgo.Opt{
kgo.SeedBrokers(o.Brokers...),
kgo.ClientID(o.ClientID),
}

if o.TLSEnabled {
tlsCfg, err := o.buildTLSConfig()
if err != nil {
return nil, fmt.Errorf("redpanda: build TLS config: %w", err)
}
opts = append(opts, kgo.DialTLSConfig(tlsCfg))
}

if o.SASL != nil {
saslOpt, err := o.buildSASLOpt()
if err != nil {
return nil, fmt.Errorf("redpanda: build SASL config: %w", err)
}
opts = append(opts, saslOpt)
}

return opts, nil
}

func (o *ConnectionOptions) buildTLSConfig() (*tls.Config, error) {
tlsCfg := &tls.Config{
MinVersion: tls.VersionTLS12,
}

if o.TLSCAFile != "" {
caCert, err := os.ReadFile(o.TLSCAFile)
if err != nil {
return nil, fmt.Errorf("redpanda: read CA cert: %w", err)
}
pool := x509.NewCertPool()
pool.AppendCertsFromPEM(caCert)
tlsCfg.RootCAs = pool
}

if o.TLSCertFile != "" && o.TLSKeyFile != "" {
cert, err := tls.LoadX509KeyPair(o.TLSCertFile, o.TLSKeyFile)
if err != nil {
return nil, fmt.Errorf("redpanda: load client cert/key: %w", err)
}
tlsCfg.Certificates = []tls.Certificate{cert}
}

return tlsCfg, nil
}

func (o *ConnectionOptions) buildSASLOpt() (kgo.Opt, error) {
if o.SASL == nil {
return nil, fmt.Errorf("redpanda: SASL config is nil")
}
switch o.SASL.Mechanism {
case "SCRAM-SHA-256":
return kgo.SASL(scram.Auth{
User: o.SASL.Username,
Pass: o.SASL.Password,
}.AsSha256Mechanism()), nil
case "SCRAM-SHA-512":
return kgo.SASL(scram.Auth{
User: o.SASL.Username,
Pass: o.SASL.Password,
}.AsSha512Mechanism()), nil
default:
return nil, fmt.Errorf("redpanda: unsupported SASL mechanism: %s", o.SASL.Mechanism)
}
}
