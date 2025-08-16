# Stage 1: Build
FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

# Stage 2: Run
FROM debian:bullseye

WORKDIR /app
COPY --from=builder /app/server .

# Install SSL certificates (kalau aplikasi call HTTPS API)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

EXPOSE 8080
CMD ["./server"]
