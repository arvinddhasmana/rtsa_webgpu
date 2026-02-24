// CLASSIFICATION: UNCLASSIFIED
package shutdown

import (
"context"
"fmt"
"os"
"os/signal"
"sync"
"syscall"
"time"

"go.uber.org/zap"
)

// ShutdownFunc is a function called during shutdown.
type ShutdownFunc func(ctx context.Context) error

// Manager orchestrates graceful shutdown of service components.
// Shutdown order is LIFO (last registered = first shutdown).
type Manager struct {
logger  *zap.Logger
timeout time.Duration
hooks   []namedHook
mu      sync.Mutex
trigger chan struct{}
once    sync.Once
}

type namedHook struct {
name string
fn   ShutdownFunc
}

// NewManager creates a shutdown manager with the given timeout.
// Listens for SIGINT and SIGTERM.
func NewManager(logger *zap.Logger, timeout time.Duration) *Manager {
return &Manager{
logger:  logger,
timeout: timeout,
trigger: make(chan struct{}),
}
}

// Register adds a shutdown hook. Hooks run in reverse order (LIFO).
func (m *Manager) Register(name string, fn ShutdownFunc) {
m.mu.Lock()
defer m.mu.Unlock()
m.hooks = append(m.hooks, namedHook{name: name, fn: fn})
}

// Wait blocks until a shutdown signal is received, then executes all hooks.
func (m *Manager) Wait() error {
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
defer signal.Stop(sigCh)

select {
case sig := <-sigCh:
m.logger.Info("shutdown signal received", zap.String("signal", sig.String()))
case <-m.trigger:
m.logger.Info("programmatic shutdown triggered")
}

return m.runHooks()
}

// Trigger programmatically initiates shutdown.
func (m *Manager) Trigger() {
m.once.Do(func() {
close(m.trigger)
})
}

func (m *Manager) runHooks() error {
m.mu.Lock()
hooks := make([]namedHook, len(m.hooks))
copy(hooks, m.hooks)
m.mu.Unlock()

ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
defer cancel()

var firstErr error
// Execute in LIFO order
for i := len(hooks) - 1; i >= 0; i-- {
h := hooks[i]
start := time.Now()
m.logger.Info("running shutdown hook", zap.String("hook", h.name))

if err := h.fn(ctx); err != nil {
m.logger.Error("shutdown hook failed",
zap.String("hook", h.name),
zap.Duration("duration", time.Since(start)),
zap.Error(err))
if firstErr == nil {
firstErr = fmt.Errorf("shutdown: hook %s failed: %w", h.name, err)
}
} else {
m.logger.Info("shutdown hook completed",
zap.String("hook", h.name),
zap.Duration("duration", time.Since(start)))
}
}
return firstErr
}
