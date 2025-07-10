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
	POD_ENV=local go run ./cmd/migrate

run: generate ## Run the application (local development)
	@if [ ! -f docs/docs.go ]; then echo "📝 Generating Swagger docs..."; swag init -g cmd/server/main.go -o docs/ || echo "⚠️ Swagger generation failed, continuing..."; fi
	POD_ENV=local go run ./cmd/server

run-dev: generate ## Run the application (dev environment - uses AWS Parameter Store)
	@if [ ! -f docs/docs.go ]; then echo "📝 Generating Swagger docs..."; swag init -g cmd/server/main.go -o docs/ || echo "⚠️ Swagger generation failed, continuing..."; fi
	POD_ENV=dev go run ./cmd/server

run-prod: generate ## Run the application (prod environment - uses AWS Parameter Store)
	@if [ ! -f docs/docs.go ]; then echo "📝 Generating Swagger docs..."; swag init -g cmd/server/main.go -o docs/ || echo "⚠️ Swagger generation failed, continuing..."; fi
	POD_ENV=prod go run ./cmd/server

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

##@ Docker

docker-build: ## Build Docker image
	docker build -t notification-engine:latest .

docker-build-dev: ## Build Docker image for dev environment
	docker build -t notification-engine:dev .

docker-build-prod: ## Build Docker image for production
	docker build -t notification-engine:prod .

docker-run: ## Run with Docker Compose (local environment)
	@echo "🚀 Starting Docker Compose..."
	@if docker network ls | grep -q notification_default; then \
		echo "🔄 Recreating Docker network..."; \
		docker-compose down -v; \
		docker network rm notification_default || true; \
	fi
	POD_ENV=local docker-compose up -d

docker-run-dev: ## Run with Docker Compose (dev environment)
	@echo "🚀 Starting Docker Compose (dev)..."
	@if docker network ls | grep -q notification_default; then \
		echo "🔄 Recreating Docker network..."; \
		docker-compose down -v; \
		docker network rm notification_default || true; \
	fi
	POD_ENV=dev docker-compose up -d

docker-run-prod: ## Run with Docker Compose (prod environment)
	@echo "🚀 Starting Docker Compose (prod)..."
	@if docker network ls | grep -q notification_default; then \
		echo "🔄 Recreating Docker network..."; \
		docker-compose down -v; \
		docker network rm notification_default || true; \
	fi
	POD_ENV=prod docker-compose up -d

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

##@ AWS Parameter Store

aws-create-dev-config: ## Create development configuration in AWS Parameter Store
	@echo "Creating development configuration in Parameter Store..."
	aws ssm put-parameter \
		--name "/dev/notification-engine" \
		--value "$$(cat scripts/dev-config.json)" \
		--type "SecureString" \
		--description "Notification Engine Development Configuration" \
		--overwrite

aws-create-prod-config: ## Create production configuration in AWS Parameter Store
	@echo "Creating production configuration in Parameter Store..."
	aws ssm put-parameter \
		--name "/prod/notification-engine" \
		--value "$$(cat scripts/prod-config.json)" \
		--type "SecureString" \
		--description "Notification Engine Production Configuration" \
		--overwrite

aws-get-dev-config: ## Get development configuration from AWS Parameter Store
	aws ssm get-parameter \
		--name "/dev/notification-engine" \
		--with-decryption \
		--query "Parameter.Value" \
		--output text | jq .

aws-get-prod-config: ## Get production configuration from AWS Parameter Store
	aws ssm get-parameter \
		--name "/prod/notification-engine" \
		--with-decryption \
		--query "Parameter.Value" \
		--output text | jq .

aws-update-dev-config: ## Update development configuration in AWS Parameter Store
	aws ssm put-parameter \
		--name "/dev/notification-engine" \
		--value "$$(cat scripts/dev-config.json)" \
		--type "SecureString" \
		--overwrite

aws-update-prod-config: ## Update production configuration in AWS Parameter Store
	aws ssm put-parameter \
		--name "/prod/notification-engine" \
		--value "$$(cat scripts/prod-config.json)" \
		--type "SecureString" \
		--overwrite

##@ Cleanup

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf docs/swagger.*
	rm -f coverage.out coverage.html

clean-docker: ## Clean Docker containers and volumes
	@echo "🧹 Cleaning Docker resources..."
	docker-compose down -v || true
	docker network rm notification_default || true
	docker system prune -f

docker-reset: clean-docker ## Reset Docker environment completely
	@echo "🔄 Resetting Docker environment..."
	docker network prune -f
	docker volume prune -f

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

api-send: ## Test send notification endpoint (requires auth token)
	curl -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer test-token" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"EMAIL","recipients":["test@example.com"],"body":"Test message","message_type":"system"}'

api-send-sms: ## Test SMS notification for tenant 1001
	curl -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer test-token" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"SMS","recipients":["+1234567890"],"body":"Test SMS message","message_type":"system"}'

api-batch: ## Test batch notification endpoint
	curl -X POST http://localhost:8080/api/v1/notifications/batch \
		-H "Authorization: Bearer test-token" \
		-H "Content-Type: application/json" \
		-d '{"tenant_id":1001,"type":"EMAIL","recipients":["user1@example.com","user2@example.com"],"body":"Batch test message","message_type":"promo"}'

api-config-get: ## Get configuration for tenant 1001
	curl -X GET http://localhost:8080/api/v1/config/1001 \
		-H "Authorization: Bearer test-token"

api-config-update: ## Test config update for tenant 1001
	curl -X PUT http://localhost:8080/api/v1/config/1001 \
		-H "Authorization: Bearer test-token" \
		-H "Content-Type: application/json" \
		-d '{"email_providers":[{"name":"test","type":"smtp","priority":1,"enabled":true,"config":{"Host":"smtp.example.com"}}],"enabled":true}'

api-docs: ## Open API documentation
	open http://localhost:8080/swagger/

##@ Environment Testing

test-local: ## Test with local configuration
	POD_ENV=local make api-health

test-dev: ## Test with dev configuration (requires AWS access)
	POD_ENV=dev make api-health

test-prod: ## Test with prod configuration (requires AWS access)
	POD_ENV=prod make api-health

##@ Production

prod-build: ## Build for production
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/notification-engine ./cmd/server

deploy-dev: prod-build docker-build-dev ## Deploy to development
	@echo "Deploying to development environment..."
	# Add your deployment commands here

deploy-prod: prod-build docker-build-prod ## Deploy to production
	@echo "Deploying to production environment..."
	# Add your deployment commands here

seed: ## Seed database with test data
	POD_ENV=local go run ./cmd/seed

seed-fresh: ## Reset and seed database with fresh test data
	make migrate
	make seed

##@ Configuration Templates

create-config-templates: ## Create configuration template files
	mkdir -p scripts
	@echo "Creating configuration templates..."
	@echo "Check scripts/ directory for config templates"