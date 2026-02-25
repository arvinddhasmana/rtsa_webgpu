// CLASSIFICATION: UNCLASSIFIED
package producer_test

import (
"encoding/json"
"testing"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-training/internal/producer"
)

func TestModelCandidateJSONMarshal(t *testing.T) {
tests := []struct {
name      string
candidate producer.ModelCandidate
}{
{
name: "noop candidate",
candidate: producer.ModelCandidate{
ModelID:   "noop-v0",
Status:    "stub",
Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
},
},
{
name: "empty candidate",
candidate: producer.ModelCandidate{},
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
data, err := json.Marshal(tt.candidate)
if err != nil {
t.Fatalf("json.Marshal() error = %v", err)
}
if len(data) == 0 {
t.Fatal("json.Marshal() returned empty bytes")
}

var decoded producer.ModelCandidate
if err := json.Unmarshal(data, &decoded); err != nil {
t.Fatalf("json.Unmarshal() error = %v", err)
}
if decoded.ModelID != tt.candidate.ModelID {
t.Errorf("ModelID = %q, want %q", decoded.ModelID, tt.candidate.ModelID)
}
if decoded.Status != tt.candidate.Status {
t.Errorf("Status = %q, want %q", decoded.Status, tt.candidate.Status)
}
})
}
}

func TestModelCandidateJSONKeys(t *testing.T) {
candidate := producer.ModelCandidate{
ModelID:   "noop-v0",
Status:    "stub",
Timestamp: time.Now().UTC(),
}

data, err := json.Marshal(candidate)
if err != nil {
t.Fatalf("json.Marshal() error = %v", err)
}

var raw map[string]interface{}
if err := json.Unmarshal(data, &raw); err != nil {
t.Fatalf("json.Unmarshal() error = %v", err)
}

for _, key := range []string{"model_id", "status", "timestamp"} {
if _, ok := raw[key]; !ok {
t.Errorf("expected JSON key %q in output", key)
}
}
}
