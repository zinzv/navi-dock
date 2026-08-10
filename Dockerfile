# syntax=docker/dockerfile:1

FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend-builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY backend/go.mod backend/go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY backend/ ./
COPY --from=frontend-builder /app/frontend/dist ./web
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/navidock ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -H -u 10001 navidock
WORKDIR /app
COPY --from=backend-builder /out/navidock /app/navidock
COPY --from=backend-builder /src/web /app/web
# Default runtime layout under /data (bind-mounted in compose; empty dirs only — no user data baked in)
RUN mkdir -p /data/database /data/assets/icons /data/assets/wallpapers \
  && chown -R navidock:navidock /data /app
USER navidock
ENV SERVER_PORT=7530
EXPOSE 7530
VOLUME ["/data"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:7530/api/health || exit 1
ENTRYPOINT ["/app/navidock"]
