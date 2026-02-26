# CLASSIFICATION: UNCLASSIFIED
# Dockerfile for web-cop React application
# Build context: repo root (to include gen/ts for @gen alias)

FROM node:20-alpine AS builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app

# Install deps first (cache layer)
COPY web-cop/package.json web-cop/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy web-cop source
COPY web-cop/ ./

# Copy generated TypeScript to /gen/ts so the @gen alias resolves correctly
# vite.config.ts: path.resolve(__dirname, "../gen/ts") = /gen/ts
COPY gen/ts /gen/ts

RUN pnpm run build

FROM nginx:alpine AS production

# Custom Nginx config:
#  - index.html: no-cache so browsers always revalidate (picks up new JS asset hashes)
#  - /assets/*:  immutable+1yr cache (Vite content-hashed filenames)
#  - SPA fallback: all routes serve index.html
RUN cat > /etc/nginx/conf.d/default.conf << 'NGINX'
server {
listen 80;
server_name localhost;
root /usr/share/nginx/html;
index index.html;

# SPA routing — always serve index.html for unknown paths
location / {
try_files $uri $uri/ /index.html;
# Never cache index.html
add_header Cache-Control "no-cache, no-store, must-revalidate";
add_header Pragma "no-cache";
add_header Expires "0";
}

# Vite content-hashed assets — safe to cache forever
location /assets/ {
add_header Cache-Control "public, max-age=31536000, immutable";
}

error_page 500 502 503 504 /50x.html;
location = /50x.html {
root /usr/share/nginx/html;
}
}
NGINX

COPY --from=builder /app/dist /usr/share/nginx/html
COPY --from=builder /app/public/sw.js /usr/share/nginx/html/sw.js

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
