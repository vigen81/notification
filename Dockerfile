FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o notification-engine ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/notification-engine .

# Copy config and translations
COPY --from=builder /app/config ./config
COPY --from=builder /app/translations ./translations

EXPOSE 8080

CMD ["./notification-engine"]
