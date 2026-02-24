// CLASSIFICATION: UNCLASSIFIED
package shutdown_test

import (
"context"
"errors"
"testing"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/shutdown"
"go.uber.org/zap"
)

func TestManager_HooksLIFO(t *testing.T) {
logger, _ := zap.NewDevelopment()
m := shutdown.NewManager(logger, 5*time.Second)

order := []string{}
m.Register("first", func(ctx context.Context) error {
order = append(order, "first")
return nil
})
m.Register("second", func(ctx context.Context) error {
order = append(order, "second")
return nil
})
m.Register("third", func(ctx context.Context) error {
order = append(order, "third")
return nil
})

// Trigger and wait
go func() {
time.Sleep(10 * time.Millisecond)
m.Trigger()
}()

if err := m.Wait(); err != nil {
t.Fatalf("unexpected error: %v", err)
}

if len(order) != 3 {
t.Fatalf("expected 3 hooks, got %d", len(order))
}
// LIFO: third, second, first
if order[0] != "third" || order[1] != "second" || order[2] != "first" {
t.Errorf("unexpected order: %v", order)
}
}

func TestManager_HookError_AllCalled(t *testing.T) {
logger, _ := zap.NewDevelopment()
m := shutdown.NewManager(logger, 5*time.Second)

called := map[string]bool{}
m.Register("a", func(ctx context.Context) error {
called["a"] = true
return errors.New("hook a failed")
})
m.Register("b", func(ctx context.Context) error {
called["b"] = true
return nil
})

go func() {
time.Sleep(10 * time.Millisecond)
m.Trigger()
}()

err := m.Wait()
if err == nil {
t.Error("expected error from hook a")
}
if !called["a"] || !called["b"] {
t.Errorf("expected all hooks called, got: %v", called)
}
}

func TestManager_ProgrammaticTrigger(t *testing.T) {
logger, _ := zap.NewDevelopment()
m := shutdown.NewManager(logger, 5*time.Second)

done := make(chan error, 1)
go func() {
done <- m.Wait()
}()

m.Trigger()

select {
case err := <-done:
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
case <-time.After(2 * time.Second):
t.Fatal("timeout waiting for shutdown")
}
}
