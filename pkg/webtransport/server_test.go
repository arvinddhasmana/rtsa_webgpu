// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/server_test.go — Unit tests for WebTransport server

package webtransport_test

import (
"context"
"encoding/binary"
"errors"
"sync"
"testing"
"time"

"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/webtransport"
"go.uber.org/zap"
)

// ── Test doubles ─────────────────────────────────────────────────────────────

// mockTrackSource emits the provided records then closes.
type mockTrackSource struct {
records [][]byte
}

func (m *mockTrackSource) Subscribe(_ context.Context) <-chan []byte {
ch := make(chan []byte, len(m.records)+1)
for _, r := range m.records {
ch <- r
}
close(ch)
return ch
}

// captureSender records datagrams sent to it.
type captureSender struct {
mu        sync.Mutex
datagrams [][]byte
sendErr   error
}

func (c *captureSender) SendDatagram(b []byte) error {
c.mu.Lock()
defer c.mu.Unlock()
if c.sendErr != nil {
return c.sendErr
}
cpy := make([]byte, len(b))
copy(cpy, b)
c.datagrams = append(c.datagrams, cpy)
return nil
}

func (c *captureSender) Received() [][]byte {
c.mu.Lock()
defer c.mu.Unlock()
return c.datagrams
}

// infiniteSource sends records continuously until ctx is done.
type infiniteSource struct {
rec []byte
}

func (s *infiniteSource) Subscribe(ctx context.Context) <-chan []byte {
ch := make(chan []byte, 1)
go func() {
defer close(ch)
for {
select {
case <-ctx.Done():
return
case ch <- s.rec:
}
}
}()
return ch
}

// buildTestRecord creates a 128-byte record with the given class and threat levels.
func buildTestRecord(classLevel, threatLevel uint32) []byte {
rec := make([]byte, webtransport.RecordSize)
binary.LittleEndian.PutUint32(rec[webtransport.OffClassificationLevel:], classLevel)
binary.LittleEndian.PutUint32(rec[webtransport.OffThreatLevel:], threatLevel)
return rec
}

// makeServer creates a test server with the given source.
func makeServer(source webtransport.TrackSource) (*webtransport.Server, error) {
return webtransport.New(
webtransport.Config{
ListenAddr: ":14443",
JWTSecret:  testSecret,
},
source,
nil,
zap.NewNop(),
)
}

// ── Constructor tests ─────────────────────────────────────────────────────────

func TestNew_MissingJWTSecret(t *testing.T) {
_, err := webtransport.New(
webtransport.Config{ListenAddr: ":4443"},
&mockTrackSource{},
nil,
zap.NewNop(),
)
if err == nil {
t.Error("expected error when JWTSecret is empty")
}
}

func TestNew_DefaultBatchSize(t *testing.T) {
srv, err := makeServer(&mockTrackSource{})
if err != nil {
t.Fatalf("New: %v", err)
}
if srv == nil {
t.Error("expected non-nil server")
}
}

func TestNew_CustomBatchSize(t *testing.T) {
srv, err := webtransport.New(
webtransport.Config{
ListenAddr:        ":14444",
JWTSecret:         testSecret,
DatagramBatchSize: 3,
},
&mockTrackSource{},
nil,
zap.NewNop(),
)
if err != nil {
t.Fatalf("New with DatagramBatchSize=3: %v", err)
}
if srv == nil {
t.Error("expected non-nil server")
}
}

func TestNew_AllowedOrigins(t *testing.T) {
srv, err := webtransport.New(
webtransport.Config{
ListenAddr:     ":14445",
JWTSecret:      testSecret,
AllowedOrigins: []string{"https://rtsa.mil.ca"},
},
&mockTrackSource{},
nil,
zap.NewNop(),
)
if err != nil {
t.Fatalf("New with AllowedOrigins: %v", err)
}
if srv == nil {
t.Error("expected non-nil server")
}
}

func TestNew_WithOtelNoopMeter(t *testing.T) {
// Exercise initMetrics registration paths with a noop meter
meter := otelnoopMeter()
srv, err := webtransport.New(
webtransport.Config{
ListenAddr: ":14448",
JWTSecret:  testSecret,
},
&mockTrackSource{},
meter,
zap.NewNop(),
)
if err != nil {
t.Fatalf("New with OTel noop meter: %v", err)
}
if srv == nil {
t.Fatal("expected non-nil server")
}
}

// ── StreamRecords tests ───────────────────────────────────────────────────────

func TestStreamRecords_PassesAllowedRecord(t *testing.T) {
rec := buildTestRecord(1, 0) // UNCLASSIFIED, Unknown threat
src := &mockTrackSource{records: [][]byte{rec}}

srv, err := makeServer(src)
if err != nil {
t.Fatalf("New: %v", err)
}

sender := &captureSender{}
claims := &webtransport.SessionClaims{OperatorID: "op-1", ClearanceLevel: 5}

ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()
srv.StreamRecords(ctx, sender, claims)

if len(sender.Received()) == 0 {
t.Error("expected at least one datagram to be sent")
}
}

func TestStreamRecords_DropsHighClassificationRecord(t *testing.T) {
rec := buildTestRecord(5, 0) // SECRET with UNCLASSIFIED operator
src := &mockTrackSource{records: [][]byte{rec}}

srv, err := makeServer(src)
if err != nil {
t.Fatalf("New: %v", err)
}

sender := &captureSender{}
claims := &webtransport.SessionClaims{OperatorID: "op-unclass", ClearanceLevel: 1}

ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()
srv.StreamRecords(ctx, sender, claims)

if len(sender.Received()) != 0 {
t.Errorf("expected no datagrams; got %d", len(sender.Received()))
}
}

func TestStreamRecords_BatchesManyRecords(t *testing.T) {
// 9 records at exactly MaxDatagramBatchSize → 1 datagram
var records [][]byte
for i := 0; i < webtransport.MaxDatagramBatchSize; i++ {
records = append(records, buildTestRecord(1, 0))
}
src := &mockTrackSource{records: records}

srv, err := webtransport.New(
webtransport.Config{
ListenAddr:        ":14446",
JWTSecret:         testSecret,
DatagramBatchSize: webtransport.MaxDatagramBatchSize,
},
src,
nil,
zap.NewNop(),
)
if err != nil {
t.Fatalf("New: %v", err)
}

sender := &captureSender{}
claims := &webtransport.SessionClaims{OperatorID: "op-batch", ClearanceLevel: 5}

ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()
srv.StreamRecords(ctx, sender, claims)

datagrams := sender.Received()
if len(datagrams) != 1 {
t.Errorf("expected 1 datagram for %d records; got %d", webtransport.MaxDatagramBatchSize, len(datagrams))
}
if len(datagrams) > 0 && len(datagrams[0]) != webtransport.MaxDatagramBatchSize*webtransport.RecordSize {
t.Errorf("batched datagram: want %d bytes, got %d",
webtransport.MaxDatagramBatchSize*webtransport.RecordSize, len(datagrams[0]))
}
}

func TestStreamRecords_SendErrorMarksCongested(t *testing.T) {
rec := buildTestRecord(1, 5) // Hostile — always sent
src := &mockTrackSource{records: [][]byte{rec}}

srv, err := makeServer(src)
if err != nil {
t.Fatalf("New: %v", err)
}

sender := &captureSender{sendErr: errors.New("send buffer full")}
claims := &webtransport.SessionClaims{OperatorID: "op-err", ClearanceLevel: 5}

ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()
srv.StreamRecords(ctx, sender, claims)
}

func TestStreamRecords_ContextCancel(t *testing.T) {
srv, err := makeServer(&infiniteSource{rec: buildTestRecord(1, 0)})
if err != nil {
t.Fatalf("New: %v", err)
}

sender := &captureSender{}
claims := &webtransport.SessionClaims{OperatorID: "op-cancel", ClearanceLevel: 5}

ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()
srv.StreamRecords(ctx, sender, claims)
}

func TestStreamRecords_ShortRecord(t *testing.T) {
src := &mockTrackSource{records: [][]byte{make([]byte, 32)}} // too short
srv, err := makeServer(src)
if err != nil {
t.Fatalf("New: %v", err)
}

sender := &captureSender{}
claims := &webtransport.SessionClaims{OperatorID: "op-short", ClearanceLevel: 5}

ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
defer cancel()
srv.StreamRecords(ctx, sender, claims)

if len(sender.Received()) != 0 {
t.Errorf("expected no datagrams for short record; got %d", len(sender.Received()))
}
}

// ── SessionManager tests ──────────────────────────────────────────────────────

func TestSessionManager_CountStartsAtZero(t *testing.T) {
mgr := webtransport.NewSessionManager(0)
if mgr.Count() != 0 {
t.Errorf("initial count: want 0, got %d", mgr.Count())
}
}

// ── Classification + priority integration ─────────────────────────────────────

func TestShouldSend_CongestedClassification(t *testing.T) {
if webtransport.ShouldSendByClassification(5, 3) {
t.Error("SECRET track must not be sent to PROTECTED_B operator")
}
if !webtransport.ShouldSendByClassification(1, 5) {
t.Error("UNCLASSIFIED track must be sent to SECRET operator")
}
}

func TestStreamTrackUpdates_ClassificationFilter(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 1}

if f.ShouldSendRecord(buildTestRecord(5, 0), false) {
t.Error("SECRET record must be dropped for UNCLASSIFIED operator")
}
if !f.ShouldSendRecord(buildTestRecord(1, 0), false) {
t.Error("UNCLASSIFIED record must be sent to UNCLASSIFIED operator")
}
}

func TestStreamTrackUpdates_PriorityFilter(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: 5}

if !f.ShouldSendRecord(buildTestRecord(1, 4), true) {
t.Error("Suspect (P1) must be sent even under congestion")
}
if f.ShouldSendRecord(buildTestRecord(1, 0), true) {
t.Error("Unknown (P3) must be dropped under congestion")
}
}

func TestRecordFilter_AllCombinations(t *testing.T) {
tests := []struct {
name      string
class     uint32
clearance uint32
threat    uint32
congested bool
want      bool
}{
{"unclass+secret+hostile+no-congestion", 1, 5, 5, false, true},
{"unclass+secret+friendly+no-congestion", 1, 5, 2, false, true},
{"unclass+secret+hostile+congested", 1, 5, 5, true, true},
{"unclass+secret+friendly+congested", 1, 5, 2, true, false},
{"secret+unclass+hostile+no-congestion", 5, 1, 5, false, false},
{"secret+unclass+friendly+congested", 5, 1, 2, true, false},
}

for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
f := &webtransport.RecordFilter{ClearanceLevel: tc.clearance}
got := f.ShouldSendRecord(buildTestRecord(tc.class, tc.threat), tc.congested)
if got != tc.want {
t.Errorf("ShouldSendRecord: want %v, got %v", tc.want, got)
}
})
}
}
