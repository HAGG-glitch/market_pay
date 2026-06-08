# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod ./
# Download deps - will create go.sum automatically
RUN go mod download || true

COPY . .

# Tidy to ensure go.sum is complete
RUN go mod tidy

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=mod -ldflags="-w -s" -o /out/api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=mod -ldflags="-w -s" -o /out/worker ./cmd/worker

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=mod -ldflags="-w -s" -o /out/migrate ./cmd/migrate

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

COPY --from=builder /out/api     ./api
COPY --from=builder /out/worker  ./worker
COPY --from=builder /out/migrate ./migrate

COPY configs/ ./configs/
COPY migrations/ ./migrations/

EXPOSE 8080 9090

HEALTHCHECK --interval=15s --timeout=5s --start-period=30s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./api"]
