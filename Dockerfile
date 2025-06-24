FROM golang:1.24.4-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first
COPY go.mod go.sum ./

# Download initial dependencies
RUN go mod download

# Copy source code
COPY . .

# Tidy modules after copying source (this will add any missing dependencies)
RUN go mod tidy

# Generate Ent code (this might add more dependencies)
RUN go generate ./ent

# Tidy again after generation to ensure all dependencies are included
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-engine ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/notification-engine .

# Copy config files
COPY --from=builder /app/config ./config

# Create non-root user
RUN addgroup -g 1001 -S notifier && \
    adduser -S notifier -u 1001 -G notifier

USER notifier

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./notification-engine"]