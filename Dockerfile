# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cli-auth ./cmd/cli-auth/

# Runtime stage
FROM alpine:3.20

# Create data directory for session file
RUN mkdir -p /data

COPY --from=builder /cli-auth /usr/local/bin/cli-auth

ENTRYPOINT ["cli-auth"]
