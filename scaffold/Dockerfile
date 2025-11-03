# Multi-stage Dockerfile for hello-world
# Uses Debian slim for compatibility with DuckDB

# Build stage
FROM golang:bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy source code
COPY . .

# Build the production binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/hello-world \
    ./cmd/hello-world/hello-world.go

# Runtime stage
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS support
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user with home directory
RUN groupadd -r appgroup && useradd -r -g appgroup -m -d /home/appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/hello-world .

# Copy configuration files (if they exist)
COPY --from=builder /build/config.yaml ./config.yaml
COPY --from=builder /build/profile.yaml ./profile.yaml

# Copy store directory if it exists (for data files)
COPY --from=builder /build/store ./store

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port if needed (uncomment and adjust as needed)
# EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/app/hello-world"]

# Default command (can be overridden)
CMD ["--log-output", "json", "--log-level", "info"]
