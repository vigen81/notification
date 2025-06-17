Notification Engine

A high-performance, multi-tenant notification engine built with Go, Ent, Fiber, Kafka, and Uber FX.

## Architecture

The notification engine follows a microservices architecture pattern:

1. **HTTP API** receives notification requests
2. **Kafka** queues all requests for async processing
3. **Workers** consume from Kafka, store in DB, then send notifications
4. **Scheduler** handles delayed/scheduled notifications

## Features

- **Multi-channel Support**: Email, SMS, and Push notifications
- **Multi-tenant Architecture**: Per-partner configuration and isolation
- **Provider Flexibility**: Support for multiple providers per channel (SendGrid, SendX, Twilio, FCM)
- **Batch Processing**: Efficient batch sending capabilities
- **Template Engine**: Dynamic templates with i18n support
- **Async Processing**: Kafka-based message queue for reliability
- **Scheduled Notifications**: Support for future-dated notifications
- **REST & Kafka APIs**: Dual API interfaces for flexibility

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
- Apache Kafka

### Installation

1. Clone the repository
```bash
git clone <repository-url>
cd notification-engine
```

2. Install dependencies
```bash
go mod download
```

3. Generate Ent code
```bash
make generate
```

4. Start infrastructure services
```bash
docker-compose up -d mysql kafka zookeeper
```

5. Run database migrations
```bash
make migrate
```

6. Start the application
```bash
make run
```

## Configuration

Copy `config/config.yaml` and adjust settings as needed. Key configurations:

- Database connection settings
- Kafka broker addresses
- Provider default settings
- Batch processing parameters
- Localization settings

## API Usage

### Send Single Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "EMAIL",
    "template_id": "welcome",
    "recipients": ["user@example.com"],
    "data": {
      "name": "John Doe"
    },
    "locale": "en"
  }'
```

### Send Batch Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications/batch \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "SMS",
    "template_id": "promo",
    "recipients": ["+1234567890", "+0987654321"],
    "data": {
      "discount": "20%"
    }
  }'
```

### Check Notification Status

```bash
curl -X GET http://localhost:8080/api/v1/notifications/status/{request_id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Direct Kafka Publishing

```bash
curl -X POST http://localhost:8080/api/v1/kafka/publish \
  -H "X-Kafka-API-Key: YOUR_KAFKA_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 123,
    "type": "EMAIL",
    "recipients": ["user@example.com"],
    "body": "Hello World"
  }'
```

## Partner Configuration

Each tenant/partner can have their own provider configuration:

```bash
curl -X PUT http://localhost:8080/api/v1/config \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_providers": [{
      "name": "primary",
      "type": "sendgrid",
      "priority": 1,
      "enabled": true,
      "config": {
        "api_key": "SG.xxx",
        "from_email": "noreply@company.com",
        "from_name": "Company Name"
      }
    }],
    "batch_config": {
      "enabled": true,
      "max_batch_size": 100,
      "flush_interval": 10
    }
  }'
```

## Template Management

Create reusable templates with dynamic content:

```bash
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "welcome",
    "type": "EMAIL",
    "subject": "Welcome {{.name}}!",
    "body": "Hello {{.name}}, welcome to our service!"
  }'
```

## Development

### Project Structure

```
notification-engine/
├── cmd/server/          # Application entry point
├── internal/
│   ├── app/            # Application lifecycle
│   ├── config/         # Configuration management
│   ├── db/             # Database client
│   ├── handlers/       # HTTP handlers
│   ├── kafka/          # Kafka pub/sub
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Domain models
│   ├── providers/      # Notification providers
│   ├── repository/     # Data access layer
│   ├── server/         # HTTP server
│   ├── services/       # Business logic
│   └── workers/        # Background workers
├── ent/
│   └── schema/         # Ent schemas
├── types/              # Custom types
├── translations/       # i18n files
└── config/            # Configuration files
```

### Running Tests

```bash
make test
```

### Building for Production

```bash
make docker-build
```

## Monitoring

The service exposes health check endpoint:

```bash
curl http://localhost:8080/health
```

## License
