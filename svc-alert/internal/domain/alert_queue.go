// CLASSIFICATION: UNCLASSIFIED
package domain

import (
"fmt"
"sort"
"sync"
"time"

commonv1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/common/v1"
inferencev1 "github.com/arvinddhasmana/RTSA_VS_Opus/gen/go/rtsa/inference/v1"
)

const (
// DefaultMaxQueueSize is the default maximum number of alerts kept in memory.
DefaultMaxQueueSize = 10000
// subscriberBufferSize is the channel buffer for each streaming subscriber.
subscriberBufferSize = 256
)

// QueuedAlert wraps an AnomalyAlert with queue metadata.
type QueuedAlert struct {
Alert        *inferencev1.AnomalyAlert
QueuedAt     time.Time
Acknowledged bool
AckedBy      string
AckedAt      *time.Time
Comment      string
}

// AlertQueue is a priority queue that orders alerts by:
//  1. Severity (CRITICAL > ELEVATED > WATCH > NORMAL) — higher severity first
//  2. Timestamp (newer first within same severity)
//
// The queue is thread-safe via an internal RWMutex.
// Streaming clients subscribe via Subscribe/Unsubscribe to receive new alerts.
type AlertQueue struct {
mu         sync.RWMutex
alerts     map[string]*QueuedAlert // indexed by alert_id
byPriority []*QueuedAlert          // sorted slice maintained in priority order

maxSize int

subsMu      sync.Mutex
subscribers []chan *inferencev1.AnomalyAlert
}

// NewAlertQueue creates a new AlertQueue with the given maximum size.
// If maxSize <= 0, DefaultMaxQueueSize is used.
func NewAlertQueue(maxSize int) *AlertQueue {
if maxSize <= 0 {
maxSize = DefaultMaxQueueSize
}
return &AlertQueue{
alerts:  make(map[string]*QueuedAlert),
maxSize: maxSize,
}
}

// Subscribe creates and registers a channel that receives newly enqueued alerts.
// The caller MUST call Unsubscribe when done to release resources.
func (q *AlertQueue) Subscribe() chan *inferencev1.AnomalyAlert {
ch := make(chan *inferencev1.AnomalyAlert, subscriberBufferSize)
q.subsMu.Lock()
q.subscribers = append(q.subscribers, ch)
q.subsMu.Unlock()
return ch
}

// Unsubscribe removes and closes a subscriber channel.
func (q *AlertQueue) Unsubscribe(ch chan *inferencev1.AnomalyAlert) {
q.subsMu.Lock()
defer q.subsMu.Unlock()
for i, sub := range q.subscribers {
if sub == ch {
q.subscribers = append(q.subscribers[:i], q.subscribers[i+1:]...)
close(ch)
return
}
}
}

// Enqueue adds an alert to the priority queue.
// If the alert ID already exists, the existing entry is updated in place.
// If the queue is at maxSize, the lowest-priority alert is dropped to make room.
// All active subscribers are notified (non-blocking).
func (q *AlertQueue) Enqueue(alert *inferencev1.AnomalyAlert) {
if alert == nil || alert.GetAlertId() == "" {
return
}

q.mu.Lock()

// Update existing entry
if existing, ok := q.alerts[alert.GetAlertId()]; ok {
existing.Alert = alert
q.resort()
q.mu.Unlock()
q.notify(alert)
return
}

// Enforce max size: evict the lowest-priority alert (last element after sort)
if len(q.byPriority) >= q.maxSize {
dropped := q.byPriority[len(q.byPriority)-1]
q.byPriority = q.byPriority[:len(q.byPriority)-1]
delete(q.alerts, dropped.Alert.GetAlertId())
}

qa := &QueuedAlert{
Alert:    alert,
QueuedAt: time.Now(),
}
q.alerts[alert.GetAlertId()] = qa
q.byPriority = append(q.byPriority, qa)
q.resort()
q.mu.Unlock()

q.notify(alert)
}

// Acknowledge marks an alert as acknowledged by an operator.
// Returns the acknowledgment timestamp or an error if the alert is not found.
func (q *AlertQueue) Acknowledge(alertID, operatorID, comment string) (*time.Time, error) {
q.mu.Lock()
defer q.mu.Unlock()

qa, ok := q.alerts[alertID]
if !ok {
return nil, fmt.Errorf("[domain].[AlertQueue.Acknowledge](%s): %w", alertID, ErrAlertNotFound)
}

now := time.Now()
qa.Acknowledged = true
qa.AckedBy = operatorID
qa.AckedAt = &now
qa.Comment = comment
// Reflect acknowledged state in the embedded proto for downstream consumers
qa.Alert.Acknowledged = true

return &now, nil
}

// Get returns the QueuedAlert for the given alert ID.
// Returns (nil, false) if the alert is not found.
func (q *AlertQueue) Get(alertID string) (*QueuedAlert, bool) {
q.mu.RLock()
defer q.mu.RUnlock()
qa, ok := q.alerts[alertID]
return qa, ok
}

// GetUnacknowledged returns all unacknowledged alerts in priority order.
// The returned slice is a snapshot; callers should not modify it.
func (q *AlertQueue) GetUnacknowledged() []*QueuedAlert {
q.mu.RLock()
defer q.mu.RUnlock()

result := make([]*QueuedAlert, 0, len(q.byPriority))
for _, qa := range q.byPriority {
if !qa.Acknowledged {
result = append(result, qa)
}
}
return result
}

// UnacknowledgedCount returns the count of unacknowledged alerts grouped by severity.
func (q *AlertQueue) UnacknowledgedCount() map[commonv1.AlertSeverity]int {
q.mu.RLock()
defer q.mu.RUnlock()

counts := make(map[commonv1.AlertSeverity]int)
for _, qa := range q.byPriority {
if !qa.Acknowledged {
counts[qa.Alert.GetSeverity()]++
}
}
return counts
}

// Size returns the total number of alerts currently in the queue.
func (q *AlertQueue) Size() int {
q.mu.RLock()
defer q.mu.RUnlock()
return len(q.byPriority)
}

// resort sorts byPriority in-place. Must be called with q.mu held (write).
func (q *AlertQueue) resort() {
sort.Slice(q.byPriority, func(i, j int) bool {
ri := severityRank(q.byPriority[i].Alert.GetSeverity())
rj := severityRank(q.byPriority[j].Alert.GetSeverity())
if ri != rj {
return ri > rj // higher rank = higher priority
}
// Equal severity: newer detected_at first
ti := q.byPriority[i].Alert.GetDetectedAt()
tj := q.byPriority[j].Alert.GetDetectedAt()
if ti == nil && tj == nil {
return false
}
if ti == nil {
return false
}
if tj == nil {
return true
}
return ti.AsTime().After(tj.AsTime())
})
}

// notify sends the alert to all active subscribers without blocking.
// If a subscriber's channel buffer is full, the notification is dropped for that subscriber.
func (q *AlertQueue) notify(alert *inferencev1.AnomalyAlert) {
q.subsMu.Lock()
subs := make([]chan *inferencev1.AnomalyAlert, len(q.subscribers))
copy(subs, q.subscribers)
q.subsMu.Unlock()

for _, ch := range subs {
select {
case ch <- alert:
default:
// Subscriber is slow; drop notification to avoid blocking the producer.
}
}
}

// severityRank returns a numeric priority rank for an AlertSeverity value.
// Higher rank = higher priority in the queue.
//
// CRITICAL = 3, ELEVATED = 2, WATCH = 1, NORMAL = 0, UNSPECIFIED = -1
func severityRank(s commonv1.AlertSeverity) int {
switch s {
case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL:
return 3
case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED:
return 2
case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:
return 1
case commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL:
return 0
default:
return -1
}
}

// SeverityLabel returns the lowercase string label for a severity (used in metrics).
func SeverityLabel(s commonv1.AlertSeverity) string {
switch s {
case commonv1.AlertSeverity_ALERT_SEVERITY_CRITICAL:
return "critical"
case commonv1.AlertSeverity_ALERT_SEVERITY_ELEVATED:
return "elevated"
case commonv1.AlertSeverity_ALERT_SEVERITY_WATCH:
return "watch"
case commonv1.AlertSeverity_ALERT_SEVERITY_NORMAL:
return "normal"
default:
return "unspecified"
}
}

// ErrAlertNotFound is returned when an operation references an unknown alert ID.
var ErrAlertNotFound = fmt.Errorf("alert not found")
