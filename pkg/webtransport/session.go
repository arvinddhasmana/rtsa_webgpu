// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/session.go — Per-client WebTransport session management
//
// The SessionManager tracks active sessions, enforces max-session limits,
// and provides broadcast capability to all connected operators.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §3

package webtransport

import (
"sync"
"sync/atomic"

webtransportgo "github.com/quic-go/webtransport-go"
)

// sessionEntry holds the WebTransport session together with the validated
// operator claims.
type sessionEntry struct {
session *webtransportgo.Session
claims  *SessionClaims
}

// SessionManager manages active WebTransport sessions.
type SessionManager struct {
mu          sync.RWMutex
sessions    map[*webtransportgo.Session]*sessionEntry
maxSessions int
count       atomic.Int64
}

// NewSessionManager creates a SessionManager that enforces a maximum concurrent
// session limit.
func NewSessionManager(maxSessions int) *SessionManager {
return &SessionManager{
sessions:    make(map[*webtransportgo.Session]*sessionEntry),
maxSessions: maxSessions,
}
}

// Register adds a session to the active set.
// Returns false if the maximum session limit has been reached.
func (m *SessionManager) Register(session *webtransportgo.Session, claims *SessionClaims) bool {
m.mu.Lock()
defer m.mu.Unlock()

if m.maxSessions > 0 && len(m.sessions) >= m.maxSessions {
return false
}
m.sessions[session] = &sessionEntry{session: session, claims: claims}
m.count.Add(1)
return true
}

// Unregister removes a session from the active set.
func (m *SessionManager) Unregister(session *webtransportgo.Session) {
m.mu.Lock()
defer m.mu.Unlock()

if _, ok := m.sessions[session]; ok {
delete(m.sessions, session)
m.count.Add(-1)
}
}

// Count returns the number of currently active sessions.
func (m *SessionManager) Count() int64 {
return m.count.Load()
}

// ForEach calls fn for every active session under a read lock.
// fn must not call Register or Unregister.
func (m *SessionManager) ForEach(fn func(session *webtransportgo.Session, claims *SessionClaims)) {
m.mu.RLock()
defer m.mu.RUnlock()

for _, entry := range m.sessions {
fn(entry.session, entry.claims)
}
}
