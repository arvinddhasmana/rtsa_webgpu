<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 09 — Feedback & Trust Scoring Service

> **Module**: 09-feedback-trust
> **Phase**: P2 (Core Processing)
> **Dependencies**: Module 02 (protos), Module 03 (shared libraries)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 4 days

---

## 1. Objective

Implement the Feedback Service (`svc-feedback`) that receives operator feedback via gRPC, computes trust scores using a 4-component weighted formula, validates feedback through anti-poisoning checks, and produces events to `feedback.operator.submissions` (all) and `feedback.operator.validated` (trusted only, score ≥ 0.5).

**Acceptance Criteria**:

- gRPC `FeedbackService.SubmitFeedback` accepts operator feedback
- Trust score computed: $Trust = 0.2C + 0.3A + 0.2T + 0.3(1-D)$
- Anti-poisoning guard with all 5 checks enforced
- All feedback → `feedback.operator.submissions`
- Validated feedback (score ≥ 0.5) → `feedback.operator.validated`
- Rate limiting: max 10 feedback/operator/minute
- Audit events emitted for all feedback actions
- `GetFeedbackHistory` queries from in-memory cache
- ≥80% line coverage

---

## 2. Service Structure

```
svc-feedback/
├── cmd/
│   └── feedback/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── trust_scorer.go         # 4-component trust calculation
│   │   ├── trust_scorer_test.go
│   │   ├── anti_poison.go          # Anti-poisoning guard (5 checks)
│   │   ├── anti_poison_test.go
│   │   ├── rate_limiter.go         # Per-operator rate limiting
│   │   └── rate_limiter_test.go
│   ├── handler/
│   │   ├── feedback.go             # gRPC handler implementation
│   │   └── feedback_test.go
│   ├── producer/
│   │   ├── feedback_producer.go
│   │   └── feedback_producer_test.go
│   └── state/
│       ├── operator_history.go     # Tracks operator accuracy, recent feedback
│       └── operator_history_test.go
├── go.mod
├── Dockerfile
└── README.md
```

---

## 3. Trust Scoring Algorithm

### 3.1 Formula

$$Trust = 0.2 \times C + 0.3 \times A + 0.2 \times T + 0.3 \times (1 - D)$$

Where:

- **C** = Clearance Score (from operator's classification clearance)
- **A** = Accuracy Score (historical accuracy ratio)
- **T** = Temporal Score (decay based on time since event)
- **D** = Deviation Score (divergence from consensus)

### 3.2 Component Calculation

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// TrustScorer computes the 4-component trust score.
type TrustScorer struct {
    weightClearance float64 // 0.2
    weightAccuracy  float64 // 0.3
    weightTemporal  float64 // 0.2
    weightDeviation float64 // 0.3
    operatorHistory *state.OperatorHistory
}

// TrustResult holds the computed trust score and breakdown.
type TrustResult struct {
    TotalScore     float64 // 0.0 to 1.0
    ClearanceScore float64
    AccuracyScore  float64
    TemporalScore  float64
    DeviationScore float64
    Validated      bool // true if TotalScore ≥ 0.5
}

// CalculateClearanceScore maps operator clearance to a score.
//   SECRET → 1.0
//   PROTECTED_C → 0.85
//   PROTECTED_B → 0.70
//   PROTECTED_A → 0.50
//   UNCLASSIFIED → 0.30
func CalculateClearanceScore(clearance commonv1.ClassificationLevel) float64 { /* implementation */ }

// CalculateAccuracyScore returns the operator's historical accuracy.
//   accuracy = confirmed_correct / total_feedback
//   Default 0.5 for operators with < 5 feedback submissions.
func (ts *TrustScorer) CalculateAccuracyScore(operatorID string) float64 { /* implementation */ }

// CalculateTemporalScore decays based on time between event and feedback.
//   ≤ 5 min → 1.0
//   5–30 min → linear decay from 1.0 to 0.5
//   30 min – 2 hr → linear decay from 0.5 to 0.1
//   > 2 hr → 0.1
func CalculateTemporalScore(eventTime, feedbackTime time.Time) float64 { /* implementation */ }

// CalculateDeviationScore measures divergence from consensus.
//   1.0 → matches all other feedback for this track
//   0.5 → mixed consensus
//   0.0 → contradicts all other feedback
//   Default 0.5 if no other feedback exists for this track.
func (ts *TrustScorer) CalculateDeviationScore(trackID string, feedbackType commonv1.FeedbackType) float64 { /* implementation */ }

// Score computes the overall trust score.
func (ts *TrustScorer) Score(params TrustParams) *TrustResult { /* implementation */ }

// TrustParams contains inputs for trust scoring.
type TrustParams struct {
    OperatorID         string
    OperatorClearance  commonv1.ClassificationLevel
    TrackID            string
    FeedbackType       commonv1.FeedbackType
    EventTime          time.Time // When the track/alert event occurred
    FeedbackTime       time.Time // When the feedback was submitted
}
```

---

## 4. Anti-Poisoning Guard

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// AntiPoisonGuard implements 5 statistical checks to detect
// bulk anomalous feedback patterns (data poisoning attempts).
type AntiPoisonGuard struct {
    operatorHistory *state.OperatorHistory
    logger          *zap.Logger
}

// PoisonCheckResult holds the outcome of anti-poisoning checks.
type PoisonCheckResult struct {
    Passed   bool
    Checks   map[string]bool   // check_name → passed
    Details  map[string]string // check_name → detail
}

// Check runs all 5 anti-poisoning checks.
// Returns Passed=true only if ALL checks pass.
//
// Check 1: Distribution Check (chi-squared)
//   - Operator's feedback type distribution should not differ significantly from global
//   - Chi-squared test, p > 0.05 → pass
//   - Requires ≥ 10 feedback submissions from operator
//
// Check 2: Label Flip Rate
//   - Percentage of operator's feedback that contradicts previous feedback on same track
//   - ≤ 20% label flip rate → pass
//
// Check 3: Source Diversity
//   - Operator must provide feedback on tracks from ≥ 3 different sensor sources
//   - Requires ≥ 5 feedback submissions
//
// Check 4: Temporal Clustering
//   - No sliding 5-minute window should contain > 40% of operator's total feedback
//   - Detects burst-submission attacks
//
// Check 5: High-Trust Ratio
//   - ≥ 60% of operator's feedback should have been independently validated
//   - Requires ≥ 10 feedback submissions
func (g *AntiPoisonGuard) Check(operatorID string) *PoisonCheckResult { /* implementation */ }
```

---

## 5. Rate Limiter

```go
// CLASSIFICATION: UNCLASSIFIED
package domain

// RateLimiter enforces per-operator rate limits using a sliding window.
type RateLimiter struct {
    mu          sync.Mutex
    windows     map[string]*slidingWindow // key: operator_id
    maxPerMinute int                       // default: 10
}

// Allow returns true if the operator has not exceeded the rate limit.
// Uses a 1-minute sliding window.
func (rl *RateLimiter) Allow(operatorID string) bool { /* implementation */ }
```

---

## 6. Operator History State

```go
// CLASSIFICATION: UNCLASSIFIED
package state

// OperatorHistory maintains per-operator feedback statistics.
type OperatorHistory struct {
    mu       sync.RWMutex
    operators map[string]*OperatorStats
}

// OperatorStats tracks an operator's feedback history.
type OperatorStats struct {
    OperatorID         string
    TotalFeedback      int
    ConfirmedCorrect   int    // Feedback independently validated as correct
    FeedbackByType     map[string]int // FeedbackType → count
    FeedbackByTrack    map[string][]FeedbackEntry
    RecentFeedback     []time.Time // Timestamps of recent submissions (for clustering check)
    TrackSensorSources map[string]bool // Unique sensor sources from tracks receiving feedback
}

// FeedbackEntry records a single feedback submission for history.
type FeedbackEntry struct {
    FeedbackID   string
    TrackID      string
    FeedbackType commonv1.FeedbackType
    Timestamp    time.Time
    TrustScore   float64
    LabelFlipped bool // true if contradicts previous feedback on same track
}

// RecordFeedback adds a feedback entry to operator history.
func (oh *OperatorHistory) RecordFeedback(entry FeedbackEntry) { /* implementation */ }

// GetStats returns operator statistics. Returns nil if operator unknown.
func (oh *OperatorHistory) GetStats(operatorID string) *OperatorStats { /* implementation */ }

// TypeDistribution returns the normalized distribution of feedback types.
func (os *OperatorStats) TypeDistribution() map[string]float64 { /* implementation */ }

// LabelFlipRate returns the proportion of label-flipped feedback.
func (os *OperatorStats) LabelFlipRate() float64 { /* implementation */ }

// UniqueSensorSources returns the count of unique sensor sources.
func (os *OperatorStats) UniqueSensorSources() int { /* implementation */ }
```

---

## 7. gRPC Handler

```go
// CLASSIFICATION: UNCLASSIFIED
package handler

// FeedbackHandler implements FeedbackService.
type FeedbackHandler struct {
    feedbackv1.UnimplementedFeedbackServiceServer

    trustScorer   *domain.TrustScorer
    antiPoison    *domain.AntiPoisonGuard
    rateLimiter   *domain.RateLimiter
    rawProducer   *producer.FeedbackProducer  // feedback.operator.submissions
    validProducer *producer.FeedbackProducer  // feedback.operator.validated
    auditEmitter  *audit.Emitter
    logger        *zap.Logger
}

// SubmitFeedback processes an operator feedback submission.
// Flow:
//   1. Validate request fields (track_id, operator_id, feedback_type required)
//   2. Check rate limit → if exceeded: codes.ResourceExhausted
//   3. Run anti-poisoning checks → if failed: log warning, continue (do not block)
//   4. Compute trust score
//   5. Build OperatorFeedback proto with trust breakdown
//   6. Produce to feedback.operator.submissions (ALL feedback)
//   7. If trust_score ≥ 0.5 → Produce to feedback.operator.validated
//   8. Record in operator history
//   9. Emit audit event
//   10. Return SubmitFeedbackResponse
//
// Note: Anti-poisoning failure does NOT block submission. It is logged
// and the feedback is still submitted but NOT validated (score adjusted).
func (h *FeedbackHandler) SubmitFeedback(ctx context.Context,
    req *feedbackv1.SubmitFeedbackRequest) (*feedbackv1.SubmitFeedbackResponse, error) { /* implementation */ }

// GetFeedbackHistory queries feedback history from in-memory state.
func (h *FeedbackHandler) GetFeedbackHistory(ctx context.Context,
    req *feedbackv1.GetFeedbackHistoryRequest) (*feedbackv1.GetFeedbackHistoryResponse, error) { /* implementation */ }
```

---

## 8. Test Scenarios

### 8.1 Unit Tests

| #   | Test                                                           | Expected               |
| --- | -------------------------------------------------------------- | ---------------------- |
| T01 | Clearance SECRET → score 1.0                                   | 1.0                    |
| T02 | Clearance UNCLASSIFIED → score 0.3                             | 0.3                    |
| T03 | Accuracy: 8/10 correct                                         | 0.8                    |
| T04 | Accuracy: new operator (< 5 feedback)                          | 0.5 (default)          |
| T05 | Temporal: feedback within 5 min                                | 1.0                    |
| T06 | Temporal: feedback at 15 min                                   | ~0.75 (linear decay)   |
| T07 | Temporal: feedback at 2+ hours                                 | 0.1                    |
| T08 | Deviation: matches all consensus                               | 1.0                    |
| T09 | Deviation: contradicts all                                     | 0.0                    |
| T10 | Deviation: no other feedback                                   | 0.5 (default)          |
| T11 | Trust composite: SECRET, accuracy=0.8, 3min, matches consensus | ~0.82, validated       |
| T12 | Trust composite: UNCLASSIFIED, accuracy=0.3, 3hr, contradicts  | ~0.24, not validated   |
| T13 | Anti-poison: all checks pass                                   | Passed=true            |
| T14 | Anti-poison: label flip > 20%                                  | Check 2 fails          |
| T15 | Anti-poison: < 3 sensor sources                                | Check 3 fails          |
| T16 | Anti-poison: burst 40%+ in 5min                                | Check 4 fails          |
| T17 | Anti-poison: < 60% validated                                   | Check 5 fails          |
| T18 | Rate limit: 10 requests in 1 min                               | Allowed                |
| T19 | Rate limit: 11th request in 1 min                              | Rejected               |
| T20 | Rate limit: window expires                                     | Next request allowed   |
| T21 | Handler: valid feedback, trust ≥ 0.5                           | Both topics produced   |
| T22 | Handler: valid feedback, trust < 0.5                           | Only submissions topic |
| T23 | Handler: rate limited                                          | ResourceExhausted      |
| T24 | Handler: missing track_id                                      | InvalidArgument        |

### 8.2 Integration Tests

| #    | Test                          | Expected                      |
| ---- | ----------------------------- | ----------------------------- |
| IT01 | Submit feedback → both topics | Message in both topics        |
| IT02 | Submit low-trust feedback     | Message only in submissions   |
| IT03 | Rate limit exceeded           | ResourceExhausted, no message |
| IT04 | Audit event emitted           | Message in audit.events       |

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement Module 09 from docs/implementation/09-feedback-trust.md

Context:
- Read docs/implementation/00-implementation-overview.md for global conventions
- Read docs/implementation/02-protobuf-schemas.md for OperatorFeedback and FeedbackService protos
- Read docs/architecture/component_design.md §6 for feedback component diagram
- Trust formula: Trust = 0.2C + 0.3A + 0.2T + 0.3(1-D)
- Anti-poisoning: 5 statistical checks (see §4)
- Rate limiting: 10 feedback/operator/minute sliding window
- All feedback goes to submissions topic; only validated (≥0.5) goes to validated topic
- Anti-poisoning failure does NOT block — it adjusts trust score

Deliverables:
1. Complete svc-feedback/ with all files
2. Trust scorer with 4-component formula
3. Anti-poisoning guard with 5 checks
4. Rate limiter per operator
5. Operator history state management
6. Unit tests (≥80% coverage)
7. Integration tests with testcontainers
```
