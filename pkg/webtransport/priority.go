// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/priority.go — Load-based track priority shedding
//
// Under QUIC congestion, low-priority tracks are dropped to ensure hostile
// and suspect tracks are always delivered.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §5

package webtransport

// Priority levels for track shedding under congestion.
const (
// PriorityP0 is always delivered (Hostile).
PriorityP0 = 0
// PriorityP1 is delivered unless severe congestion (Suspect).
PriorityP1 = 1
// PriorityP2 is dropped first under load (Neutral, Pending).
PriorityP2 = 2
// PriorityP3 is dropped aggressively under load (Friendly, Unknown).
PriorityP3 = 3
)

// ThreatLevelPriority maps ThreatLevel uint32 values (from the .fbs schema)
// to a Priority level.
//
//0=Unknown→P3, 1=Pending→P2, 2=Friendly→P3, 3=Neutral→P2, 4=Suspect→P1, 5=Hostile→P0
func ThreatLevelPriority(threatLevel uint32) int {
switch threatLevel {
case 5: // Hostile
return PriorityP0
case 4: // Suspect
return PriorityP1
case 3: // Neutral
return PriorityP2
case 1: // Pending
return PriorityP2
default: // Unknown (0), Friendly (2)
return PriorityP3
}
}

// ShouldSend returns true if the track should be sent given the current
// congestion state and its threat level.
//
// Rules:
//   - P0/P1 (Hostile/Suspect): always send
//   - P2/P3 (others): drop when congested
//
// Reference: webtransport_guidelines.md §5.2
func ShouldSend(threatLevel uint32, congested bool) bool {
if !congested {
return true
}
p := ThreatLevelPriority(threatLevel)
return p <= PriorityP1
}

// ShouldSendByClassification returns false if the track's classification_level
// exceeds the operator's clearance_level.
//
// Reference: webtransport_guidelines.md §7.4
func ShouldSendByClassification(trackClassLevel, operatorClearance uint32) bool {
return trackClassLevel <= operatorClearance
}
