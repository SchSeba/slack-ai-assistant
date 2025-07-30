# Multi-stage Dockerfile for Slack AI Assistant

# Build stage
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_TIME=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=1 \
    GOOS=linux \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags "-linkmode external -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT_HASH} -X main.buildTime=${BUILD_TIME}" \
    -o slack-ai-assistant \
    ./cmd/server

# Final stage
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/slack-ai-assistant /slack-ai-assistant

# Create a directory for the database
VOLUME ["/data"]

# Set working directory
WORKDIR /data

# Expose any necessary ports (none needed for this app)

# Labels
LABEL maintainer="SchSeba"
LABEL app="slack-ai-assistant"
LABEL description="A Slack AI Assistant Bot that integrates with AnythingLLM"

# Health check (optional - requires implementing health endpoint)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/slack-ai-assistant", "health"] || exit 1

# Run as non-root user for security
USER 65534:65534

# Entry point
ENTRYPOINT ["/slack-ai-assistant"]

# Default command
CMD ["--help"]