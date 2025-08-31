# Multi-stage build for optimal image size
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the applications
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gotasker-server cmd/server/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o gotasker-cli cmd/cli/main.go

# Final stage - minimal image
FROM alpine:latest

# Install runtime dependencies and security updates
RUN apk --no-cache add \
    ca-certificates \
    sqlite-libs \
    tzdata \
    wget \
    && update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1001 -S gotask && \
    adduser -u 1001 -S gotask -G gotask

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/gotasker-server ./gotask
COPY --from=builder /app/gotasker-cli ./gotask-cli

# Copy configuration and API documentation
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/api ./api

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/logs && \
    chown -R gotask:gotask /app && \
    chmod +x ./gotask ./gotask-cli

# Switch to non-root user
USER gotask

# Set environment variables
ENV STORAGE_TYPE=json
ENV STORAGE_FILE_PATH=/app/data/tasks.json
ENV PORT=8080
ENV LOG_LEVEL=info

# Expose port
EXPOSE 8080

# Health check with improved configuration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["./gotask"]