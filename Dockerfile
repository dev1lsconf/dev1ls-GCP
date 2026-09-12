# Stage 1: Build the Go application statically
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

# Cache dependencies
COPY go.mod ./
# Download dependencies (if any added later)
RUN go mod download

# Copy source code
COPY . .

# Compile static, stripped binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server

# Stage 2: Minimal, secure production image
FROM alpine:3.20

# Non-root user for security (defense in depth)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy CA certificates for HTTPS requests and the compiled binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /app/server

# Use non-privileged user
USER appuser:appgroup

# Cloud Run default port
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
