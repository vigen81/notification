# Updated Dockerfile for AWS Parameter Store Configuration
FROM golang:1.24.4-alpine AS builder

# Install dependencies including Git and build tools
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install swag for swagger generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy source code
COPY . .

# Generate Ent code first
RUN go generate ./ent

# Create empty docs package to satisfy import
RUN mkdir -p docs && echo "package docs" > docs/docs.go

# Generate Swagger docs (this will overwrite the empty docs.go)
RUN swag init -g cmd/server/main.go -o docs/

# Tidy modules after all generation
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-engine ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/notification-engine .

# Create non-root user
RUN addgroup -g 1001 -S notifier && \
    adduser -S notifier -u 1001 -G notifier

USER notifier

EXPOSE 8080

# Environment variables for AWS Parameter Store
ENV AWS_REGION=eu-central-1

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./notification-engine"]