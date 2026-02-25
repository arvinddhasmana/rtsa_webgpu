// CLASSIFICATION: UNCLASSIFIED
package consumer_test

import (
"encoding/json"
"testing"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/svc-training/internal/consumer"
)

func TestNoopModelCandidateJSONMarshal(t *testing.T) {
tests := []struct {
name      string
candidate consumer.NoopModelCandidate
}{
{
name: "noop candidate",
candidate: consumer.NoopModelCandidate{
ModelID:   "noop-v0",
Status:    "stub",
Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
},
},
{
name: "empty candidate",
candidate: consumer.NoopModelCandidate{},
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

var decoded consumer.NoopModelCandidate
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

func TestNoopModelCandidateJSONFields(t *testing.T) {
candidate := consumer.NoopModelCandidate{
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

if _, ok := raw["model_id"]; !ok {
t.Error("expected field 'model_id' in JSON output")
}
if _, ok := raw["status"]; !ok {
t.Error("expected field 'status' in JSON output")
}
if _, ok := raw["timestamp"]; !ok {
t.Error("expected field 'timestamp' in JSON output")
}
}
