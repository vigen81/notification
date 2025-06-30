#!/bin/bash

echo "🚀 Setting up Notification Engine..."

# Create directory structure
echo "📁 Creating directory structure..."
mkdir -p cmd/{server,migrate}
mkdir -p internal/{app,config,db,handlers,kafka,logger,middleware,models,providers/{email,sms},repository,server,services,workers}
mkdir -p ent/schema
mkdir -p types
mkdir -p config
mkdir -p docs

# Check for required files
echo "📋 Checking required files..."

required_files=(
    "go.mod"
    "types/address.go"
    "ent/generate.go"
    "ent/schema/notification.go"
    "ent/schema/partnerconfig.go"
    "cmd/server/main.go"
    "cmd/migrate/main.go"
    "internal/app/application.go"
    "internal/config/config.go"
    "internal/db/client.go"
    "internal/handlers/notification_handler.go"
    "internal/handlers/config_handler.go"
    "internal/handlers/health_handler.go"
    "internal/kafka/config.go"
    "internal/kafka/publisher.go"
    "internal/kafka/subscriber.go"
    "internal/logger/logger.go"
    "internal/middleware/auth.go"
    "internal/models/notification.go"
    "internal/models/partner_config.go"
    "internal/providers/interfaces.go"
    "internal/providers/registry.go"
    "internal/providers/manager.go"
    "internal/providers/email/smtp.go"
    "internal/repository/notification_repository.go"
    "internal/repository/partner_config_repository.go"
    "internal/server/fiber_server.go"
    "internal/services/notification_service.go"
    "internal/services/batch_service.go"
    "internal/workers/notification_worker.go"
    "internal/workers/scheduler_worker.go"
    "config/config.yaml"
    "docker-compose.yaml"
    "Dockerfile"
    "Makefile"
)

missing_files=()
for file in "${required_files[@]}"; do
    if [[ ! -f "$file" ]]; then
        missing_files+=("$file")
    fi
done

if [[ ${#missing_files[@]} -gt 0 ]]; then
    echo "❌ Missing files:"
    printf '   %s\n' "${missing_files[@]}"
    echo ""
    echo "Please create the missing files from the artifacts provided."
    exit 1
else
    echo "✅ All required files are present!"
fi

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Generate Ent code
echo "🔄 Generating Ent code..."
go generate ./ent

if [[ $? -eq 0 ]]; then
    echo "✅ Ent code generated successfully!"
else
    echo "❌ Failed to generate Ent code"
    exit 1
fi

# Build the application
echo "🔨 Building application..."
go build -o bin/notification-engine ./cmd/server

if [[ $? -eq 0 ]]; then
    echo "✅ Application built successfully!"
else
    echo "❌ Failed to build application"
    exit 1
fi

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Start infrastructure: docker-compose up -d"
echo "2. Run migrations: make migrate"
echo "3. Start application: make run"
echo "4. Test health: curl http://localhost:8080/health"
echo ""
echo "Available services:"
echo "- Notification Engine: http://localhost:8080"
echo "- Kafka UI: http://localhost:8081"
echo "- MySQL: localhost:3306"