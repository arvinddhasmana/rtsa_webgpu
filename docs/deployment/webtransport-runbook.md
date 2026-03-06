<!-- CLASSIFICATION: UNCLASSIFIED -->

# WebTransport Service — Operator Runbook

> **Document**: RTSA WebTransport Deployment & Operations
> **Version**: 1.0
> **Classification**: UNCLASSIFIED
> **Audience**: System administrators and operations staff

---

## 1. Service Overview

The WebTransport service (`pkg/webtransport`) delivers real-time track data to the WebGPU COP frontend via QUIC datagrams over WebTransport (RFC 9297). It is the hot path for 50k-track @ 60 FPS delivery.

| Property             | Value                                              |
| -------------------- | -------------------------------------------------- |
| Port                 | 4443 (QUIC/UDP)                                    |
| Protocol             | WebTransport over HTTP/3                           |
| TLS                  | TLS 1.3 with CSE-approved cipher suites            |
| Auth                 | JWT (HS256 dev, RS256 production)                  |
| Max sessions         | 100 simultaneous                                   |
| Record format        | 128-byte FlatBuffer (binary), up to 9 per datagram |

---

## 2. Prerequisites

- TLS 1.3 certificate and private key (issued by the site CA)
- JWT signing key (RS256 private key in production; HS256 secret in dev)
- Firewall rules: allow UDP/4443 inbound from client networks
- Redpanda topic `track-updates` accessible from the server

---

## 3. Configuration

All configuration is via environment variables. No secrets in config files.

| Variable                  | Required | Default              | Description                                      |
| ------------------------- | -------- | -------------------- | ------------------------------------------------ |
| `WT_LISTEN_ADDR`          | Yes      | `0.0.0.0:4443`       | Listen address for the QUIC endpoint             |
| `WT_TLS_CERT_FILE`        | Yes      | —                    | Path to TLS certificate (PEM)                    |
| `WT_TLS_KEY_FILE`         | Yes      | —                    | Path to TLS private key (PEM)                    |
| `WT_JWT_SECRET`           | Dev only | —                    | HMAC-SHA256 JWT secret (development only)        |
| `WT_JWT_PUBLIC_KEY_FILE`  | Prod     | —                    | Path to RS256 public key for JWT validation      |
| `WT_REDPANDA_BROKERS`     | Yes      | `localhost:9092`     | Comma-separated Redpanda broker addresses        |
| `WT_TRACK_TOPIC`          | Yes      | `track-updates`      | Redpanda topic for track records                 |
| `WT_MAX_SESSIONS`         | No       | `100`                | Maximum concurrent WebTransport sessions         |
| `WT_ALLOWED_ORIGINS`      | No       | (all)                | Comma-separated allowed origins (empty = allow all, dev only) |
| `WT_CONGESTION_THRESHOLD` | No       | `0.8`                | QUIC send buffer fill ratio to trigger priority shedding |

---

## 4. Deployment

### 4.1 Docker Compose (Development)

```yaml
# deploy/docker-compose.dev.yml (excerpt)
webtransport:
  image: rtsa/webtransport:latest
  ports:
    - "4443:4443/udp"
  environment:
    WT_LISTEN_ADDR: "0.0.0.0:4443"
    WT_TLS_CERT_FILE: "/certs/dev.crt"
    WT_TLS_KEY_FILE: "/certs/dev.key"
    WT_JWT_SECRET: "${WT_JWT_SECRET}"
    WT_REDPANDA_BROKERS: "redpanda:9092"
    WT_TRACK_TOPIC: "track-updates"
  volumes:
    - ./certs:/certs:ro
  depends_on:
    - redpanda
```

### 4.2 Kubernetes (Production)

Deploy using the Helm chart in `deploy/helm/webtransport/`:

```bash
helm upgrade --install rtsa-webtransport deploy/helm/webtransport \
  --namespace rtsa \
  --set tls.certFile=/certs/prod.crt \
  --set tls.keyFile=/certs/prod.key \
  --set jwt.publicKeyFile=/keys/jwt.pub \
  --set redpanda.brokers="redpanda-0.redpanda:9092,redpanda-1.redpanda:9092" \
  --set maxSessions=100
```

### 4.3 QUIC Firewall Rules

QUIC operates over UDP. Ensure the following firewall rules are applied:

```
# Allow QUIC inbound on port 4443
-A INPUT -p udp --dport 4443 -m state --state NEW,ESTABLISHED -j ACCEPT
-A OUTPUT -p udp --sport 4443 -m state --state ESTABLISHED -j ACCEPT
```

> **Note**: Many enterprise firewalls block UDP. If operators report connection failures, verify that UDP/4443 is open through all intermediate firewalls and load balancers.

---

## 5. Health Checks

### HTTP Health Endpoint

The service exposes an HTTP/1.1 health endpoint on port 8081:

```
GET http://host:8081/healthz
```

Returns `200 OK` with body `{"status":"healthy","sessions":N}` when the service is operational.

### Readiness Check

```
GET http://host:8081/readyz
```

Returns `200 OK` when the Redpanda consumer is connected and the TLS certificate is valid.

---

## 6. Monitoring

### Metrics (Prometheus)

The service exports Prometheus metrics on port 9101:

| Metric                                   | Type    | Description                            |
| ---------------------------------------- | ------- | -------------------------------------- |
| `wt_sessions_active`                     | Gauge   | Currently active WebTransport sessions |
| `wt_datagrams_sent_total`                | Counter | Total datagrams sent across all sessions |
| `wt_records_sent_total`                  | Counter | Total track records forwarded          |
| `wt_records_dropped_total`               | Counter | Records dropped (priority shedding)    |
| `wt_auth_failures_total`                 | Counter | JWT validation failures                |
| `wt_session_duration_seconds`            | Histogram | Session lifetime distribution         |
| `wt_datagram_send_duration_seconds`      | Histogram | Per-datagram send latency             |

### Grafana Dashboard

Import `deploy/grafana/webtransport-dashboard.json` for the pre-built dashboard showing:
- Active sessions and throughput
- Priority shedding rate (should be < 1% under normal load)
- Authentication failure rate

### Alerts

| Alert                             | Threshold                        | Action                                     |
| --------------------------------- | -------------------------------- | ------------------------------------------ |
| `WTSessionCount > 95`             | > 95% of max sessions            | Scale out or throttle new connections       |
| `WTAuthFailures > 10/min`         | 10 failures per minute           | Investigate potential token replay attack   |
| `WTDropRate > 5%`                 | Drop rate above 5%               | Check Redpanda consumer lag and QUIC buffer |
| `WTService down`                  | Health check fails 2 consecutive | Page on-call, initiate rollback if needed   |

---

## 7. TLS Certificate Management

### Certificate Requirements

- TLS 1.3 minimum
- Cipher suites: `TLS_AES_256_GCM_SHA384`, `TLS_CHACHA20_POLY1305_SHA256`
- Subject Alternative Names must include the FQDN and the IP address (WebTransport clients validate both)
- Key size: RSA 4096 or ECDSA P-384

### Certificate Rotation

1. Issue a new certificate from the site CA
2. Copy to the server: `scp new.crt new.key server:/certs/`
3. Reload the service (graceful restart supported):
   ```bash
   kubectl rollout restart deployment/rtsa-webtransport -n rtsa
   ```
4. Verify the new certificate is served:
   ```bash
   openssl s_client -connect host:4443 -alpn h3 </dev/null | openssl x509 -noout -dates
   ```

---

## 8. JWT Token Lifecycle

WebTransport sessions are authenticated with short-lived JWT tokens issued by the gRPC cold-path auth service:

1. The operator authenticates via the gRPC `AuthService.Login` RPC.
2. The auth service returns a short-lived JWT (15 minutes TTL in production).
3. The browser presents the JWT in the WebTransport URL: `?token=<jwt>`.
4. The WebTransport service validates the JWT signature, expiry, and clearance claim.
5. The clearance level from the JWT is used for server-side classification filtering.

### JWT Claims

```json
{
  "operator_id": "op-12345",
  "clearance_level": 3,
  "iss": "rtsa-auth-service",
  "aud": "rtsa-webtransport",
  "exp": 1780000000,
  "iat": 1779999100
}
```

> `clearance_level` values: 0=Unknown, 1=Unclassified, 2=Confidential, 3=Secret, 4=TopSecret

---

## 9. Production Cutover Procedure

Refer to `docs/implementation/v4/phase4_hardening_cutover.md §H4-8` for the full cutover plan.

### Pre-Cutover Checklist

- [ ] `web-cop-gpu` smoke test with production data: 30 min sustained 50k tracks
- [ ] All Playwright E2E tests green in CI
- [ ] Visual regression suite green in CI
- [ ] Security audit report signed off
- [ ] WebTransport TLS certificate valid for production domain
- [ ] QUIC firewall rules applied on all production firewalls
- [ ] Grafana dashboard confirmed operational
- [ ] Rollback DNS TTL reduced to 60 seconds (before cutover)

### Cutover Steps

1. Deploy `web-cop-gpu` alongside `web-cop` (parallel operation, 1 week canary)
2. Internal team validates `web-cop-gpu` on production data
3. Switch Envoy routing to `web-cop-gpu`:
   ```bash
   kubectl apply -f deploy/envoy/web-cop-gpu-routing.yaml
   ```
4. Monitor for 24 hours using Grafana dashboards
5. If no issues: decommission `web-cop` (archive React code)

### Rollback

If critical issues discovered:
1. Revert Envoy routing:
   ```bash
   kubectl apply -f deploy/envoy/web-cop-routing.yaml
   ```
2. `web-cop` remains running throughout cutover period — no data migration needed
3. Open an incident ticket and report to team lead

---

## 10. Troubleshooting

| Symptom                                  | Cause                                           | Resolution                                               |
| ---------------------------------------- | ----------------------------------------------- | -------------------------------------------------------- |
| Clients cannot connect via WebTransport  | UDP/4443 blocked by firewall                   | Open UDP/4443 on all firewall tiers                      |
| `401 Unauthorized` in browser console   | Expired or invalid JWT token                    | Re-issue token via auth service; check server clock sync |
| High drop rate (`wt_records_dropped`)    | QUIC congestion or Redpanda consumer lag        | Scale WebTransport replicas; check Redpanda health       |
| Tracks not updating despite connection  | SAB not being updated by Data Worker            | Check Data Worker logs in browser DevTools console       |
| Visual glitches at 50k tracks            | GPU buffer upload bottleneck                    | Profile with Chrome DevTools GPU tab; check FrameTimer   |
| Certificate error in browser            | Mismatched SAN or expired certificate           | Rotate certificate; verify FQDN matches                  |
| SharedArrayBuffer unavailable           | Missing COOP/COEP headers on frontend           | Verify Envoy adds COOP/COEP headers in routing config    |
