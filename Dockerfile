# Build stage
FROM golang:1.25.1-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=docker
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/svanhalla/prompt-lab/greetd/internal/version.Version=${VERSION} \
              -X github.com/svanhalla/prompt-lab/greetd/internal/version.Commit=${COMMIT} \
              -X github.com/svanhalla/prompt-lab/greetd/internal/version.BuildTime=${BUILD_TIME}" \
    -o greetd ./cmd/greetd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S greetd && \
    adduser -u 1001 -S greetd -G greetd

WORKDIR /home/greetd

# Copy binary from builder stage
COPY --from=builder /app/greetd .

# Create data directory
RUN mkdir -p .greetd && chown -R greetd:greetd .greetd

# Switch to non-root user
USER greetd

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./greetd health || exit 1

# Run the application
CMD ["./greetd", "api"]
