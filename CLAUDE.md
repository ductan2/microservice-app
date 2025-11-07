# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a sophisticated English learning platform built with a **polyglot microservice architecture**. Each service is independently deployable and optimized for its specific domain, with clear separation of concerns and modern infrastructure patterns.

### Core Services

#### 1. BFF (Backend for Frontend) Service
- **Location:** `bff-services/`
- **Language:** Go 1.24.6 with Gin framework
- **Purpose:** API gateway that aggregates responses from downstream services
- **Key Features:** Authentication, rate limiting, CORS, route composition
- **Port:** 8010

#### 2. User Management Service
- **Location:** `user-services/`
- **Language:** Go 1.24.6 with Gin framework
- **Database:** PostgreSQL with GORM
- **Purpose:** Centralized authentication and user management
- **Key Features:** JWT auth, MFA, sessions, account management
- **Port:** 8001

#### 3. Lesson & Progress Service
- **Location:** `lesson-services/`
- **Language:** Python 3.11 with FastAPI
- **Database:** PostgreSQL with SQLAlchemy
- **Purpose:** Learning progress, quizzes, spaced repetition, gamification
- **Key Features:** Lesson tracking, quiz attempts, streaks, leaderboards
- **Port:** 8005

#### 4. Content Management Service
- **Location:** `content-services/`
- **Language:** Go 1.24.0 with Gin + GraphQL (gqlgen)
- **Database:** MongoDB for content storage
- **Storage:** AWS S3/MinIO for media assets
- **Purpose:** Content creation, management, and distribution
- **Key Features:** Lessons, quizzes, flashcards, media management
- **Port:** 8004

#### 5. Notification Service
- **Location:** `notification-services/`
- **Language:** TypeScript/Node.js with Express
- **Database:** PostgreSQL
- **Purpose:** Email and notification delivery
- **Key Features:** Multi-provider email, async processing, templates
- **Port:** Configurable

### Infrastructure Stack

**Databases:**
- **PostgreSQL:** User data, progress tracking, audit logs
- **MongoDB:** Content management, flexible schema
- **Redis:** Session caching, rate limiting, leaderboards

**Messaging:**
- **RabbitMQ:** Event-driven communication between services
- **Outbox Pattern:** Reliable event publishing with transactional guarantees

**API Gateway:**
- **Traefik:** Load balancing, TLS termination, routing
- **Path-based routing:** `/api/user`, `/api/lesson`, `/api/content`, `/api/bff`

**Monitoring & Observability:**
- **Prometheus:** Metrics collection
- **Grafana:** Dashboards and visualization
- **Loki + Promtail:** Centralized logging
- **Exporters:** PostgreSQL, Redis, RabbitMQ metrics

## Development Commands

### Local Development Setup

```bash
# Start the entire infrastructure stack
cd infrastructure
docker-compose up -d

# View service logs
docker-compose logs -f [service-name]

# Stop the stack
docker-compose down
```

### Service-Specific Commands

**Go Services (BFF, User, Content):**
```bash
cd [service-directory]

# Build the service
make build

# Run locally (with hot reload if available)
make run

# Run tests
make test

# Format code
make fmt

# Tidy dependencies
make tidy

# Generate GraphQL code (content-services only)
make generate
```

**Python Service (Lessons):**
```bash
cd lesson-services

# Install dependencies
pip install -r requirements.txt

# Run the application
python main.py

# Run tests
pytest

# Run tests with coverage
pytest --cov=.
```

**Node.js Service (Notifications):**
```bash
cd notification-services

# Install dependencies
npm install

# Run the application
npm start

# Run in development mode
npm run dev
```

### Database Operations

**PostgreSQL:**
- Access via pgadmin at `http://localhost:5050` (admin/password)
- Database: `english_app`
- User: `user` (configurable via environment variables)

**MongoDB:**
- Connection: `mongodb://localhost:27017`
- Database: `content`

**Redis:**
- Connection: `localhost:6379`
- Password: `redis_password` (configurable)

### Message Queue

**RabbitMQ:**
- Management UI: `http://localhost:15672`
- User: `user`, Password: `password` (configurable)

## Configuration

### Environment Variables

All services use environment variables for configuration. Create `.env` files in each service directory:

```bash
# Common variables
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=english_app
REDIS_HOST=redis
REDIS_PASSWORD=redis_password
RABBITMQ_USER=user
RABBITMQ_PASSWORD=password

# Service-specific variables
PORT=8001  # for user-services
PORT=8005  # for lesson-services
PORT=8004  # for content-services
PORT=8010  # for bff-services
```

### Configuration Files

- `infrastructure/docker-compose.yml`: Main orchestration
- `infrastructure/docker-compose.override.yml`: Local development overrides
- `infrastructure/prometheus.yml`: Metrics collection configuration
- Service-specific `.env` files and `Dockerfile`s

## API Architecture

### Service Communication

1. **Client → API Gateway**: Single entry point via Traefik
2. **API Gateway → BFF**: Route composition and aggregation
3. **BFF → Services**: HTTP calls for data retrieval
4. **Services → RabbitMQ**: Event publishing for async operations

### Request Flow Example

```
Client Request → Traefik (/api/bff) → BFF Service → User Service (HTTP) + Lesson Service (HTTP) → Response Aggregation
```

### Key API Patterns

- **GraphQL**: Content service for flexible data fetching
- **REST**: Most services follow REST conventions
- **Event-Driven**: Services publish events for state changes

## Testing

### Unit Testing

**Go Services:**
```bash
# Run all tests
make test

# Run specific test file
go test ./internal/services/user_service_test.go -v
```

**Python Service:**
```bash
# Run all tests
pytest

# Run specific test
pytest tests/test_lesson_service.py

# Run with coverage
pytest --cov=app
```

### Integration Testing

Use the Docker Compose stack for integration testing. Services are configured with proper health checks and dependencies.

## Development Patterns

### Service-to-Service Communication

- **HTTP Synchronous**: BFF calls downstream services directly
- **Asynchronous Events**: RabbitMQ for cross-service notifications
- **Circuit Breaking**: Implement timeout and retry logic
- **Service Discovery**: Use Docker service names for inter-service communication

### Data Access Patterns

- **Repository Pattern**: Each service implements repository abstraction
- **Unit of Work**: Database transactions where appropriate
- **Outbox Pattern**: Reliable event publishing with database transaction
- **CQRS**: Separate read and write models (emerging pattern)

### Security Patterns

- **JWT Authentication**: Token-based auth with refresh tokens
- **MFA Support:** Multi-factor authentication for sensitive operations
- **Service Auth:** Consider mutual TLS for service-to-service communication
- **Rate Limiting:** Redis-based rate limiting at API gateway and service levels

## Monitoring and Debugging

### Health Checks

All services implement health check endpoints:
- `/health` - Basic health status
- `/ready` - Service readiness for traffic

### Logging

- **Structured Logging**: JSON format across all services
- **Log Levels**: Consistent use of info, warn, error levels
- **Correlation IDs:** Include for request tracing

### Metrics

Key metrics to monitor:
- Request latency and error rates
- Database connection pool usage
- Redis memory usage and connection counts
- RabbitMQ queue lengths and message rates

## Deployment

### Docker Build Patterns

- **Multi-stage builds:** Production-optimized images
- **Health checks:** Proper container health monitoring
- **Graceful shutdown:** Handle SIGTERM signals properly

### Environment-Specific Configurations

- **Development:** Hot reload, debug logging, local databases
- **Staging:** Production-like environment, monitoring enabled
- **Production:** TLS, HTTPS, monitoring, logging aggregation

## Common Development Workflows

### Adding a New Service

1. Create new service directory with proper structure
2. Add Docker Compose configuration
3. Configure Traefik routing
4. Add service to BFF client dependencies
5. Implement health check endpoint
6. Add Prometheus metrics if needed

### Database Schema Changes

1. Update migration files (Alembic for Python, GORM auto-migration for Go)
2. Update data models
3. Run migrations on startup
4. Consider backward compatibility for production

### Adding New API Endpoints

1. Add route configuration
2. Implement controller function
3. Add service layer logic
4. Update DTOs/Request models
5. Add tests
6. Update API documentation