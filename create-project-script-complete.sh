#!/bin/bash

# Create Complete Notification Engine Project
echo "Creating Complete Notification Engine project..."

# Create base directory
mkdir -p notification-engine
cd notification-engine

# Create directory structure
echo "Creating directory structure..."
mkdir -p cmd/server
mkdir -p internal/{app,config,db,handlers,kafka,middleware,models,providers,repository,server,services,workers}
mkdir -p internal/providers/{email,sms,push}
mkdir -p ent/schema
mkdir -p types
mkdir -p translations
mkdir -p config

# Create a function to write files
write_file() {
    local filepath=$1
    local content=$2
    echo "Creating $filepath..."
    cat > "$filepath" << 'EOF'
$content
EOF
}

# Create cmd/server/main.go
cat > cmd/server/main.go << 'EOF'
package main

import (
	"context"
	"log"

	"gitlab.smartbet.am/golang/notification/internal/app"
	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/db"
	"gitlab.smartbet.am/golang/notification/internal/handlers"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/providers"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/server"
	"gitlab.smartbet.am/golang/notification/internal/services"
	"gitlab.smartbet.am/golang/notification/internal/workers"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		// Configuration
		fx.Provide(config.NewConfig),
		fx.Provide(config.NewLogger),
		
		// Database
		fx.Provide(db.NewDatabase),
		fx.Provide(db.NewEntClient),
		
		// Kafka
		fx.Provide(kafka.NewKafkaConfig),
		fx.Provide(kafka.NewPublisher),
		fx.Provide(kafka.NewSubscriber),
		
		// Repositories
		fx.Provide(repository.NewNotificationRepository),
		fx.Provide(repository.NewPartnerConfigRepository),
		fx.Provide(repository.NewTemplateRepository),
		
		// Provider Factory and Managers
		fx.Provide(providers.NewProviderRegistry),
		fx.Provide(providers.NewEmailProviderManager),
		fx.Provide(providers.NewSMSProviderManager),
		fx.Provide(providers.NewPushProviderManager),
		
		// Services
		fx.Provide(services.NewNotificationService),
		fx.Provide(services.NewTemplateService),
		fx.Provide(services.NewLocalizationService),
		fx.Provide(services.NewBatchService),
		fx.Provide(services.NewSchedulerService),
		
		// Handlers
		fx.Provide(handlers.NewNotificationHandler),
		fx.Provide(handlers.NewConfigHandler),
		fx.Provide(handlers.NewTemplateHandler),
		
		// Workers
		fx.Provide(workers.NewNotificationWorker),
		fx.Provide(workers.NewSchedulerWorker),
		
		// Server
		fx.Provide(server.NewFiberServer),
		
		// Application
		fx.Provide(app.NewApplication),
		
		// Lifecycle
		fx.Invoke(func(lifecycle fx.Lifecycle, app *app.Application) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return app.Start(ctx)
				},
				OnStop: func(ctx context.Context) error {
					return app.Stop(ctx)
				},
			})
		}),
	)

	app.Run()
}
EOF

# Create all the internal files using a script
cat > create_internal_files.sh << 'SCRIPT_EOF'
#!/bin/bash

# This script creates all internal files

# Create types/address.go
cat > types/address.go << 'EOF'
package types

import (
	"database/sql/driver"
	"fmt"
)

type Address string

func (a Address) Value() (driver.Value, error) {
	return string(a), nil
}

func (a *Address) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*a = Address(v)
		return nil
	case []byte:
		*a = Address(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Address", src)
	}
}
EOF

# Create models
cat > internal/models/notification.go << 'EOF'
package models

import (
	"encoding/json"
	"time"

	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/ent/notification"
)

type NotificationType string

const (
	TypeEmail NotificationType = "EMAIL"
	TypeSMS   NotificationType = "SMS"
	TypePush  NotificationType = "PUSH"
)

type NotificationStatus string

const (
	StatusActive    NotificationStatus = "ACTIVE"
	StatusCompleted NotificationStatus = "COMPLETED"
	StatusCancel    NotificationStatus = "CANCEL"
	StatusPending   NotificationStatus = "PENDING"
	StatusFailed    NotificationStatus = "FAILED"
)

type NotificationRequest struct {
	RequestID    string                 `json:"request_id"`
	TenantID     int64                  `json:"tenant_id"`
	Type         NotificationType       `json:"type"`
	TemplateID   string                 `json:"template_id,omitempty"`
	Recipients   []string               `json:"recipients"`
	Body         string                 `json:"body,omitempty"`
	Headline     string                 `json:"headline,omitempty"`
	From         string                 `json:"from,omitempty"`
	ReplyTo      string                 `json:"reply_to,omitempty"`
	Tag          string                 `json:"tag,omitempty"`
	ScheduleTS   *int64                 `json:"schedule_ts,omitempty"`
	Meta         *NotificationMeta      `json:"meta,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Locale       string                 `json:"locale,omitempty"`
	BatchID      string                 `json:"batch_id,omitempty"`
}

type NotificationMeta struct {
	Service    string                 `json:"service,omitempty"`
	TemplateID string                 `json:"template_id,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Attachment *Attachment            `json:"attachment,omitempty"`
	Data       json.RawMessage        `json:"data,omitempty"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Disposition string `json:"disposition"`
	Type        string `json:"type"`
}
EOF

echo "Internal files created successfully!"
SCRIPT_EOF

# Make the script executable and run it
chmod +x create_internal_files.sh
./create_internal_files.sh

# Create configuration files
echo "Creating configuration files..."

cat > config/config.yaml << 'EOF'
server:
  port: ":8080"
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 120s

database:
  host: localhost
  port: "3306"
  user: notification_user
  password: notification_pass
  database: notification_db
  max_open_conns: 25
  max_idle_conns: 25
  conn_max_lifetime: 300

kafka:
  brokers:
    - localhost:9092
  consumer_group: notification-engine
  topics:
    notifications: notifications
    events: notification-events
    dead_letter: notifications-dlq

providers:
  sendgrid:
    sandbox_mode: false
  twilio:
    edge_location: sydney
  fcm:
    validate_only: false

localization:
  default_locale: en
  supported_locales:
    - en
    - es
    - fr
    - de
    - ru
    - hy
  translations_path: ./translations

batch_defaults:
  max_batch_size: 100
  flush_interval: 10s
  max_retries: 3
  retry_backoff: 5s
EOF

# Create docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'

services:
  notification-engine:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/config/config.yaml
      - ENVIRONMENT=development
    volumes:
      - ./config:/app/config
      - ./translations:/app/translations
    depends_on:
      - mysql
      - kafka

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: notification_db
      MYSQL_USER: notification_user
      MYSQL_PASSWORD: notification_pass
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1

volumes:
  mysql_data:
EOF

# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# Environment variables
.env
.env.local
.env.*.local

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Build artifacts
dist/
build/

# Logs
*.log

# Database
*.db
*.sqlite

# Generated files
ent/
EOF

echo "✅ Project structure created successfully!"
echo ""
echo "Next steps:"
echo "1. cd notification-engine"
echo "2. Create the remaining internal files manually or use the provided code"
echo "3. go mod init gitlab.smartbet.am/golang/notification"
echo "4. go mod tidy"
echo "5. go generate ./ent"
echo "6. docker-compose up -d"
echo "7. go run ./cmd/server"
