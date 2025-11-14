# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install build tools (optional, keeps image small)
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/order-services ./cmd/server

# -------- Runtime --------
FROM alpine:latest
WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy binary only
COPY --from=builder /out/order-services /app/order-services

# Make binary executable
RUN chmod +x /app/order-services

# Service listens on PORT (default via .env/ENV). Expose common default.
EXPOSE 8006

# Optional: allow overriding PORT at runtime
ENV PORT=8006

# Run
CMD ["/bin/sh", "-c", "exec /app/order-services"]

