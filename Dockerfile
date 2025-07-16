# Updated Dockerfile for AWS Parameter Store Configuration
FROM 499144353299.dkr.ecr.eu-central-1.amazonaws.com/docker-hub/library/golang:1.24-alpine AS build

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
RUN go build -o app ./cmd/server


FROM scratch

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs /etc/ssl/certs
COPY --from=build /app/app /app

ENTRYPOINT ["/app"]