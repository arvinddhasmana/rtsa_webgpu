# CLASSIFICATION: UNCLASSIFIED
# Dockerfile for web-cop React application

FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json ./
RUN npm install --frozen-lockfile

COPY . .
RUN npm run build

FROM nginx:alpine AS production

COPY --from=builder /app/dist /usr/share/nginx/html
COPY --from=builder /app/public/sw.js /usr/share/nginx/html/sw.js

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
