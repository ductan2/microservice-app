# English Learning App Backend

## Overview

This backend is designed for an English learning application using a **microservices architecture**. Each service is independently deployed and communicates via API/gRPC and a message broker. The system includes database, cache, message queue, monitoring, and logging components for scalability and maintainability.

---

## Core Services

### 1. User Service

- **Features:**
  - User registration, login, JWT authentication
  - Profile management (name, email, avatar)
  - Session management, refresh tokens, MFA
- **Purpose:** Acts as the identity provider for the system and centralizes authentication.

### 2. Content Service

- **Features:**
  - CRUD for learning content: lessons, flashcards, quizzes
  - Audio/image management (uploads to S3/MinIO)
  - Categorization by topic and level
- **Purpose:** Manages and distributes learning content.

### 3. Lesson/Progress Service

- **Features:**
  - Tracks user learning progress (streaks, scores)
  - Stores quiz results, updates leaderboards
  - Emits events when a user completes a lesson
- **Purpose:** Maintains user learning state and supports gamification.


### 4. Notification Service

- **Features:**
  - Sends email with html template
  - Tracks email 
  - Integrations: SMTP or SendGrid/Mailgun

- **Purpose:** Keeps users engaged & informed.

### 5. Aggregator service
- **Purpose:** The Aggregator Service (also called API Composition Layer / Backend-for-Frontend) is responsible for combining data from multiple domain services (User, Lesson, Progress, Content) into a single API response. Instead of the client making multiple calls, the aggregator merges responses and optimizes communication.


---

## Infrastructure Stack

- **PostgreSQL:**  
  Relational database for User and Progress Services. Stores user info, quiz results, and progress tracking.

- **MongoDB:**  
  NoSQL database suitable for storing lesson content as JSON, documents, and dynamic metadata.

- **Redis:**  
  In-memory cache for sessions, token storage, rate limiting, and leaderboards (using Sorted Sets).

- **RabbitMQ:**  
  Message broker for inter-service communication. Used to send events like `UserCreated` and `LessonCompleted` to Progress and Notification Services.

---

## API Gateway & Service Discovery

**API Gateway** (Traefik):

- Acts as a single entry point for all client requests
- Handles routing and load balancing across services
- Performs JWT verification and rate limiting
- Terminates TLS/SSL connections

**Service Discovery:**

- Uses Kubernetes Services or Consul for dynamic service registration and discovery
- Automatically registers and deregisters services to support scaling and resilience

---

## Monitoring & Observability

- **Prometheus:**  
  Collects metrics from services and exporters (CPU, RAM, DB connections, queue length).

- **Grafana:**  
  Visualizes metrics and logs. Dashboards for User, Content, and Lesson Services.

- **Loki + Promtail:**  
  Centralized log storage and querying. Promtail collects logs from containers and sends them to Loki. Grafana queries logs alongside metrics.

- **Exporters:**  
  - Postgres Exporter: Database metrics  
  - Redis Exporter: Redis metrics  
  - RabbitMQ Exporter: Queue metrics

---

## DevOps & Deployment

- **Docker Compose:**  
  Quickly launches the entire stack.

- **Kubernetes (optional):**  
  Scales services in production.

- **CI/CD Pipeline:**  
  Automates build, test, and deployment for each microservice.

---

## Position in Architecture
```
  Client (Web/Mobile)
          ↓
  API Gateway (Traefik: auth, routing, TLS, rate limiting)
          ↓
  Aggregator Service (Golang REST/GraphQL)
          ↓
  Core Microservices (User, Lesson, Progress, Content, Notification)
```
- Gateway → Security & infrastructure (auth, SSL, routing).
- Aggregator → Business composition (merging User + Lesson + Progress data).
- Domain services → Independent, focused on their own data and logic.
## Getting Started

To run the stack: