// CLASSIFICATION: UNCLASSIFIED
// pkg/webtransport/server.go — WebTransport server for the RTSA hot-path
//
// Accepts WebTransport connections authenticated via JWT, filters tracks by
// operator clearance level, applies priority shedding under congestion, and
// sends 128-byte FlatBuffer records as QUIC datagrams at up to 60 Hz.
//
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §3

package webtransport

import (
"context"
"crypto/tls"
"errors"
"net/http"
"sync/atomic"
"time"

"go.opentelemetry.io/otel/metric"
"go.uber.org/zap"

"github.com/quic-go/quic-go/http3"
webtransportgo "github.com/quic-go/webtransport-go"
)

// Config holds all server configuration.
type Config struct {
// ListenAddr is the QUIC listener address (default ":4443").
ListenAddr string

// TLSCert is the path to the TLS certificate file.
TLSCert string

// TLSKey is the path to the TLS private key file.
TLSKey string

// JWTSecret is the HMAC-SHA256 secret used to validate JWT tokens.
// In production, use asymmetric keys (RS256).
JWTSecret []byte

// AllowedOrigins is the set of permitted Origin header values.
// Empty slice disables origin checking (dev only).
AllowedOrigins []string

// MaxSessions is the maximum number of concurrent WebTransport sessions.
// Zero means unlimited.
MaxSessions int

// DatagramBatchSize is the number of 128-byte records per datagram (max 9).
DatagramBatchSize int
}

// TrackSource is implemented by the Redpanda consumer adapter.
// It provides 128-byte FlatBuffer records for broadcasting.
type TrackSource interface {
// Subscribe returns a channel that receives serialised 128-byte FlatBuffer
// records. The channel is closed when ctx is cancelled.
Subscribe(ctx context.Context) <-chan []byte
}

// DatagramSender abstracts the datagram send operation.
// Implemented by *webtransportgo.Session in production and by mocks in tests.
type DatagramSender interface {
SendDatagram(b []byte) error
}

// MaxDatagramBatchSize is the maximum records per QUIC datagram.
// 9 x 128 = 1152 bytes < 1200-byte QUIC MTU.
const MaxDatagramBatchSize = 9

// Server is the WebTransport server.
type Server struct {
cfg       Config
wts       *webtransportgo.Server
sessions  *SessionManager
validator *TokenValidator
source    TrackSource
logger    *zap.Logger

// congested is 1 when QUIC backpressure is detected, 0 otherwise.
congested atomic.Int32

// OTel metrics (nil if no meter provided)
mSessionsActive   metric.Int64ObservableGauge
mDatagramsSent    metric.Int64Counter
mDatagramsDropped metric.Int64Counter
mByteSent         metric.Int64Counter
}

// New creates a Server. The caller must call ListenAndServeTLS to start.
func New(cfg Config, source TrackSource, meter metric.Meter, logger *zap.Logger) (*Server, error) {
if cfg.ListenAddr == "" {
cfg.ListenAddr = ":4443"
}
if cfg.DatagramBatchSize == 0 || cfg.DatagramBatchSize > MaxDatagramBatchSize {
cfg.DatagramBatchSize = MaxDatagramBatchSize
}
if len(cfg.JWTSecret) == 0 {
return nil, errors.New("webtransport: JWTSecret must not be empty")
}

s := &Server{
cfg:       cfg,
sessions:  NewSessionManager(cfg.MaxSessions),
validator: NewTokenValidator(cfg.JWTSecret),
source:    source,
logger:    logger,
}

mux := http.NewServeMux()
mux.HandleFunc("/wt", s.handleSession)

s.wts = &webtransportgo.Server{
H3: &http3.Server{
Addr:    cfg.ListenAddr,
Handler: mux,
},
CheckOrigin: s.checkOrigin,
}

if err := s.initMetrics(meter); err != nil {
return nil, err
}

return s, nil
}

// ListenAndServeTLS starts the WebTransport server using TLS certificate files.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
return s.wts.ListenAndServeTLS(certFile, keyFile)
}

// ListenAndServeTLSConfig starts the server using a pre-built tls.Config.
func (s *Server) ListenAndServeTLSConfig(tlsCfg *tls.Config) error {
s.wts.H3.TLSConfig = tlsCfg
return s.wts.ListenAndServe()
}

// Close shuts down the server.
func (s *Server) Close() error {
return s.wts.Close()
}

// handleSession is the HTTP handler for WebTransport upgrade requests.
func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
// ── 1. Authenticate ──────────────────────────────────────────────────
claims, err := s.validator.ValidateRequest(r)
if err != nil {
s.logger.Warn("WebTransport auth rejected",
zap.String("remote_addr", r.RemoteAddr),
zap.Error(err),
)
http.Error(w, "unauthorized", http.StatusUnauthorized)
return
}

// ── 2. Upgrade connection ────────────────────────────────────────────
session, err := s.wts.Upgrade(w, r)
if err != nil {
s.logger.Error("WebTransport upgrade failed",
zap.String("operator_id", claims.OperatorID),
zap.Error(err),
)
return
}

// ── 3. Enforce session limit ─────────────────────────────────────────
if !s.sessions.Register(session, claims) {
s.logger.Warn("Session limit reached, rejecting",
zap.String("operator_id", claims.OperatorID),
)
_ = session.CloseWithError(0, "server at capacity")
return
}
defer s.sessions.Unregister(session)

s.logger.Info("WebTransport session opened",
zap.String("operator_id", claims.OperatorID),
zap.Uint32("clearance_level", claims.ClearanceLevel),
zap.String("remote_addr", r.RemoteAddr),
)

start := time.Now()
s.StreamRecords(session.Context(), session, claims)
s.logger.Info("WebTransport session closed",
zap.String("operator_id", claims.OperatorID),
zap.Duration("duration", time.Since(start)),
)
}

// StreamRecords reads from the TrackSource and forwards 128-byte records to the
// DatagramSender, applying classification and priority filters. It is exported
// to allow integration testing with mock senders.
//
// Reference: webtransport_guidelines.md §3.3
func (s *Server) StreamRecords(
ctx context.Context,
sender DatagramSender,
claims *SessionClaims,
) {
ch := s.source.Subscribe(ctx)
filter := &RecordFilter{ClearanceLevel: claims.ClearanceLevel}

batch := make([]byte, 0, s.cfg.DatagramBatchSize*RecordSize)

flush := func() {
if len(batch) == 0 {
return
}
if err := sender.SendDatagram(batch); err != nil {
s.logger.Debug("datagram send failed",
zap.String("operator_id", claims.OperatorID),
zap.Error(err),
)
s.congested.Store(1)
if s.mDatagramsDropped != nil {
s.mDatagramsDropped.Add(ctx, int64(len(batch)/RecordSize),
metric.WithAttributes(attrReasonCongestion),
)
}
} else {
s.congested.Store(0)
if s.mDatagramsSent != nil {
s.mDatagramsSent.Add(ctx, int64(len(batch)/RecordSize),
metric.WithAttributes(attrPriorityAll),
)
}
if s.mByteSent != nil {
s.mByteSent.Add(ctx, int64(len(batch)))
}
}
batch = batch[:0]
}

for {
select {
case <-ctx.Done():
flush()
return
case rec, ok := <-ch:
if !ok {
flush()
return
}
if len(rec) < RecordSize {
continue
}

congested := s.congested.Load() != 0
if !filter.ShouldSendRecord(rec, congested) {
if s.mDatagramsDropped != nil {
classLevel := readU32LE(rec, OffClassificationLevel)
if !ShouldSendByClassification(classLevel, claims.ClearanceLevel) {
s.mDatagramsDropped.Add(ctx, 1,
metric.WithAttributes(attrReasonClassification),
)
} else {
s.mDatagramsDropped.Add(ctx, 1,
metric.WithAttributes(attrReasonCongestion),
)
}
}
continue
}

batch = append(batch, rec[:RecordSize]...)
if len(batch)/RecordSize >= s.cfg.DatagramBatchSize {
flush()
}
}
}
}

// checkOrigin validates the request Origin header against the allowed list.
func (s *Server) checkOrigin(r *http.Request) bool {
return IsAllowedOrigin(r.Header.Get("Origin"), s.cfg.AllowedOrigins)
}

// ── OTel metric helpers ───────────────────────────────────────────────────────

func (s *Server) initMetrics(meter metric.Meter) error {
if meter == nil {
return nil
}
var err error

s.mSessionsActive, err = meter.Int64ObservableGauge(
"webtransport_sessions_active",
metric.WithDescription("Number of active WebTransport sessions"),
metric.WithInt64Callback(func(_ context.Context, obs metric.Int64Observer) error {
obs.Observe(s.sessions.Count())
return nil
}),
)
if err != nil {
return err
}

s.mDatagramsSent, err = meter.Int64Counter(
"webtransport_datagrams_sent_total",
metric.WithDescription("Total FlatBuffer records sent via WebTransport datagrams"),
)
if err != nil {
return err
}

s.mDatagramsDropped, err = meter.Int64Counter(
"webtransport_datagrams_dropped_total",
metric.WithDescription("Total records dropped due to congestion or classification filtering"),
)
if err != nil {
return err
}

s.mByteSent, err = meter.Int64Counter(
"webtransport_bytes_sent_total",
metric.WithDescription("Total bytes sent via WebTransport datagrams"),
)
return err
}
