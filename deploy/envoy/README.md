<!-- CLASSIFICATION: UNCLASSIFIED -->

# Envoy API Gateway — RTSA COP Web Application

Envoy serves as the single entry point for the COP Web Application. It terminates mTLS from operator workstations, translates gRPC-Web to native gRPC, enforces rate limits, and routes requests to backend services.

---

## Configuration Files

| File | Purpose |
|---|---|
| `envoy.yaml` | Production — full mTLS, TLS 1.3 only, CSE-approved ciphers, rate limiting |
| `envoy-dev.yaml` | Development — relaxed (no client certs, plaintext upstream, localhost CORS) |

---

## Port Assignments

| Port | Purpose |
|---|---|
| `8443` | gRPC-Web listener (TLS) |
| `9901` | Envoy admin interface (production: localhost only) |

---

## Route Summary

| gRPC Service | Route Prefix | Backend | Timeout |
|---|---|---|---|
| TrackService | `/rtsa.entity.v1.TrackService/` | `svc-track:50070` | `0s` (streaming) |
| AlertService | `/rtsa.inference.v1.AlertService/` | `svc-alert:50071` | `0s` (streaming) |
| QueryService | `/rtsa.query.v1.QueryService/` | `svc-query:50072` | `60s` |
| FeedbackService | `/rtsa.feedback.v1.FeedbackService/` | `svc-feedback:50062` | `10s` |
| AuditService | `/rtsa.audit.v1.AuditService/` | `svc-audit:50073` | `60s` |
| Health | `/grpc.health.v1.Health/` | `svc-track:50070` | `5s` |

---

## Security

### Production (`envoy.yaml`)

- **mTLS**: Client certificates required (`require_client_certificate: true`)
- **TLS version**: TLS 1.3 only (`tls_minimum_protocol_version: TLSv1_3`, `tls_maximum_protocol_version: TLSv1_3`)
- **Cipher suites** (CSE-approved per CCCS ITSP.40.062):
  - `TLS_AES_256_GCM_SHA384`
  - `TLS_CHACHA20_POLY1305_SHA256`
- **Upstream connections**: mTLS to all backend services
- **Rate limiting**: 100 req/s per client (429 on breach)
- **Admin interface**: Bound to `127.0.0.1` only
- **Access log**: JSON format, includes `%DOWNSTREAM_PEER_SUBJECT%` for audit trail

### Development (`envoy-dev.yaml`)

- Client certificates **not** required
- TLS on listener, **plaintext** to upstream services
- CORS allows any `localhost` origin
- Admin interface open on `0.0.0.0:9901`
- Rate limiting **disabled**

---

## Certificate Paths

| File | Mount Path | Description |
|---|---|---|
| CA certificate | `/certs/ca.crt` | Root CA for validation |
| Server certificate | `/certs/server.crt` | Envoy TLS server cert |
| Server key | `/certs/server.key` | Envoy TLS server key |
| Client certificate | `/certs/client.crt` | mTLS client cert (prod only) |
| Client key | `/certs/client.key` | mTLS client key (prod only) |

Generate development certificates using:

```bash
./scripts/setup/gen-dev-certs.sh
```

---

## Running with Docker Compose

### Development (default)

```bash
docker compose -f deploy/docker-compose.yml up envoy
```

The dev compose entry mounts `envoy-dev.yaml` as the active config.

### Production

Override the volume mount to use `envoy.yaml` and provide valid certificates:

```bash
docker compose -f deploy/docker-compose.yml \
  -e ENVOY_CONFIG=./envoy/envoy.yaml \
  up envoy
```

Or edit `docker-compose.yml` to replace the volume mount directly.

---

## Access Logging

Production access logs are written to `/var/log/envoy/access.log` in JSON format. Each entry includes:

- `timestamp` — ISO-8601 request start time
- `method` — HTTP method
- `path` — gRPC method path
- `response_code` — HTTP response code
- `grpc_status` — gRPC status code
- `duration_ms` — Request duration in milliseconds
- `upstream_host` — Backend host that handled the request
- `downstream_peer_subject` — X.509 subject of the operator client certificate (audit trail)
- `request_id` — Unique request identifier
- `trace_id` — Distributed trace ID (B3 format)

---

## Health Checks

Envoy performs gRPC health checks on all upstream clusters every 10 seconds. A cluster is marked unhealthy after 3 consecutive failures and healthy after 1 success.

To verify Envoy itself is running, query the admin endpoint:

```bash
curl http://localhost:9901/ready
```

To inspect cluster health:

```bash
curl http://localhost:9901/clusters
```
