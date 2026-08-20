# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy go.mod first for layer caching.
COPY go.mod ./

# Download dependencies (none in this case, but keeps the pattern).
RUN go mod download

# Copy source code.
COPY *.go ./

# Build a static binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /title-validator .

# ============================================================
# Stage 2: Runtime
# ============================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from builder.
COPY --from=builder /title-validator .

# Switch to non-root user.
USER appuser

# Expose the default port.
EXPOSE 8080

# Health check.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./title-validator"]
