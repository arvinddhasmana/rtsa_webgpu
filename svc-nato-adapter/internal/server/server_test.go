// CLASSIFICATION: UNCLASSIFIED
package server_test

import (
	"path/filepath"
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/config"
	"github.com/arvinddhasmana/RTSA_VS_Opus/svc-nato-adapter/internal/server"
	"go.uber.org/zap"
)

func TestServer_New(t *testing.T) {
	cfg := &config.Config{
		GRPCAddr:   ":0", // ephemeral port
		TLSEnabled: false,
	}
	logger := zap.NewNop()
	srv, lis, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer srv.Stop()
	defer lis.Close()

	if srv == nil {
		t.Error("expected server to be non-nil")
	}
}

func TestServer_New_TLS_Success(t *testing.T) {
	// Root of workspace is /home/arvind/workspace/RTSA_VS_Opus
	// Certs are in certs/dev/
	basePath := "/home/arvind/workspace/RTSA_VS_Opus"
	cfg := &config.Config{
		GRPCAddr:      ":0",
		TLSEnabled:    true,
		TLSCACert:     filepath.Join(basePath, "certs/dev/ca.crt"),
		TLSServerCert: filepath.Join(basePath, "certs/dev/server.crt"),
		TLSServerKey:  filepath.Join(basePath, "certs/dev/server.key"),
	}
	logger := zap.NewNop()
	srv, lis, err := server.New(cfg, logger)
	if err != nil {
		t.Fatalf("unexpected error with real certs: %v", err)
	}
	defer srv.Stop()
	defer lis.Close()
}

func TestServer_New_TLS_Fail(t *testing.T) {
	cfg := &config.Config{
		TLSEnabled:    true,
		TLSServerCert: "/nonexistent",
		TLSServerKey:  "/nonexistent",
	}
	logger := zap.NewNop()
	_, _, err := server.New(cfg, logger)
	if err == nil {
		t.Error("expected error for nonexistent certificates")
	}
}
