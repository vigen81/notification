GO_VERSION := 1.24.4

.PHONY: help build run test clean docker-build docker-run generate migrate swagger deps

help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $1, $2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Install dependencies
	go mod download
	go mod tidy

generate: ## Generate Ent code
	go generate ./ent

swagger: ## Generate Swagger documentation
	swag init -g cmd/server/main.go -o docs/

build: deps generate ## Build the application
	go build -o bin/notification-engine ./cmd/server

migrate: ## Run database migrations
	go run ./cmd/migrate

run: generate ## Run the application
	go run ./cmd/server

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

##@ Docker

docker-build: ## Build Docker image
	docker build -t notification-engine:latest .

docker-run: ## Run with Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f notification-engine

##@ Database

db-reset: ## Reset database (WARNING: This will delete all data)
	docker-compose down -v
	docker-compose up -d mysql
	sleep 10
	make migrate

##@ Cleanup

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf docs/swagger.*
	rm -f coverage.out coverage.html

clean-docker: ## Clean Docker containers and volumes
	docker-compose down -v
	docker system prune -f

##@ Quality

lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

##@ Development Helpers

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	make deps
	make generate
	make swagger
	@echo "Development environment ready!"

dev-reset: ## Reset development environment
	make clean
	make docker-down
	make dev-setup
	make docker-run

##@ Monitoring

logs: ## View application logs
	docker-compose logs -f notification-engine

kafka-topics: ## List Kafka topics
	docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-console: ## Open Kafka console consumer
	docker-compose exec kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic notifications --from-beginning

##@ API Testing

api-health: ## Test health endpoint
	curl -X GET http://localhost:8080/health

api-send: ## Test send notification endpoint (requires global token)
	curl -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"EMAIL","recipients":["test@example.com"],"body":"Test message","message_type":"system"}'

api-send-sms: ## Test SMS notification for tenant 1001
	curl -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"SMS","recipients":["+1234567890"],"body":"Test SMS message","message_type":"system"}'

api-batch: ## Test batch notification endpoint
	curl -X POST http://localhost:8080/api/v1/notifications/batch \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"EMAIL","recipients":["user1@example.com","user2@example.com"],"body":"Batch test message","message_type":"promo"}'

api-config-get: ## Get configuration for tenant 1001
	curl -X GET http://localhost:8080/api/v1/config/1001 \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN"

api-config-get-1002: ## Get configuration for tenant 1002
	curl -X GET http://localhost:8080/api/v1/config/1002 \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN"

api-config-update: ## Test config update for tenant 1001
	curl -X PUT http://localhost:8080/api/v1/config/1001 \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"email_providers":[{"name":"test","type":"smtp","priority":1,"enabled":true,"config":{"Host":"smtp.example.com"}}],"enabled":true}'

api-add-email-provider: ## Add email provider to tenant 1001
	curl -X POST http://localhost:8080/api/v1/config/1001/providers/email \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"secondary","type":"smtp","priority":2,"enabled":true,"config":{"Host":"smtp2.example.com","Port":"587","Username":"user","Password":"pass"}}'

api-add-sms-provider: ## Add SMS provider to tenant 1001
	curl -X POST http://localhost:8080/api/v1/config/1001/providers/sms \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"name":"twilio_backup","type":"twilio","priority":2,"enabled":true,"config":{"account_sid":"AC123","auth_token":"token123","from_number":"+1234567890"}}'

api-remove-provider: ## Remove email provider from tenant 1001
	curl -X DELETE http://localhost:8080/api/v1/config/1001/providers/email/secondary \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN"

api-kafka-publish: ## Test direct Kafka publishing
	curl -X POST http://localhost:8080/api/v1/kafka/publish \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "X-Kafka-API-Key: YOUR_KAFKA_KEY" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"EMAIL","recipients":["kafka@example.com"],"body":"Direct Kafka test","message_type":"system"}'

api-status: ## Check notification status (replace with actual request_id)
	curl -X GET http://localhost:8080/api/v1/notifications/status/550e8400-e29b-41d4-a716-446655440000 \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN"

api-docs: ## Open API documentation
	open http://localhost:8080/swagger/

##@ Multi-tenant Testing

api-test-all-tenants: ## Test notifications for all seeded tenants
	@echo "Testing tenant 1001 (Goodwin Casino)..."
	curl -s -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"EMAIL","recipients":["goodwin@example.com"],"body":"Test for Goodwin","message_type":"bonus"}' | jq .
	@echo "\nTesting tenant 1002 (StarBet)..."
	curl -s -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer YOUR_GLOBAL_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1002,"type":"EMAIL","recipients":["starbet@example.com"],"body":"Test for StarBet","message_type":"promo"}' | jq .

api-configs-all: ## Get configurations for all tenants
	@echo "=== Tenant 1001 (Goodwin Casino) ==="
	curl -s -X GET http://localhost:8080/api/v1/config/1001 -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" | jq .
	@echo "\n=== Tenant 1002 (StarBet) ==="
	curl -s -X GET http://localhost:8080/api/v1/config/1002 -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" | jq .
	@echo "\n=== Tenant 1003 (LuckyPlay) ==="
	curl -s -X GET http://localhost:8080/api/v1/config/1003 -H "Authorization: Bearer YOUR_GLOBAL_TOKEN" | jq .

##@ Production

prod-build: ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/notification-engine ./cmd/server

deploy: prod-build docker-build ## Deploy to production (customize as needed)
	@echo "Deploying to production..."
	# Add your deployment commands here

seed: ## Seed database with test data
	go run ./cmd/seed

seed-fresh: ## Reset and seed database with fresh test data
	make migrate
	make seed