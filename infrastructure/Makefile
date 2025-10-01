# English App Infrastructure Makefile

.PHONY: help start stop restart logs clean build status

# Default target
help:
	@echo "English App Infrastructure Commands:"
	@echo ""
	@echo "  make start     - Start all services"
	@echo "  make stop      - Stop all services"
	@echo "  make restart   - Restart all services"
	@echo "  make logs      - Show logs for all services"
	@echo "  make status    - Show status of all services"
	@echo "  make clean     - Stop and remove all containers and volumes"
	@echo "  make build     - Build all services"
	@echo ""
	@echo "Service-specific commands:"
	@echo "  make logs-user     - Show user-services logs"
	@echo "  make logs-lesson   - Show lesson-services logs"
	@echo "  make logs-content  - Show content-services logs"
	@echo "  make logs-golang   - Show golang-init logs"
	@echo ""

# Start all services
start:
	@echo "🚀 Starting all services..."
	docker compose --env-file docker-compose.env up -d --build

# Stop all services
stop:
	@echo "🛑 Stopping all services..."
	docker compose --env-file docker-compose.env down

# Restart all services
restart: stop start

# Show logs for all services
logs:
	docker compose --env-file docker-compose.env logs -f

# Show status of all services
status:
	docker compose --env-file docker-compose.env ps

# Clean everything
clean:
	@echo "🧹 Cleaning up..."
	docker compose --env-file docker-compose.env down -v
	docker system prune -f

# Build all services
build:
	docker compose --env-file docker-compose.env build

# Service-specific logs
logs-user:
	docker compose --env-file docker-compose.env logs -f user-services

logs-lesson:
	docker compose --env-file docker-compose.env logs -f lesson-services

logs-content:
	docker compose --env-file docker-compose.env logs -f content-services

logs-golang:
	docker compose --env-file docker-compose.env logs -f golang-init

# Database commands
db-shell:
	docker compose --env-file docker-compose.env exec postgres psql -U user -d english_app

redis-shell:
	docker compose --env-file docker-compose.env exec redis redis-cli

# Monitoring
monitor:
	@echo "📊 Monitoring URLs:"
	@echo "  Grafana: http://localhost:3000 (admin/admin)"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  RabbitMQ: http://localhost:15672 (user/password)"
