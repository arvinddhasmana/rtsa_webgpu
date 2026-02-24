// CLASSIFICATION: UNCLASSIFIED
package shutdown

import (
"context"
"log/slog"
"os"
"os/signal"
"syscall"
)

// WaitForSignal blocks until SIGINT or SIGTERM is received, then cancels the context.
func WaitForSignal(cancel context.CancelFunc, logger *slog.Logger) {
ch := make(chan os.Signal, 1)
signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
sig := <-ch
logger.Info("shutdown signal received", "signal", sig.String())
cancel()
}
