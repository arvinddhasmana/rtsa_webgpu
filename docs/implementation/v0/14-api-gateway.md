<!-- CLASSIFICATION: UNCLASSIFIED -->

# Module 14 — API Gateway (Envoy)

> **Module**: 14-api-gateway
> **Phase**: P3 (Presentation)
> **Dependencies**: Module 01 (infrastructure), Module 02 (protos)
> **Agent**: `@greatest-ever-developer`
> **Estimated Effort**: 2 days

---

## 1. Objective

Configure the Envoy-based API Gateway that serves as the single entry point for the COP Web Application. Envoy terminates mTLS from operator workstations, translates gRPC-Web to native gRPC, enforces rate limits, and routes to backend services.

**This module produces configuration files only — no Go code.** Envoy is a COTS proxy.

**Acceptance Criteria**:

- Envoy front-proxy listening on port 8443 (TLS)
- gRPC-Web → gRPC translation for browser clients
- mTLS with operator X.509 certificates (TLS 1.3 only)
- Routes to Track Service (50070), Alert Service (50071), Query Service (50072), Audit Service (50073), Feedback Service (50062)
- CORS configuration for COP Web App origin
- Rate limiting per operator (100 req/s default)
- Health check endpoints for all upstream services
- Access logging for audit trail

---

## 2. File Structure

```
deploy/
├── envoy/
│   ├── envoy.yaml                  # Main Envoy configuration
│   ├── envoy-dev.yaml              # Development overrides (relaxed mTLS)
│   └── README.md
├── docker-compose.yml              # (already exists — add envoy service)
```

---

## 3. Envoy Configuration

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/envoy/envoy.yaml
# Envoy API Gateway for RTSA COP Web Application

admin:
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 9901
  access_log:
    - name: envoy.access_loggers.file
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
        path: /dev/null # Admin access log disabled for security

static_resources:
  listeners:
    - name: grpc_web_listener
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 8443
      filter_chains:
        - transport_socket:
            name: envoy.transport_sockets.tls
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
              require_client_certificate: true
              common_tls_context:
                tls_params:
                  tls_minimum_protocol_version: TLSv1_3
                  tls_maximum_protocol_version: TLSv1_3
                  cipher_suites:
                    - TLS_AES_256_GCM_SHA384
                    - TLS_CHACHA20_POLY1305_SHA256
                tls_certificates:
                  - certificate_chain:
                      filename: /certs/server.crt
                    private_key:
                      filename: /certs/server.key
                validation_context:
                  trusted_ca:
                    filename: /certs/ca.crt
          filters:
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                codec_type: AUTO
                stat_prefix: grpc_web_ingress
                access_log:
                  - name: envoy.access_loggers.file
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
                      path: /var/log/envoy/access.log
                      log_format:
                        json_format:
                          timestamp: "%START_TIME%"
                          method: "%REQ(:METHOD)%"
                          path: "%REQ(:PATH)%"
                          protocol: "%PROTOCOL%"
                          response_code: "%RESPONSE_CODE%"
                          grpc_status: "%GRPC_STATUS%"
                          duration_ms: "%DURATION%"
                          bytes_sent: "%BYTES_SENT%"
                          bytes_received: "%BYTES_RECEIVED%"
                          upstream_host: "%UPSTREAM_HOST%"
                          downstream_peer_subject: "%DOWNSTREAM_PEER_SUBJECT%"
                          request_id: "%REQ(X-REQUEST-ID)%"
                          trace_id: "%REQ(X-B3-TRACEID)%"
                route_config:
                  name: grpc_web_routes
                  virtual_hosts:
                    - name: rtsa_services
                      domains: ["*"]
                      cors:
                        allow_origin_string_match:
                          - exact: "https://cop.rtsa.local"
                          - exact: "https://localhost:3000"
                        allow_methods: "GET, POST, OPTIONS"
                        allow_headers: "content-type, x-grpc-web, x-user-agent, grpc-timeout, rtsa-classification"
                        expose_headers: "grpc-status, grpc-message, grpc-status-details-bin"
                        max_age: "86400"
                        allow_credentials: true
                      routes:
                        # Track Service
                        - match:
                            prefix: "/rtsa.entity.v1.TrackService/"
                          route:
                            cluster: track_service
                            timeout: 0s # Server-streaming: no timeout
                            max_stream_duration:
                              grpc_timeout_header_max: 0s
                        # Alert Service
                        - match:
                            prefix: "/rtsa.inference.v1.AlertService/"
                          route:
                            cluster: alert_service
                            timeout: 0s # Server-streaming: no timeout
                            max_stream_duration:
                              grpc_timeout_header_max: 0s
                        # Query Service
                        - match:
                            prefix: "/rtsa.query.v1.QueryService/"
                          route:
                            cluster: query_service
                            timeout: 60s
                        # Feedback Service
                        - match:
                            prefix: "/rtsa.feedback.v1.FeedbackService/"
                          route:
                            cluster: feedback_service
                            timeout: 10s
                        # Audit Service
                        - match:
                            prefix: "/rtsa.audit.v1.AuditService/"
                          route:
                            cluster: audit_service
                            timeout: 60s
                        # Health Check
                        - match:
                            prefix: "/grpc.health.v1.Health/"
                          route:
                            cluster: track_service
                            timeout: 5s
                http_filters:
                  # gRPC-Web to gRPC translation
                  - name: envoy.filters.http.grpc_web
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb
                  # CORS support
                  - name: envoy.filters.http.cors
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
                  # Rate limiting
                  - name: envoy.filters.http.local_ratelimit
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
                      stat_prefix: http_local_rate_limiter
                      token_bucket:
                        max_tokens: 100
                        tokens_per_fill: 100
                        fill_interval: 1s
                      filter_enabled:
                        runtime_key: local_rate_limit_enabled
                        default_value:
                          numerator: 100
                          denominator: HUNDRED
                      filter_enforced:
                        runtime_key: local_rate_limit_enforced
                        default_value:
                          numerator: 100
                          denominator: HUNDRED
                  # Router (must be last)
                  - name: envoy.filters.http.router
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

  clusters:
    - name: track_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: track_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-track
                      port_value: 50070
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          common_tls_context:
            tls_params:
              tls_minimum_protocol_version: TLSv1_3
            tls_certificates:
              - certificate_chain:
                  filename: /certs/client.crt
                private_key:
                  filename: /certs/client.key
            validation_context:
              trusted_ca:
                filename: /certs/ca.crt
      health_checks:
        - timeout: 5s
          interval: 10s
          unhealthy_threshold: 3
          healthy_threshold: 1
          grpc_health_check: {}

    - name: alert_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: alert_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-alert
                      port_value: 50071
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          common_tls_context:
            tls_params:
              tls_minimum_protocol_version: TLSv1_3
            tls_certificates:
              - certificate_chain:
                  filename: /certs/client.crt
                private_key:
                  filename: /certs/client.key
            validation_context:
              trusted_ca:
                filename: /certs/ca.crt
      health_checks:
        - timeout: 5s
          interval: 10s
          unhealthy_threshold: 3
          healthy_threshold: 1
          grpc_health_check: {}

    - name: query_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: query_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-query
                      port_value: 50072
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          common_tls_context:
            tls_params:
              tls_minimum_protocol_version: TLSv1_3
            tls_certificates:
              - certificate_chain:
                  filename: /certs/client.crt
                private_key:
                  filename: /certs/client.key
            validation_context:
              trusted_ca:
                filename: /certs/ca.crt
      health_checks:
        - timeout: 5s
          interval: 10s
          unhealthy_threshold: 3
          healthy_threshold: 1
          grpc_health_check: {}

    - name: feedback_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: feedback_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-feedback
                      port_value: 50062
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          common_tls_context:
            tls_params:
              tls_minimum_protocol_version: TLSv1_3
            tls_certificates:
              - certificate_chain:
                  filename: /certs/client.crt
                private_key:
                  filename: /certs/client.key
            validation_context:
              trusted_ca:
                filename: /certs/ca.crt
      health_checks:
        - timeout: 5s
          interval: 10s
          unhealthy_threshold: 3
          healthy_threshold: 1
          grpc_health_check: {}

    - name: audit_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: audit_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: svc-audit
                      port_value: 50073
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          common_tls_context:
            tls_params:
              tls_minimum_protocol_version: TLSv1_3
            tls_certificates:
              - certificate_chain:
                  filename: /certs/client.crt
                private_key:
                  filename: /certs/client.key
            validation_context:
              trusted_ca:
                filename: /certs/ca.crt
      health_checks:
        - timeout: 5s
          interval: 10s
          unhealthy_threshold: 3
          healthy_threshold: 1
          grpc_health_check: {}
```

---

## 4. Development Configuration

```yaml
# CLASSIFICATION: UNCLASSIFIED
# deploy/envoy/envoy-dev.yaml
# Development override — relaxed for local testing
# Uses self-signed certs from certs/dev/
# Does NOT require client certificates

admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901

static_resources:
  listeners:
    - name: grpc_web_listener
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 8443
      filter_chains:
        - transport_socket:
            name: envoy.transport_sockets.tls
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext
              require_client_certificate: false # Relaxed for dev
              common_tls_context:
                tls_certificates:
                  - certificate_chain:
                      filename: /certs/server.crt
                    private_key:
                      filename: /certs/server.key
                validation_context:
                  trusted_ca:
                    filename: /certs/ca.crt
          filters:
            # Same HTTP filters as production but with relaxed CORS
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                codec_type: AUTO
                stat_prefix: grpc_web_ingress
                route_config:
                  name: grpc_web_routes
                  virtual_hosts:
                    - name: rtsa_services
                      domains: ["*"]
                      cors:
                        allow_origin_string_match:
                          - prefix: "http://localhost"
                          - prefix: "https://localhost"
                        allow_methods: "GET, POST, OPTIONS"
                        allow_headers: "*"
                        expose_headers: "grpc-status, grpc-message, grpc-status-details-bin"
                        max_age: "86400"
                        allow_credentials: true
                      routes:
                        - match: { prefix: "/rtsa.entity.v1.TrackService/" }
                          route: { cluster: track_service, timeout: 0s }
                        - match: { prefix: "/rtsa.inference.v1.AlertService/" }
                          route: { cluster: alert_service, timeout: 0s }
                        - match: { prefix: "/rtsa.query.v1.QueryService/" }
                          route: { cluster: query_service, timeout: 60s }
                        - match:
                            { prefix: "/rtsa.feedback.v1.FeedbackService/" }
                          route: { cluster: feedback_service, timeout: 10s }
                        - match: { prefix: "/rtsa.audit.v1.AuditService/" }
                          route: { cluster: audit_service, timeout: 60s }
                        - match: { prefix: "/grpc.health.v1.Health/" }
                          route: { cluster: track_service, timeout: 5s }
                http_filters:
                  - name: envoy.filters.http.grpc_web
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb
                  - name: envoy.filters.http.cors
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors
                  - name: envoy.filters.http.router
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

  clusters:
    # Dev clusters use plaintext to simplify local setup
    - name: track_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: track_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address: { address: svc-track, port_value: 50070 }

    - name: alert_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: alert_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address: { address: svc-alert, port_value: 50071 }

    - name: query_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: query_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address: { address: svc-query, port_value: 50072 }

    - name: feedback_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: feedback_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address: { address: svc-feedback, port_value: 50062 }

    - name: audit_service
      connect_timeout: 5s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
          explicit_http_config:
            http2_protocol_options: {}
      load_assignment:
        cluster_name: audit_service
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address: { address: svc-audit, port_value: 50073 }
```

---

## 5. Docker Compose Service Entry

Add the following to `deploy/docker-compose.yml`:

```yaml
envoy:
  image: envoyproxy/envoy:v1.30-latest
  container_name: rtsa-envoy
  ports:
    - "8443:8443"
    - "9901:9901" # Admin interface (dev only)
  volumes:
    - ./envoy/envoy-dev.yaml:/etc/envoy/envoy.yaml:ro
    - ../certs/dev:/certs:ro
  depends_on:
    - svc-track
    - svc-alert
    - svc-query
    - svc-feedback
    - svc-audit
  networks:
    - rtsa-net
  restart: unless-stopped
```

---

## 6. Security Considerations

| Aspect              | Production                         | Development        |
| ------------------- | ---------------------------------- | ------------------ |
| Client certificates | Required (mTLS)                    | Not required       |
| TLS version         | TLS 1.3 only                       | TLS 1.3 only       |
| Cipher suites       | AES-256-GCM, ChaCha20              | Default            |
| CORS origins        | `https://cop.rtsa.local`           | `localhost:*`      |
| Rate limiting       | 100 req/s per client               | Disabled           |
| Admin interface     | Localhost only, no access log      | Open for debugging |
| Access logging      | JSON format with peer cert subject | JSON format        |

---

## 7. Route Summary

| gRPC Service    | Route Prefix                         | Backend Cluster      | Timeout        |
| --------------- | ------------------------------------ | -------------------- | -------------- |
| TrackService    | `/rtsa.entity.v1.TrackService/`      | `svc-track:50070`    | Streaming (0s) |
| AlertService    | `/rtsa.inference.v1.AlertService/`   | `svc-alert:50071`    | Streaming (0s) |
| QueryService    | `/rtsa.query.v1.QueryService/`       | `svc-query:50072`    | 60s            |
| FeedbackService | `/rtsa.feedback.v1.FeedbackService/` | `svc-feedback:50062` | 10s            |
| AuditService    | `/rtsa.audit.v1.AuditService/`       | `svc-audit:50073`    | 60s            |
| HealthCheck     | `/grpc.health.v1.Health/`            | `svc-track:50070`    | 5s             |

---

## 8. Test Scenarios

| #   | Test                             | Expected                    |
| --- | -------------------------------- | --------------------------- |
| T01 | gRPC-Web request to TrackService | Proxied to svc-track:50070  |
| T02 | gRPC-Web request to AlertService | Proxied to svc-alert:50071  |
| T03 | CORS preflight OPTIONS           | Returns proper CORS headers |
| T04 | Invalid route prefix             | 404 Not Found               |
| T05 | Rate limit exceeded              | 429 Too Many Requests       |
| T06 | Health check endpoint            | Returns health status       |
| T07 | mTLS: valid client cert (prod)   | Connection accepted         |
| T08 | mTLS: no client cert (prod)      | Connection rejected         |

---

## 9. Agent Invocation

```
@greatest-ever-developer Implement Module 14 from docs/implementation/14-api-gateway.md

Context:
- Read docs/implementation/00-implementation-overview.md for port assignments
- Read docs/architecture/security_architecture.md §4.1 for TLS requirements
- Read docs/architecture/security_architecture.md §5.1 for cipher suites

Deliverables:
1. deploy/envoy/envoy.yaml (production config with full mTLS)
2. deploy/envoy/envoy-dev.yaml (development config, relaxed)
3. Update deploy/docker-compose.yml with envoy service
4. deploy/envoy/README.md with usage instructions

CRITICAL:
- TLS 1.3 ONLY (no TLS 1.2 fallback)
- CSE-approved cipher suites: TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256
- Server-streaming routes (Track, Alert) need timeout: 0s
- Access logs MUST include downstream peer certificate subject
- gRPC-Web filter MUST be before router filter
```
