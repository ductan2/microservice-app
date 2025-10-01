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

## Getting Started

To run the stack: