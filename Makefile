.PHONY: help build run test clean docker-up docker-down docker-logs docker-restart migrate dev

help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the API server locally"
	@echo "  make run-sim        - Run the device simulator"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-up      - Start all Docker containers (detached)"
	@echo "  make docker-run     - Start all Docker containers (attached)"
	@echo "  make docker-down    - Stop all Docker containers"
	@echo "  make docker-logs    - View container logs"
	@echo "  make docker-restart - Restart all containers"
	@echo "  make docker-clean   - Stop containers and remove volumes"
	@echo "  make migrate        - Run database migrations"
	@echo "  make dev            - Run locally with dependencies in Docker"

build:
	go build -o bin/api cmd/api/main.go
	go build -o bin/simulator cmd/simulator/main.go

run:
	go run cmd/api/main.go

run-sim:
	go run cmd/simulator/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

# Docker commands - Keep it simple!
docker-up:
	docker compose -f docker-local.compose.yml up -d
	@echo "All services starting..."
	@echo "Run 'make docker-logs' to view logs"

run-rabbit:
	docker compose -f docker-local.compose.yml up -d rabbitmq

docker-run:
	docker compose -f docker-local.compose.yml up --build

docker-down:
	docker compose -f docker-local.compose.yml down

docker-logs:
	docker compose -f docker-local.compose.yml logs -f

docker-restart:
	docker compose -f docker-local.compose.yml restart

docker-stop: docker-down

docker-clean:
	docker compose -f docker-local.compose.yml down -v
	@echo "Stopped containers and removed volumes"

# Check health of services
docker-health:
	@echo "Checking service health..."
	@docker compose -f docker-local.compose.yml ps
	@echo "\nRabbitMQ Management UI: http://localhost:15672 (guest/guest)"

migrate:
	@echo "Running migrations..."
	@sleep 2  # Give postgres a moment if just started
	docker exec -i iot-inventory-management-db-local psql -U postgres -d iot_inventory_local < migrations/001_initial_schema.sql
	@echo "Migrations completed"

# Development mode - run dependencies in Docker, app locally
dev:
	@echo "Starting dependencies in Docker..."
	docker compose -f docker-local.compose.yml up -d postgres redis rabbitmq
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@echo "Running migrations..."
	@make migrate
	@echo "Starting application locally..."
	@echo "Make sure to set: RABBITMQ_URL=amqp://guest:guest@localhost:5672/"
	@echo "                   DB_HOST=localhost"
	@echo "                   REDIS_HOST=localhost"
	go run cmd/api/main.go

# Quick restart for development
dev-restart:
	docker compose -f docker-local.compose.yml restart rabbitmq
	@sleep 3
	@echo "RabbitMQ restarted"