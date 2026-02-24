<!-- CLASSIFICATION: UNCLASSIFIED -->

# Envoy API Gateway — RTSA

Envoy serves as the single entry point for the COP Web Application. It terminates mTLS from operator workstations, translates gRPC-Web to native gRPC, enforces rate limits, and routes to backend services.

---

## Route Table

| gRPC Service    | Route Prefix                         | Backend          | Port  | Timeout        |
|-----------------|--------------------------------------|------------------|-------|----------------|
| TrackService    | `/rtsa.entity.v1.TrackService/`      | svc-track        | 50070 | Streaming (0s) |
| AlertService    | `/rtsa.inference.v1.AlertService/`   | svc-alert        | 50071 | Streaming (0s) |
| QueryService    | `/rtsa.query.v1.QueryService/`       | svc-query        | 50072 | 60s            |
| FeedbackService | `/rtsa.feedback.v1.FeedbackService/` | svc-feedback     | 50062 | 10s            |
| AuditService    | `/rtsa.audit.v1.AuditService/`       | svc-audit        | 50073 | 60s            |
| HealthCheck     | `/grpc.health.v1.Health/`            | svc-track        | 50070 | 5s             |

---

## Security Comparison

| Aspect              | Production (`envoy.yaml`)          | Development (`envoy-dev.yaml`) |
|---------------------|------------------------------------|--------------------------------|
| Client certificates | Required (mTLS)                    | Not required                   |
| TLS version         | TLS 1.3 only                       | TLS 1.3 only                   |
| Cipher suites       | AES-256-GCM, ChaCha20              | Default                        |
| CORS origins        | `https://cop.rtsa.local`           | `localhost:*`                  |
| Rate limiting       | 100 req/s per client               | Disabled                       |
| Admin interface     | `127.0.0.1:9901` (loopback only)   | `0.0.0.0:9901` (open)          |
| Upstream TLS        | mTLS to all backends               | Plaintext h2c                  |
| Access logging      | JSON file with peer cert subject   | Not configured                 |

---

## Certificate Requirements

### Production

```
/certs/
├── ca.crt          # CA certificate (for validating clients and servers)
├── server.crt      # Envoy server certificate
├── server.key      # Envoy server private key
├── client.crt      # Envoy client certificate (for upstream mTLS)
└── client.key      # Envoy client private key
```

### Development

```
certs/dev/
├── ca.crt          # Self-signed CA
├── server.crt      # Self-signed server certificate
└── server.key      # Server private key
```

---

## Development Quick Start

```bash
# Start the full dev stack (includes Envoy)
docker compose -f deploy/docker-compose.yml up -d

# Verify Envoy is running
curl -k https://localhost:8443/grpc.health.v1.Health/Check

# Check admin stats
curl http://localhost:9901/stats

# View clusters status
curl http://localhost:9901/clusters
```

---

## Production Deployment

```bash
# Mount production certificates
docker run -d \
  --name rtsa-envoy \
  -p 8443:8443 \
  -v /path/to/prod/certs:/certs:ro \
  -v $(pwd)/deploy/envoy/envoy.yaml:/etc/envoy/envoy.yaml:ro \
  envoyproxy/envoy:v1.30-latest

# Verify admin is NOT externally accessible
# Admin is bound to 127.0.0.1:9901 only
```

---

## Test Scenarios

| #   | Test                             | Command                                                                 | Expected                    |
|-----|----------------------------------|-------------------------------------------------------------------------|-----------------------------|
| T01 | gRPC-Web to TrackService         | `grpcurl -H 'Content-Type: application/grpc-web' ...`                  | Proxied to svc-track:50070  |
| T02 | gRPC-Web to AlertService         | `grpcurl -H 'Content-Type: application/grpc-web' ...`                  | Proxied to svc-alert:50071  |
| T03 | CORS preflight OPTIONS           | `curl -X OPTIONS -H 'Origin: https://cop.rtsa.local' https://...`      | Returns CORS headers        |
| T04 | Invalid route prefix             | `curl https://localhost:8443/invalid/`                                  | 404 Not Found               |
| T05 | Rate limit exceeded              | Send >100 req/s                                                         | 429 Too Many Requests       |
| T06 | Health check endpoint            | `curl -k https://localhost:8443/grpc.health.v1.Health/Check`           | Returns health status       |
| T07 | mTLS: valid client cert (prod)   | `curl --cert client.crt --key client.key --cacert ca.crt https://...`  | Connection accepted         |
| T08 | mTLS: no client cert (prod)      | `curl --cacert ca.crt https://...`                                      | Connection rejected         |

---

## Troubleshooting

**Envoy won't start**
```bash
docker logs rtsa-envoy
# Check: certificate files exist and are readable
# Check: envoy.yaml syntax with `envoy --mode validate -c /etc/envoy/envoy.yaml`
```

**503 on upstream**
```bash
curl http://localhost:9901/clusters | grep health
# Check upstream service is running and listening on correct port
```

**TLS handshake failure**
```bash
openssl s_client -connect localhost:8443 -CAfile certs/dev/ca.crt
# Check certificate chain and expiry
```
