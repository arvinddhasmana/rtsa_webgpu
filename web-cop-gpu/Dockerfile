# CLASSIFICATION: UNCLASSIFIED
# web-cop-gpu/Dockerfile
# Multi-stage build: Rust Wasm decoder → Node frontend → Nginx runtime
# Produces the SolidJS + WebGPU Common Operating Picture static bundle.

# ── Stage 1: Build Wasm decoder ──────────────────────────────
FROM rust:1.82-slim AS wasm-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Download pre-built wasm-pack binary — avoids compiling from source (saves ~10 min, no OpenSSL dep)
RUN curl -sSL https://github.com/rustwasm/wasm-pack/releases/download/v0.13.1/wasm-pack-v0.13.1-x86_64-unknown-linux-musl.tar.gz \
    | tar xz -C /tmp && \
    mv /tmp/wasm-pack-v0.13.1-x86_64-unknown-linux-musl/wasm-pack /usr/local/bin/wasm-pack && \
    rm -rf /tmp/wasm-pack-v0.13.1-x86_64-unknown-linux-musl

WORKDIR /build
COPY web-cop-gpu/wasm-decoder/ ./wasm-decoder/

WORKDIR /build/wasm-decoder
RUN wasm-pack build --target web --release --out-dir pkg

# ── Stage 2: Build frontend ─────────────────────────────────
FROM node:20-alpine AS node-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

# Install dependencies first (layer cache)
COPY web-cop-gpu/package.json web-cop-gpu/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy Wasm artifacts into public/wasm/
COPY --from=wasm-builder /build/wasm-decoder/pkg/ ./public/wasm/

# Copy source
COPY web-cop-gpu/ ./

# Copy generated protobuf TypeScript (resolves @gen/* alias: tsconfig maps ../gen/ts/ → /gen/ts/)
COPY gen/ts/ /gen/ts/
# Rollup walks up from /gen/ts/ looking for node_modules — symlink so it finds /app/node_modules
RUN ln -s /app/node_modules /gen/node_modules

RUN pnpm build

# ── Stage 3: Nginx runtime ──────────────────────────────────
FROM nginx:1.27-alpine AS runtime

# Remove default config
RUN rm /etc/nginx/conf.d/default.conf

# Custom Nginx config with SPA fallback + security headers
COPY <<'NGINX_CONF' /etc/nginx/conf.d/rtsa-cop.conf
server {
    listen       80;
    server_name  _;
    root         /usr/share/nginx/html;
    index        index.html;

    # ── Security headers ─────────────────────────────────────
    # Cross-origin isolation — required for SharedArrayBuffer (WebGPU)
    add_header Cross-Origin-Opener-Policy   "same-origin"       always;
    add_header Cross-Origin-Embedder-Policy "require-corp"      always;

    # Defence-in-depth
    add_header X-Content-Type-Options "nosniff"                 always;
    add_header X-Frame-Options        "DENY"                    always;
    add_header Referrer-Policy        "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: https://*.tile.openstreetmap.org; connect-src 'self' https: wss:; worker-src 'self' blob:; font-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'" always;

    # ── Wasm MIME type ───────────────────────────────────────
    types {
        application/wasm wasm;
    }

    # ── SPA fallback ─────────────────────────────────────────
    location / {
        try_files $uri $uri/ /index.html;
    }

    # ── Cache static assets ──────────────────────────────────
    location ~* \.(js|css|wasm|png|jpg|svg|ico|woff2?)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        # Re-add COOP/COEP on cached assets (add_header in location overrides server block)
        add_header Cross-Origin-Opener-Policy   "same-origin"       always;
        add_header Cross-Origin-Embedder-Policy "require-corp"      always;
    }
}
NGINX_CONF

COPY --from=node-builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
    CMD wget -qO- http://localhost:80/ || exit 1
