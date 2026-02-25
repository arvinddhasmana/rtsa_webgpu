// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"sync"
"testing"
"time"
)

// T18: Rate limit: 10 requests in 1 min → all allowed
func TestRateLimiter_T18_TenAllowed(t *testing.T) {
rl := NewRateLimiter(10)
for i := 0; i < 10; i++ {
if !rl.Allow("op-001") {
t.Errorf("T18: request %d should be allowed", i+1)
}
}
}

// T19: Rate limit: 11th request in 1 min → rejected
func TestRateLimiter_T19_EleventhRejected(t *testing.T) {
rl := NewRateLimiter(10)
for i := 0; i < 10; i++ {
rl.Allow("op-002")
}
if rl.Allow("op-002") {
t.Error("T19: 11th request should be rejected")
}
}

// T20: Rate limit window expires → next request allowed
func TestRateLimiter_T20_WindowExpires(t *testing.T) {
rl := NewRateLimiter(2)
// Inject two timestamps that are older than the window.
rl.mu.Lock()
rl.windows["op-003"] = &slidingWindow{
timestamps: []time.Time{
time.Now().Add(-2 * time.Minute),
time.Now().Add(-2 * time.Minute),
},
}
rl.mu.Unlock()

// Both old entries should be evicted → new request allowed.
if !rl.Allow("op-003") {
t.Error("T20: request should be allowed after window expiry")
}
}

// Different operators have independent windows.
func TestRateLimiter_IsolatedPerOperator(t *testing.T) {
rl := NewRateLimiter(3)
for i := 0; i < 3; i++ {
rl.Allow("op-A")
}
// op-A is exhausted; op-B should still be free.
if !rl.Allow("op-B") {
t.Error("op-B should have independent window from op-A")
}
// op-A should now be blocked.
if rl.Allow("op-A") {
t.Error("op-A should be rate-limited")
}
}

// Concurrent Allow calls must not race.
func TestRateLimiter_ConcurrentSafe(t *testing.T) {
rl := NewRateLimiter(100)
var wg sync.WaitGroup
for i := 0; i < 50; i++ {
wg.Add(1)
go func() {
defer wg.Done()
rl.Allow("op-concurrent")
}()
}
wg.Wait()
}

func TestRateLimiter_ZeroInitially(t *testing.T) {
rl := NewRateLimiter(1)
if !rl.Allow("op-fresh") {
t.Error("first request for fresh operator should always be allowed")
}
}
