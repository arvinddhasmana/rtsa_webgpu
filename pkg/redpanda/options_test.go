// CLASSIFICATION: UNCLASSIFIED
package redpanda_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/redpanda"
)

func TestConnectionOptions_BuildKgoOpts(t *testing.T) {
	t.Run("basic options", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers:  []string{"localhost:9092"},
			ClientID: "test-client",
		}
		kgoOpts, err := opts.BuildKgoOpts()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(kgoOpts) == 0 {
			t.Error("expected kgoOpts to be populated")
		}
	})

	t.Run("sasl scram 256", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers: []string{"localhost:9092"},
			SASL: &redpanda.SASLConfig{
				Mechanism: "SCRAM-SHA-256",
				Username:  "user",
				Password:  "pass",
			},
		}
		_, err := opts.BuildKgoOpts()
		if err != nil {
			t.Errorf("unexpected error for SASL SCRAM-SHA-256: %v", err)
		}
	})

	t.Run("sasl scram 512", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers: []string{"localhost:9092"},
			SASL: &redpanda.SASLConfig{
				Mechanism: "SCRAM-SHA-512",
				Username:  "user",
				Password:  "pass",
			},
		}
		_, err := opts.BuildKgoOpts()
		if err != nil {
			t.Errorf("unexpected error for SASL SCRAM-SHA-512: %v", err)
		}
	})

	t.Run("sasl empty mechanism", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers: []string{"localhost:9092"},
			SASL: &redpanda.SASLConfig{
				Username: "user",
			},
		}
		_, err := opts.BuildKgoOpts()
		if err == nil {
			t.Error("expected error for empty SASL mechanism")
		}
	})

	t.Run("sasl invalid mechanism", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers: []string{"localhost:9092"},
			SASL: &redpanda.SASLConfig{
				Mechanism: "INVALID",
			},
		}
		_, err := opts.BuildKgoOpts()
		if err == nil {
			t.Error("expected error for invalid SASL mechanism")
		}
	})

	t.Run("tls non-existent ca cert", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers:    []string{"localhost:9092"},
			TLSEnabled: true,
			TLSCAFile:  "nonexistent-ca.crt",
		}
		_, err := opts.BuildKgoOpts()
		if err == nil {
			t.Error("expected error for non-existent CA file")
		}
	})

	t.Run("tls non-existent cert/key files", func(t *testing.T) {
		tmpDir := t.TempDir()
		caFile := filepath.Join(tmpDir, "ca.crt")
		_ = os.WriteFile(caFile, []byte("fake-cert"), 0644)

		opts := &redpanda.ConnectionOptions{
			Brokers:     []string{"localhost:9092"},
			TLSEnabled:  true,
			TLSCAFile:   caFile,
			TLSCertFile: "nonexistent-cert.crt",
			TLSKeyFile:  "nonexistent-key.key",
		}
		_, err := opts.BuildKgoOpts()
		if err == nil {
			t.Error("expected error for non-existent client cert/key files")
		}
	})

	t.Run("tls disabled", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers:    []string{"localhost:9092"},
			TLSEnabled: false,
		}
		kgoOpts, err := opts.BuildKgoOpts()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(kgoOpts) == 0 {
			t.Error("expected kgoOpts to be populated")
		}
	})

	t.Run("tls only ca", func(t *testing.T) {
		tmpDir := t.TempDir()
		caFile := filepath.Join(tmpDir, "ca.crt")
		certContent := "-----BEGIN CERTIFICATE-----\nMIICVjCCAb6gAwIBAgIKYm8=\n-----END CERTIFICATE-----"
		_ = os.WriteFile(caFile, []byte(certContent), 0644)

		opts := &redpanda.ConnectionOptions{
			Brokers:    []string{"localhost:9092"},
			TLSEnabled: true,
			TLSCAFile:  caFile,
		}
		_, err := opts.BuildKgoOpts()
		if err != nil {
			t.Errorf("unexpected error for TLS CA: %v", err)
		}
	})

	t.Run("tls full success with fake files", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers:    []string{"localhost:9092"},
			TLSEnabled: true,
		}
		_, err := opts.BuildKgoOpts()
		if err != nil {
			t.Errorf("unexpected error for TLS enabled without files: %v", err)
		}
	})

	t.Run("tls cert without key", func(t *testing.T) {
		opts := &redpanda.ConnectionOptions{
			Brokers:     []string{"localhost:9092"},
			TLSEnabled:  true,
			TLSCertFile: "fake.crt",
		}
		_, err := opts.BuildKgoOpts()
		if err != nil {
			t.Errorf("unexpected error for TLS cert without key: %v", err)
		}
	})
}

func TestBuildKgoOpts_NoSASL_NoTLS(t *testing.T) {
	opts := &redpanda.ConnectionOptions{
		Brokers: []string{"localhost:9092"},
	}
	kgoOpts, err := opts.BuildKgoOpts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kgoOpts) != 2 {
		t.Errorf("expected 2 opts, got %d", len(kgoOpts))
	}
}

func TestBuildSASLOpt_Errors(t *testing.T) {
	opts := &redpanda.ConnectionOptions{}
	_, err := opts.BuildKgoOpts()
	if err != nil {
		t.Errorf("unexpected error for nil SASL: %v", err)
	}
}
