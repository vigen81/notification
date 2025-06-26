GO_VERSION := 1.24.4.PHONY: help build run test clean docker-build docker-run generate migrate swagger deps

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

api-send: ## Test send notification endpoint (requires token)
	curl -X POST http://localhost:8080/api/v1/notifications/send \
		-H "Authorization: Bearer YOUR_TOKEN" \
		-H "Content-Type: application/json" \
		-d '{"type":"EMAIL","recipients":["test@example.com"],"body":"Test message"}'

api-docs: ## Open API documentation
	open http://localhost:8080/swagger/

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