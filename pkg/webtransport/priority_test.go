// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/priority_test.go — Unit tests for priority shedding logic

package webtransport_test

import (
"testing"

"github.com/arvinddhasmana/rtsa_webgpu/pkg/webtransport"
)

func TestShouldSend_NoCongestion(t *testing.T) {
// When not congested, all tracks must pass through
threatLevels := []uint32{0, 1, 2, 3, 4, 5}
for _, tl := range threatLevels {
if !webtransport.ShouldSend(tl, false) {
t.Errorf("ShouldSend(threatLevel=%d, congested=false) returned false", tl)
}
}
}

func TestShouldSend_CongestedAlwaysDeliversP0P1(t *testing.T) {
// Hostile (5) and Suspect (4) must always be sent
if !webtransport.ShouldSend(5, true) {
t.Error("Hostile (5) must be sent even under congestion")
}
if !webtransport.ShouldSend(4, true) {
t.Error("Suspect (4) must be sent even under congestion")
}
}

func TestShouldSend_CongestedDropsLowPriority(t *testing.T) {
// Friendly (2), Unknown (0), Neutral (3), Pending (1) dropped under congestion
droppedLevels := []uint32{0, 1, 2, 3}
for _, tl := range droppedLevels {
if webtransport.ShouldSend(tl, true) {
t.Errorf("ShouldSend(threatLevel=%d, congested=true) must return false", tl)
}
}
}

func TestShouldSendByClassification(t *testing.T) {
tests := []struct {
trackClass      uint32
operatorClearance uint32
want            bool
}{
{1, 5, true},  // UNCLASSIFIED track, SECRET clearance → send
{3, 3, true},  // PROTECTED_B track, PROTECTED_B clearance → send
{4, 3, false}, // PROTECTED_C track, PROTECTED_B clearance → drop
{5, 4, false}, // SECRET track, PROTECTED_C clearance → drop
{0, 0, true},  // unspecified levels → send (0 <= 0)
}

for _, tc := range tests {
got := webtransport.ShouldSendByClassification(tc.trackClass, tc.operatorClearance)
if got != tc.want {
t.Errorf(
"ShouldSendByClassification(track=%d, clearance=%d): want %v, got %v",
tc.trackClass, tc.operatorClearance, tc.want, got,
)
}
}
}

func TestThreatLevelPriority(t *testing.T) {
tests := []struct {
level uint32
want  int
}{
{5, webtransport.PriorityP0}, // Hostile
{4, webtransport.PriorityP1}, // Suspect
{3, webtransport.PriorityP2}, // Neutral
{1, webtransport.PriorityP2}, // Pending
{2, webtransport.PriorityP3}, // Friendly
{0, webtransport.PriorityP3}, // Unknown
}
for _, tc := range tests {
got := webtransport.ThreatLevelPriority(tc.level)
if got != tc.want {
t.Errorf("ThreatLevelPriority(%d): want %d, got %d", tc.level, tc.want, got)
}
}
}
