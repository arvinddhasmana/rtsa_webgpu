// CLASSIFICATION: UNCLASSIFIED

package domain

import (
"sync"
"time"
)

// slidingWindow tracks request timestamps within a rolling time window.
type slidingWindow struct {
timestamps []time.Time
}

// RateLimiter enforces per-operator rate limits using a sliding window.
// Default window is 60 seconds (1 minute).
// All methods are safe for concurrent use.
type RateLimiter struct {
mu           sync.Mutex
windows      map[string]*slidingWindow
maxPerMinute int
windowSize   time.Duration
}

// NewRateLimiter constructs a RateLimiter with the specified per-minute limit.
func NewRateLimiter(maxPerMinute int) *RateLimiter {
return &RateLimiter{
windows:      make(map[string]*slidingWindow),
maxPerMinute: maxPerMinute,
windowSize:   time.Minute,
}
}

// Allow returns true if the operator has not exceeded the rate limit within
// the sliding 60-second window. It records the request timestamp on allow.
func (rl *RateLimiter) Allow(operatorID string) bool {
rl.mu.Lock()
defer rl.mu.Unlock()

now := time.Now()
cutoff := now.Add(-rl.windowSize)

w, ok := rl.windows[operatorID]
if !ok {
w = &slidingWindow{}
rl.windows[operatorID] = w
}

// Evict timestamps older than the window.
valid := w.timestamps[:0]
for _, t := range w.timestamps {
if t.After(cutoff) {
valid = append(valid, t)
}
}
w.timestamps = valid

if len(w.timestamps) >= rl.maxPerMinute {
return false
}

w.timestamps = append(w.timestamps, now)
return true
}
