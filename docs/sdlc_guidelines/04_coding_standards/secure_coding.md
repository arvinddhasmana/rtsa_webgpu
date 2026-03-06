# Secure Coding Standards

> **CLASSIFICATION: UNCLASSIFIED**
> **Document Type**: Coding Standard
> **Parent**: `04_coding_standards/general_coding.md`
> **Compliance**: ITSG-33, NIST 800-53, OWASP
> **Last Updated**: 2026-02-23

---

## 1. Purpose

This document defines secure coding practices for all RTSA code. Given the Protected C / Secret classification of the operational environment, every line of code is a potential attack surface. These standards apply to Go, TypeScript, Protobuf, SQL, and configuration files.

## 2. Input Validation

### 2.1 Core Principle

ALL external data is untrusted. This includes:
- Sensor feeds (any of the 6 sensor types)
- Operator input (feedback, queries, configuration)
- NATO data exchange (STANAG 5516, NFFI, MIP messages)
- gRPC API payloads from any client
- Environment variables (validate at startup)
- Configuration files

### 2.2 Validation Rules

| Data Type | Validation | Example |
|---|---|---|
| Coordinates (latitude) | Range: `[-90.0, 90.0]` | `if lat < -90 || lat > 90 { reject }` |
| Coordinates (longitude) | Range: `[-180.0, 180.0]` | `if lon < -180 || lon > 180 { reject }` |
| Altitude (meters MSL) | Range: `[-500.0, 100000.0]` | Reasonable operational range |
| Speed (m/s) | Range: `[0.0, 10000.0]` | Up to ~Mach 30 for space objects |
| Heading (degrees) | Range: `[0.0, 360.0)` | Circular range |
| Timestamp | Range: `[2020-01-01, now + 1 hour]` | Reject far-future or ancient timestamps |
| String fields | Max length; UTF-8 valid; no control characters | `if len(s) > maxLen { reject }` |
| Enum fields | Must be a defined value; reject UNSPECIFIED for required fields | `if val == UNSPECIFIED { reject }` |
| UUIDs | RFC 4122 format validation | `uuid.Validate(id)` |
| Protobuf messages | Validate against schema; check required fields present | Schema-level validation |

### 2.3 Go Input Validation Pattern

```go
// CLASSIFICATION: UNCLASSIFIED

func ValidateSensorEvent(event *pb.SensorEvent) error {
    if event == nil {
        return fmt.Errorf("event is nil")
    }
    if event.GetSensorId() == "" {
        return fmt.Errorf("sensor_id is required")
    }
    if event.GetSensorType() == pb.SENSOR_TYPE_UNSPECIFIED {
        return fmt.Errorf("sensor_type must be specified")
    }
    pos := event.GetPosition()
    if pos != nil {
        if pos.GetLatitudeDeg() < -90 || pos.GetLatitudeDeg() > 90 {
            return fmt.Errorf("latitude out of range: %f", pos.GetLatitudeDeg())
        }
        if pos.GetLongitudeDeg() < -180 || pos.GetLongitudeDeg() > 180 {
            return fmt.Errorf("longitude out of range: %f", pos.GetLongitudeDeg())
        }
    }
    ts := event.GetEventTime()
    if ts != nil {
        eventTime := time.Unix(ts.GetSeconds(), int64(ts.GetNanos()))
        if eventTime.Before(minAllowedTime) || eventTime.After(time.Now().Add(1*time.Hour)) {
            return fmt.Errorf("event_time out of allowed range: %v", eventTime)
        }
    }
    return nil
}
```

## 3. Output Encoding

### 3.1 gRPC Responses

- Always return well-formed Protobuf messages — never raw bytes
- Include classification marking in response messages
- Strip internal error details in production (return generic `INTERNAL` for unexpected errors)
- Never include stack traces in error responses

### 3.2 Web Responses (SolidJS / WebGPU COP)

- HTML-encode all user-generated content rendered in the DOM
- SolidJS auto-escapes text in JSX (never use `innerHTML` directive with untrusted data)
- Set Content-Security-Policy headers to prevent XSS (include `'wasm-unsafe-eval'` for WebAssembly)
- CORS: restrict to same-origin; no wildcard origins
- Cross-origin isolation: set COOP/COEP headers for SharedArrayBuffer support

## 4. Cryptographic Standards

### 4.1 CSE-Approved Algorithms Only

| Purpose | Algorithm | Parameters |
|---|---|---|
| Symmetric encryption | AES-256-GCM | 256-bit key, 96-bit nonce, authenticated |
| Hashing | SHA-256, SHA-384, SHA-512 | No SHA-1 or MD5 |
| Digital signatures | ECDSA with P-384 | Or RSA-3072+ (prefer ECDSA) |
| Key exchange | ECDH with P-384 or X25519 | Perfect forward secrecy required |
| Password hashing | N/A | No password-based auth; use certificate-based |
| MAC | HMAC-SHA-256 | For message authentication |

### 4.2 Prohibited

- MD5 for any purpose (including checksums for security decisions)
- SHA-1 for any security purpose
- DES, 3DES, RC4, Blowfish
- ECB mode for any block cipher
- Custom or proprietary cryptographic algorithms
- Hardcoded cryptographic keys, IVs, or salts
- Use of `math/rand` for security-sensitive operations (use `crypto/rand`)

### 4.3 Implementation

```go
// GOOD — using crypto/rand for security-sensitive operations
import "crypto/rand"

func generateNonce() ([]byte, error) {
    nonce := make([]byte, 12)
    if _, err := rand.Read(nonce); err != nil {
        return nil, fmt.Errorf("generateNonce: %w", err)
    }
    return nonce, nil
}

// BAD — using math/rand for anything security-related
import "math/rand"
nonce := rand.Int()  // PROHIBITED for security
```

## 5. Secret Management

### 5.1 Rules

- **NEVER** hardcode secrets in source code, configuration files, or environment-specific configs
- Store secrets in Kubernetes Secrets (encrypted at rest) or HashiCorp Vault
- Rotate secrets on a defined schedule (certificates: 90 days; API keys: 180 days)
- Access secrets at runtime via environment variables or mounted volumes
- CI/CD pipeline secrets use GitHub Actions secrets or equivalent

### 5.2 Detection

Pre-commit hooks and CI pipeline scan for accidentally committed secrets:

| Pattern | Tool | Action |
|---|---|---|
| Hardcoded passwords | `gitleaks` | Block commit / Block merge |
| API keys / tokens | `gitleaks` | Block commit / Block merge |
| Private keys | `gitleaks` | Block commit / Block merge |
| Connection strings | `semgrep` rules | Block merge |

## 6. gRPC Security — mTLS Configuration

### 6.1 TLS Requirements

- TLS 1.3 ONLY (no fallback to 1.2)
- Mutual TLS (mTLS) on ALL gRPC channels
- Certificate-based authentication via project PKI

### 6.2 Go mTLS Server Setup

```go
// CLASSIFICATION: UNCLASSIFIED

// ITSG-33: SC-8, SC-13 — Transmission confidentiality and cryptographic protection

func newTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("load server cert: %w", err)
    }

    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, fmt.Errorf("read CA cert: %w", err)
    }

    caPool := x509.NewCertPool()
    if !caPool.AppendCertsFromPEM(caCert) {
        return nil, fmt.Errorf("failed to parse CA certificate")
    }

    return &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caPool,
        MinVersion:   tls.VersionTLS13,
        CipherSuites: nil, // TLS 1.3 cipher suites are fixed; no config needed
    }, nil
}
```

## 7. SQL Injection Prevention (ClickHouse)

### 7.1 Rules

- **ALWAYS** use parameterized queries — never concatenate user input into SQL
- Use the ClickHouse Go client's prepared statement API
- Validate query parameters before execution
- Restrict query complexity (max rows, max execution time)

### 7.2 Example

```go
// GOOD — parameterized query
rows, err := ch.Query(ctx,
    `SELECT entity_id, detection_time, entity_type
     FROM entity_tracks
     WHERE sensor_type = @sensorType
       AND detection_time BETWEEN @startTime AND @endTime
     LIMIT @limit`,
    clickhouse.Named("sensorType", sensorType),
    clickhouse.Named("startTime", startTime),
    clickhouse.Named("endTime", endTime),
    clickhouse.Named("limit", maxRows),
)

// BAD — string concatenation
query := fmt.Sprintf("SELECT * FROM entity_tracks WHERE sensor_type = '%s'", sensorType)
rows, err := ch.Query(ctx, query) // SQL INJECTION VULNERABILITY
```

## 8. Denial of Service Protection

### 8.1 Rate Limiting

- Per-sensor rate limits at the ingestion layer
- Per-operator rate limits for feedback submission
- Per-query complexity limits for ClickHouse queries
- gRPC connection limits per source IP

### 8.2 Resource Limits

```go
// gRPC server options for DoS protection
grpc.NewServer(
    grpc.MaxRecvMsgSize(4 * 1024 * 1024),    // 4MB max message
    grpc.MaxSendMsgSize(4 * 1024 * 1024),
    grpc.MaxConcurrentStreams(100),
    grpc.ConnectionTimeout(30 * time.Second),
    grpc.KeepaliveParams(keepalive.ServerParameters{
        MaxConnectionIdle: 5 * time.Minute,
        Time:              1 * time.Minute,
        Timeout:           20 * time.Second,
    }),
)
```

## 9. Anti-Poisoning Safeguards

### 9.1 Feedback Trust Scoring

All operator feedback that influences model retraining must be trust-scored before acceptance. Trust scoring defends against adversarial or erroneous feedback corrupting model behavior.

**Trust score design best practices:**
- Compute a composite trust score from multiple independent factors
- Common factors to incorporate:
  - **Operator authority level** — higher access/clearance yields a higher base trust
  - **Historical accuracy** — track record of the feedback provider against verified ground truth
  - **Temporal consistency** — whether feedback timing is plausible given the event timeline
  - **Statistical deviation** — how far the feedback diverges from the model's current consensus or from peer feedback
- Weight factors based on domain analysis — no single factor should dominate
- Normalize the composite score to a [0.0, 1.0] range for consistent thresholding

### 9.2 Trust Score Action Tiers

Define configurable threshold tiers (do not hard-code specific values into business logic):

| Tier | Action |
|---|---|
| **High trust** | Auto-accept for training pipeline |
| **Moderate trust** | Accept with logging; include in training batch |
| **Low trust** | Flag for human review; exclude from automated training |
| **Very low trust** | Reject; alert Security Operations; trigger audit investigation |

**Best practices:**
- Store threshold values in configuration (environment variables / config files) — not in code
- Log all trust score computations for auditability
- Monitor the distribution of trust scores over time — sudden shifts may indicate coordinated poisoning
- Implement rate limiting per feedback provider to prevent volume-based attacks

## 10. AI Agent Instructions

When generating any code:

1. Validate ALL external input — coordinates, timestamps, strings, enums
2. Never concatenate user input into SQL queries — always use parameterized queries
3. Never hardcode secrets — use environment variables
4. Use `crypto/rand` (not `math/rand`) for any security-sensitive random values
5. TLS 1.3 with mTLS for all gRPC connections — include setup boilerplate
6. Set resource limits on gRPC servers (max message size, concurrent streams)
7. Include trust scoring logic for any feedback path
8. Never return internal error details in production responses
9. Use SolidJS auto-escaping for XSS protection — never use `innerHTML` directive with untrusted data
10. Reference ITSG-33/NIST controls in security-critical code comments
