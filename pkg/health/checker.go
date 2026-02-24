// CLASSIFICATION: UNCLASSIFIED
package health

import "sync"

// Status represents the health status of a component.
type Status int

const (
StatusUnknown    Status = 0
StatusServing    Status = 1
StatusNotServing Status = 2
)

// Checker maintains the health status of registered components.
type Checker struct {
mu     sync.RWMutex
checks map[string]Status
}

// NewChecker creates a new Checker with no registered components.
func NewChecker() *Checker {
return &Checker{
checks: make(map[string]Status),
}
}

// Register adds a component to health tracking with initial StatusUnknown.
func (c *Checker) Register(name string) {
c.mu.Lock()
defer c.mu.Unlock()
c.checks[name] = StatusUnknown
}

// SetStatus updates the health status of a named component.
func (c *Checker) SetStatus(name string, status Status) {
c.mu.Lock()
defer c.mu.Unlock()
c.checks[name] = status
}

// Overall returns StatusServing if ALL components are Serving,
// StatusNotServing if ANY component is NotServing,
// StatusUnknown if ANY component is Unknown and none are NotServing.
func (c *Checker) Overall() Status {
c.mu.RLock()
defer c.mu.RUnlock()

if len(c.checks) == 0 {
return StatusUnknown
}

hasUnknown := false
for _, s := range c.checks {
if s == StatusNotServing {
return StatusNotServing
}
if s == StatusUnknown {
hasUnknown = true
}
}
if hasUnknown {
return StatusUnknown
}
return StatusServing
}

// ComponentStatus returns the status of a single component.
func (c *Checker) ComponentStatus(name string) Status {
c.mu.RLock()
defer c.mu.RUnlock()
if s, ok := c.checks[name]; ok {
return s
}
return StatusUnknown
}
