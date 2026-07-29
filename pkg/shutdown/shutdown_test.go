// CLASSIFICATION: UNCLASSIFIED
package shutdown_test

import (
	"context"
	"log/slog"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/arvinddhasmana/rtsa_webgpu/pkg/shutdown"
)

func TestWaitForSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	done := make(chan struct{})
	go func() {
		shutdown.WaitForSignal(cancel, logger)
		close(done)
	}()

	// Send SIGINT to ourselves
	time.Sleep(50 * time.Millisecond) // Give the goroutine time to register the notify
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	if err := process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send signal: %v", err)
	}

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Error("timeout waiting for WaitForSignal to return")
	}

	if err := ctx.Err(); err != context.Canceled {
		t.Errorf("expected context to be canceled, got %v", err)
	}
}
