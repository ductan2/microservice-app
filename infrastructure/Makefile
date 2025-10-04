# English App Infrastructure Makefile

.PHONY: help start stop restart logs clean build status dev prod security

# Default target
help:
	@echo "English App Infrastructure Commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev        - Start development environment"
	@echo "  make start      - Start all services (development)"
	@echo "  make stop       - Stop all services"
	@echo "  make restart    - Restart all services"
	@echo "  make logs       - Show logs for all services"
	@echo "  make status     - Show status of all services"
	@echo "  make clean      - Stop and remove all containers and volumes"
	@echo "  make build      - Build all services"
	@echo ""
	@echo "Production:"
	@echo "  make prod       - Start production environment"
	@echo "  make prod-stop  - Stop production environment"
	@echo "  make prod-logs  - Show production logs"
	@echo ""
	@echo "Security & Networking:"
	@echo "  make security   - Show security status"
	@echo "  make networks   - Show network configuration"
	@echo "  make backup     - Backup all data"
	@echo "  make restore    - Restore from backup"
	@echo ""
	@echo "Service-specific commands:"
	@echo "  make logs-user     - Show user-services logs"
	@echo "  make logs-lesson   - Show lesson-services logs"
	@echo "  make logs-content  - Show content-services logs"
	@echo "  make logs-golang   - Show golang-init logs"
	@echo ""

# Development environment
dev:
	@echo "🚀 Starting development environment..."
	docker compose --env-file docker-compose.env up -d --build

# Start all services (development)
start: dev

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

# Production environment
prod:
	@echo "🚀 Starting production environment..."
	@if [ ! -f docker-compose.prod.env ]; then \
		echo "❌ Production environment file not found!"; \
		echo "Please copy docker-compose.prod.env.example to docker-compose.prod.env and configure it."; \
		exit 1; \
	fi
	docker compose --env-file docker-compose.prod.env -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# Stop production environment
prod-stop:
	@echo "🛑 Stopping production environment..."
	docker compose --env-file docker-compose.prod.env -f docker-compose.yml -f docker-compose.prod.yml down

# Show production logs
prod-logs:
	docker compose --env-file docker-compose.prod.env -f docker-compose.yml -f docker-compose.prod.yml logs -f

# Security status
security:
	@echo "🔒 Security Status:"
	@echo "=================="
	@echo "Exposed Ports:"
	@docker ps --format "table {{.Names}}\t{{.Ports}}" | grep -E "(5432|6379|5672|27017)" || echo "✅ No database ports exposed"
	@echo ""
	@echo "Network Configuration:"
	@docker network ls | grep english-app || echo "❌ Networks not found"
	@echo ""
	@echo "SSL Certificates:"
	@docker exec traefik ls -la /letsencrypt/ 2>/dev/null || echo "❌ SSL certificates not found"

# Network configuration
networks:
	@echo "🌐 Network Configuration:"
	@echo "========================="
	@docker network ls | grep english-app
	@echo ""
	@echo "Frontend Network:"
	@docker network inspect english-app-frontend --format "{{range .Containers}}{{.Name}} {{end}}" 2>/dev/null || echo "❌ Frontend network not found"
	@echo ""
	@echo "Backend Network:"
	@docker network inspect english-app-backend --format "{{range .Containers}}{{.Name}} {{end}}" 2>/dev/null || echo "❌ Backend network not found"

# Backup all data
backup:
	@echo "💾 Creating backup..."
	@mkdir -p backups/$(shell date +%Y%m%d_%H%M%S)
	@echo "Backing up PostgreSQL..."
	@docker exec postgres pg_dump -U user english_app > backups/$(shell date +%Y%m%d_%H%M%S)/postgres.sql
	@echo "Backing up MongoDB..."
	@docker exec mongodb mongodump --out /backup
	@docker cp mongodb:/backup backups/$(shell date +%Y%m%d_%H%M%S)/mongodb
	@echo "Backing up volumes..."
	@docker run --rm -v postgres_data:/data -v $(PWD)/backups/$(shell date +%Y%m%d_%H%M%S):/backup alpine tar czf /backup/postgres_data.tar.gz -C /data .
	@echo "✅ Backup completed in backups/$(shell date +%Y%m%d_%H%M%S)/"

# Restore from backup
restore:
	@echo "📥 Restore from backup..."
	@echo "Available backups:"
	@ls -la backups/ 2>/dev/null || echo "❌ No backups found"
	@echo "Please specify backup directory: make restore BACKUP_DIR=backups/20240101_120000"

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
