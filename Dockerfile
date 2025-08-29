# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o anybase cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S anybase && \
    adduser -u 1000 -S anybase -G anybase

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/anybase .

# Copy config file if exists (optional, can be mounted)
COPY --from=builder /app/config.yaml* ./

# Change ownership
RUN chown -R anybase:anybase /app

# Switch to non-root user
USER anybase

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./anybase"]