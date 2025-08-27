# Stage 1: Build
# Force use Alpine-based Go image
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Install git, swag, dan dependencies
RUN apk add --no-cache git ca-certificates build-base \
    && go install github.com/swaggo/swag/cmd/swag@latest

# Verify Go version
RUN go version

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate swagger docs (harus ada sebelum go build)
RUN swag init -g cmd/server/main.go

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Stage 2: Run
FROM alpine:3.18

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/server .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

EXPOSE 8080

CMD ["./server"]