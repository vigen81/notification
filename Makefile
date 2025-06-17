.PHONY: help build run test clean docker-build docker-run generate migrate

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run with Docker Compose"
	@echo "  make generate     - Generate Ent code"
	@echo "  make migrate      - Run database migrations"

build:
	go build -o bin/notification-engine ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-build:
	docker build -t notification-engine:latest .

docker-run:
	docker-compose up -d

docker-down:
	docker-compose down

generate:
	go generate ./ent

migrate:
	go run ./cmd/migrate
