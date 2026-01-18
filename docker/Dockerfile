# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT_SHA=unknown

# Build application with version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA} -s -w" \
    -o bin/api cmd/api/main.go

# Build migration tool
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-s -w" \
    -o bin/migrate cmd/migrate/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 madabank && \
    adduser -D -u 1000 -G madabank madabank

WORKDIR /home/madabank

# Copy binaries from builder
COPY --from=builder /app/bin/api .
COPY --from=builder /app/bin/migrate .
COPY --from=builder /app/migrations ./migrations

# Copy entrypoint script
COPY scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Set ownership
RUN chown -R madabank:madabank /home/madabank

# Switch to non-root user
USER madabank

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Set entrypoint and default command
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./api"]