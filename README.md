# Notification Engine

A high-performance, multi-tenant notification engine built with Go, Ent, Fiber, Kafka, Watermill, and Uber FX. Supports Email, SMS, and Push notifications with per-partner configurations and batch processing capabilities.

## 🚀 Features

- **Multi-tenant Architecture**: Per-partner configurations with isolated data
- **Multiple Notification Types**: Email, SMS, and Push notifications
- **Provider Flexibility**: Support for multiple providers per channel
  - **Email**: SendGrid, SendX, SMTP
  - **SMS**: Twilio, Nexmo
  - **Push**: FCM (Firebase Cloud Messaging)
- **Dual API Support**: HTTP REST API and Kafka messaging
- **Batch Processing**: Efficient batch sending with configurable thresholds
- **Scheduled Notifications**: Support for future-dated notifications
- **Message Type Based Routing**: Different from addresses based on message type (bonus, promo, system, etc.)
- **Global Authentication**: Manage any tenant from a single authenticated session
- **Comprehensive Logging**: Structured logging with Graylog integration
- **Swagger Documentation**: Auto-generated API documentation
- **Docker Ready**: Complete containerization with Docker Compose

## 📋 Prerequisites

- Go 1.21+
- Docker & Docker Compose
- MySQL 8.0+
- Apache Kafka
- (Optional) Graylog for centralized logging

## 🛠 Installation

### Quick Start with Docker

1. **Clone the repository**
```bash
git clone <repository-url>
cd notification-engine
```

2. **Start the entire stack**
```bash
make docker-run
```

3. **Check health**
```bash
make api-health
```

### Development Setup

1. **Install dependencies**
```bash
make deps
```

2. **Generate code and documentation**
```bash
make generate
make swagger
```

3. **Start infrastructure services**
```bash
docker-compose up -d mysql kafka zookeeper
```

4. **Run database migrations**
```bash
make migrate
```

5. **Start the application**
```bash
make run
```

## 📖 API Documentation

- **Swagger UI**: http://localhost:8080/swagger/
- **Health Check**: http://localhost:8080/health
- **Kafka UI**: http://localhost:8081 (monitoring)

## 🔧 Configuration

### Partner Configuration Example

Each tenant can configure multiple providers with specific settings:

```json
{
  "email_providers": [
    {
      "name": "primary",
      "type": "smtp",
      "priority": 1,
      "enabled": true,
      "config": {
        "Host": "smtp.sendgrid.net",
        "Port": "465",
        "Username": "apikey",
        "Password": "your_api_key",
        "SMTPAuth": "1",
        "SMTPSecure": "ssl",
        "MSGBonusFrom": "bonus@goodwin.am",
        "MSGPromoFrom": "noreply@goodwin.am",
        "MSGSystemFrom": "noreply@goodwin.am",
        "MSGBonusFromName": "Goodwin Bonus",
        "MSGPromoFromName": "Goodwin Promo",
        "MSGSystemFromName": "Goodwin System"
      }
    }
  ],
  "batch_config": {
    "enabled": true,
    "max_batch_size": 100,
    "flush_interval_seconds": 10
  }
}
```

## 📡 API Usage

### 🔑 Authentication

All API endpoints require a global JWT token that allows access to any tenant:

```bash
Authorization: Bearer YOUR_GLOBAL_TOKEN
```

The token should contain claims like:
```json
{
  "user_id": "admin123",
  "username": "admin@company.com", 
  "role": "admin"
}
```

### Send Single Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1001,
    "type": "EMAIL",
    "recipients": ["user@example.com"],
    "body": "Hello! This is your notification.",
    "headline": "Important Update",
    "message_type": "bonus"
  }'
```

### Send SMS Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1002,
    "type": "SMS",
    "recipients": ["+1234567890"],
    "body": "Your verification code is: 123456",
    "message_type": "system"
  }'
```

### Send Batch Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications/batch \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1001,
    "type": "EMAIL",
    "recipients": ["user1@example.com", "user2@example.com"],
    "body": "Batch notification content",
    "headline": "Weekly Newsletter",
    "message_type": "promo"
  }'
```

### Send Scheduled Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1001,
    "type": "EMAIL",
    "recipients": ["user@example.com"],
    "body": "This message is scheduled for later",
    "schedule_ts": 1640995200,
    "message_type": "system"
  }'
```

### Check Notification Status

```bash
curl -X GET http://localhost:8080/api/v1/notifications/status/{request_id} \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN"
```

**Response includes tenant information:**
```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "COMPLETED",
  "type": "EMAIL",
  "tenant_id": 1001,
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:01:00Z"
}
```

### Get Tenant Configuration

```bash
# Get config for Goodwin Casino (tenant 1001)
curl -X GET http://localhost:8080/api/v1/config/1001 \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN"

# Get config for StarBet (tenant 1002)  
curl -X GET http://localhost:8080/api/v1/config/1002 \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN"
```

### Update Tenant Configuration

```bash
curl -X PUT http://localhost:8080/api/v1/config/1001 \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email_providers": [
      {
        "name": "primary",
        "type": "smtp",
        "priority": 1,
        "enabled": true,
        "config": {
          "Host": "smtp.sendgrid.net",
          "Port": "465",
          "Username": "apikey",
          "Password": "your_api_key",
          "SMTPAuth": "1",
          "SMTPSecure": "ssl",
          "MSGBonusFrom": "bonus@goodwin.am",
          "MSGPromoFrom": "noreply@goodwin.am",
          "MSGSystemFrom": "noreply@goodwin.am"
        }
      }
    ],
    "batch_config": {
      "enabled": true,
      "max_batch_size": 100,
      "flush_interval_seconds": 10
    },
    "enabled": true
  }'
```

### Add Email Provider

```bash
curl -X POST http://localhost:8080/api/v1/config/1001/providers/email \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "backup_smtp",
    "type": "smtp",
    "priority": 2,
    "enabled": true,
    "config": {
      "Host": "smtp.backup.com",
      "Port": "587",
      "Username": "backup_user",
      "Password": "backup_pass",
      "SMTPAuth": "1",
      "SMTPSecure": "tls"
    }
  }'
```

### Add SMS Provider

```bash
curl -X POST http://localhost:8080/api/v1/config/1002/providers/sms \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "twilio_primary",
    "type": "twilio",
    "priority": 1,
    "enabled": true,
    "config": {
      "account_sid": "AC_your_account_sid",
      "auth_token": "your_auth_token",
      "from_number": "+1234567890"
    }
  }'
```

### Remove Provider

```bash
# Remove email provider "backup_smtp" from tenant 1001
curl -X DELETE http://localhost:8080/api/v1/config/1001/providers/email/backup_smtp \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN"
```

### Direct Kafka Publishing

For high-throughput scenarios, publish directly to Kafka:

```bash
curl -X POST http://localhost:8080/api/v1/kafka/publish \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -H "X-Kafka-API-Key: YOUR_KAFKA_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1001,
    "type": "EMAIL",
    "recipients": ["user@example.com"],
    "body": "Direct Kafka notification",
    "headline": "Kafka Test",
    "message_type": "system"
  }'
```

### Multi-Tenant Management Examples

```bash
# Send different notifications for different tenants
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -d '{"tenant_id":1001,"type":"EMAIL","recipients":["goodwin@example.com"],"body":"Welcome to Goodwin Casino!","message_type":"bonus"}'

curl -X POST http://localhost:8080/api/v1/notifications/send \
  -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
  -d '{"tenant_id":1002,"type":"SMS","recipients":["+1234567890"],"body":"StarBet: Your account is verified","message_type":"system"}'
```

### Message Type Based Routing

The system automatically selects the appropriate "from" address based on message type:

```bash
# Bonus message - uses MSGBonusFrom
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -d '{"tenant_id":1001,"type":"EMAIL","recipients":["user@example.com"],"body":"You got a bonus!","message_type":"bonus"}'

# Promo message - uses MSGPromoFrom  
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -d '{"tenant_id":1001,"type":"EMAIL","recipients":["user@example.com"],"body":"Special offer!","message_type":"promo"}'

# System message - uses MSGSystemFrom
curl -X POST http://localhost:8080/api/v1/notifications/send \
  -d '{"tenant_id":1001,"type":"EMAIL","recipients":["user@example.com"],"body":"Account updated","message_type":"system"}'
```

## 🏗 Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   Kafka Client  │    │   Scheduler     │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Fiber Server   │    │  Kafka Producer │    │ Scheduler Worker│
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 ▼
                    ┌─────────────────┐
                    │  Kafka Topics   │
                    └─────────┬───────┘
                              ▼
                    ┌─────────────────┐
                    │ Notification    │
                    │    Worker       │
                    └─────────┬───────┘
                              ▼
          ┌─────────────────────────────────────┐
          │            Database                 │
          │         (MySQL + Ent)               │
          └─────────┬───────────────────────────┘
                    ▼
          ┌─────────────────┐    ┌─────────────────┐
          │   Email Provider│    │   SMS Provider  │
          │   (SMTP/SG/SX)  │    │  (Twilio/Nexmo) │
          └─────────────────┘    └─────────────────┘
```

## 🚀 Development

### Available Commands

```bash
make help              # Show all available commands
make dev-setup         # Setup development environment
make generate          # Generate Ent code
make swagger           # Generate API documentation
make test              # Run tests
make lint              # Run linter
make docker-run        # Start with Docker Compose
make logs              # View application logs
```

### Testing Commands

```bash
make api-health                # Test health endpoint
make api-send                  # Test notification sending
make api-config-get            # Test config retrieval
make api-test-all-tenants      # Test multiple tenants
make api-configs-all           # Get all tenant configs
```

### Project Structure

```
notification-engine/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── app/               # Application lifecycle
│   ├── config/            # Configuration management
│   ├── handlers/          # HTTP handlers
│   ├── kafka/             # Kafka pub/sub
│   ├── logger/            # Structured logging
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Domain models
│   ├── providers/         # Notification providers
│   ├── repository/        # Data access layer
│   ├── server/            # HTTP server
│   ├── services/          # Business logic
│   └── workers/           # Background workers
├── ent/                   # Ent schema definitions
├── docs/                  # API documentation
├── config/                # Configuration files
└── types/                 # Custom types
```

## 🔍 Monitoring & Logging

- **Health Checks**: `/health`, `/ready`, `/live` endpoints
- **Metrics**: Application metrics via structured logging
- **Tracing**: Request tracing with correlation IDs
- **Logging**: Centralized logging with Graylog
- **Kafka Monitoring**: Kafka UI for topic and consumer monitoring

## 🔒 Security

- **Global JWT Authentication**: Bearer token authentication for HTTP API
- **API Key Authentication**: Additional security for Kafka endpoints
- **Tenant Isolation**: Multi-tenant data isolation
- **Rate Limiting**: Configurable rate limits per tenant
- **Input Validation**: Comprehensive request validation

## 🚀 Deployment

### Production Build

```bash
make prod-build
make docker-build
```

### Environment Variables

```bash
CONFIG_PATH=/app/config/config.yaml
ENVIRONMENT=production
DB_HOST=mysql-host
DB_PASSWORD=secure_password
KAFKA_BROKERS=kafka1:9092,kafka2:9092
GRAYLOG_ADDR=graylog:12201
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new features
- Update documentation for API changes
- Use conventional commit messages
- Ensure all tests pass before submitting PR

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Troubleshooting

### Common Issues

**Database Connection Issues**
```bash
# Check MySQL container status
docker-compose ps mysql

# View MySQL logs
docker-compose logs mysql

# Reset database
make db-reset
```

**Kafka Connection Issues**
```bash
# Check Kafka status
docker-compose ps kafka

# View Kafka topics
make kafka-topics

# Monitor Kafka messages
make kafka-console
```

**Application Not Starting**
```bash
# Check application logs
make logs

# Verify configuration
cat config/config.yaml

# Check health endpoints
curl http://localhost:8080/health
```

### Performance Tuning

**Database Optimization**
- Adjust connection pool settings in config
- Monitor slow queries
- Add appropriate indexes

**Kafka Optimization**
- Tune batch settings for throughput
- Adjust consumer group settings
- Monitor lag and throughput

**Application Tuning**
- Adjust worker concurrency
- Tune batch processing settings
- Monitor memory usage

## 📊 Metrics & Monitoring

### Key Metrics to Monitor

- **Notification Throughput**: Messages processed per second
- **Error Rates**: Failed notification percentage by provider
- **Latency**: End-to-end notification delivery time
- **Queue Depth**: Kafka topic lag and message backlog
- **Database Performance**: Connection pool usage and query performance

### Alerting

Set up alerts for:
- High error rates (>5%)
- Queue depth exceeding thresholds
- Database connection issues
- Provider API failures
- Memory/CPU usage spikes

## 🔧 Advanced Configuration

### Custom Provider Implementation

To add a new notification provider:

1. Implement the provider interface:
```go
type EmailProvider interface {
    Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error
    SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error
    ValidateConfig() error
    GetType() string
}
```

2. Register the provider:
```go
registry.RegisterEmailProvider("your-provider", NewYourProvider)
```

3. Add configuration support in models

### Environment-Specific Configurations

**Development**
```yaml
server:
  port: ":8080"
database:
  host: localhost
kafka:
  brokers: ["localhost:9092"]
```

**Production**
```yaml
server:
  port: ":8080"
  read_timeout: 30s
  write_timeout: 30s
database:
  host: production-db-host
  max_open_conns: 50
kafka:
  brokers: ["kafka1:9092", "kafka2:9092", "kafka3:9092"]
```

## 🧪 Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
make test-integration
```

### Load Testing
```bash
# Start the application
make docker-run

# Run load tests (example with hey)
hey -n 10000 -c 100 -m POST \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"tenant_id":1001,"type":"EMAIL","recipients":["test@example.com"],"body":"Load test"}' \
  http://localhost:8080/api/v1/notifications/send
```

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Ent Documentation](https://entgo.io/docs/getting-started)
- [Fiber Documentation](https://docs.gofiber.io/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
- [Uber FX Documentation](https://pkg.go.dev/go.uber.org/fx)

## 🙋‍♀️ Support

- **Documentation**: Check the `/docs` folder for detailed guides
- **Issues**: Use GitHub Issues for bug reports and feature requests
- **Discussions**: Use GitHub Discussions for questions and community support
- **API Documentation**: Available at `/swagger/` when running
- **Logs**: Application logs are sent to the configured Graylog endpoint

## 🎯 Key Benefits of Global Auth

1. **Multi-tenant Management**: Manage any tenant from one authenticated session
2. **Admin Panel Friendly**: Perfect for admin interfaces where you select tenant
3. **API Simplicity**: Clear separation between auth and tenant targeting
4. **Operational Tools**: Easy to build monitoring and management tools
5. **Flexibility**: Switch between tenants without re-authentication

---

**Built with ❤️ using Go, Ent, Fiber, Kafka, and Uber FX**