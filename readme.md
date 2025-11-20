# English Learning App Backend

## Overview

This backend is designed for an English learning application using a **microservices architecture**. Each service is independently deployed and communicates via API/gRPC and a message broker. The system includes database, cache, message queue, monitoring, and logging components for scalability and maintainability.

## Table of Contents

- [Architecture Diagram](#architecture-diagram)
- [Core Services](#core-services)
- [Infrastructure Stack](#infrastructure-stack)
- [API Gateway & Service Discovery](#api-gateway--service-discovery)
- [Monitoring & Observability](#monitoring--observability)
- [DevOps & Deployment](#devops--deployment)
- [Communication Patterns](#communication-patterns)
- [gRPC Implementation Guide](#grpc-implementation-guide) ⭐
- [Getting Started](#getting-started)
- [Additional Documentation](#additional-documentation)

---

## Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        Client[Web/Mobile Client]
    end

    subgraph "API Gateway"
        Traefik[Traefik API Gateway<br/>:80, :443<br/>TLS, Routing, Load Balancing]
    end

    subgraph "BFF Layer"
        BFF[BFF Service<br/>:8010<br/>Go/Gin<br/>API Composition & Aggregation]
    end

    subgraph "Microservices"
        UserService[User Service<br/>:8001<br/>Go/Gin<br/>Auth & User Management]
        LessonService[Lesson Service<br/>:8005<br/>Python/FastAPI<br/>Progress & Gamification]
        ContentService[Content Service<br/>:8004<br/>Go/GraphQL<br/>Content Management]
        NotificationService[Notification Service<br/>:8003<br/>Node.js/Express<br/>Email & Notifications]
        OrderService[Order Service<br/>:8006<br/>Go/Gin<br/>Payments & Orders]
    end

    subgraph "Data Layer"
        PostgreSQL[(PostgreSQL<br/>:5432<br/>User, Progress, Orders Data)]
        MongoDB[(MongoDB<br/>:27017<br/>Content & Media)]
        Redis[(Redis<br/>:6379<br/>Cache & Sessions)]
    end

    subgraph "Message Broker"
        RabbitMQ[RabbitMQ<br/>:5672, :15672<br/>Event Bus]
    end

    subgraph "Monitoring Stack"
        Prometheus[Prometheus<br/>:9090<br/>Metrics Collection]
        Grafana[Grafana<br/>:3000<br/>Dashboards & Alerts]
        Loki[Loki + Promtail<br/>:3100<br/>Log Aggregation]
        PGExporter[PostgreSQL Exporter<br/>:9187]
        RedisExporter[Redis Exporter<br/>:9121]
        RMQExporter[RabbitMQ Exporter<br/>:9419]
    end

    subgraph "External Services"
        S3[S3/MinIO<br/>Media Storage]
        Stripe[Stripe<br/>Payment Gateway]
        SMTP[SMTP/SendGrid<br/>Email Provider]
    end

    %% Client to Gateway
    Client -->|HTTPS| Traefik

    %% Gateway to Services
    Traefik -->|HTTP| BFF
    Traefik -->|HTTP| UserService
    Traefik -->|HTTP| LessonService
    Traefik -->|HTTP| ContentService
    Traefik -->|HTTP| NotificationService
    Traefik -->|HTTP| OrderService

    %% BFF to Services (Synchronous)
    BFF -->|HTTP| UserService
    BFF -->|HTTP| LessonService
    BFF -->|HTTP| ContentService
    BFF -->|HTTP| OrderService

    %% Services to Databases
    UserService -->|SQL| PostgreSQL
    LessonService -->|SQL| PostgreSQL
    OrderService -->|SQL| PostgreSQL
    ContentService -->|NoSQL| MongoDB
    ContentService -->|SQL| PostgreSQL

    %% Services to Cache
    UserService -->|Cache| Redis
    LessonService -->|Cache| Redis
    ContentService -->|Cache| Redis
    BFF -->|Cache| Redis
    OrderService -->|Cache| Redis

    %% Services to Message Broker (Asynchronous)
    UserService -.->|Publish Events| RabbitMQ
    LessonService -.->|Publish Events| RabbitMQ
    ContentService -.->|Publish Events| RabbitMQ
    OrderService -.->|Publish Events| RabbitMQ
    RabbitMQ -.->|Subscribe| NotificationService
    RabbitMQ -.->|Subscribe| LessonService
    RabbitMQ -.->|Subscribe| UserService

    %% External Integrations
    ContentService -->|Upload/Download| S3
    OrderService -->|Payment Processing| Stripe
    NotificationService -->|Send Email| SMTP

    %% Monitoring Connections
    PGExporter -->|Scrape| PostgreSQL
    RedisExporter -->|Scrape| Redis
    RMQExporter -->|Scrape| RabbitMQ
    
    Prometheus -->|Scrape Metrics| PGExporter
    Prometheus -->|Scrape Metrics| RedisExporter
    Prometheus -->|Scrape Metrics| RMQExporter
    Prometheus -->|Scrape Metrics| UserService
    Prometheus -->|Scrape Metrics| LessonService
    Prometheus -->|Scrape Metrics| ContentService
    Prometheus -->|Scrape Metrics| NotificationService
    Prometheus -->|Scrape Metrics| OrderService
    Prometheus -->|Scrape Metrics| BFF
    
    Grafana -->|Query| Prometheus
    Grafana -->|Query| Loki
    Loki -->|Collect Logs| UserService
    Loki -->|Collect Logs| LessonService
    Loki -->|Collect Logs| ContentService
    Loki -->|Collect Logs| NotificationService
    Loki -->|Collect Logs| OrderService
    Loki -->|Collect Logs| BFF

    %% Styling
    classDef gateway fill:#ff6b6b,stroke:#c92a2a,stroke-width:2px,color:#fff
    classDef service fill:#4dabf7,stroke:#1971c2,stroke-width:2px,color:#fff
    classDef database fill:#51cf66,stroke:#2f9e44,stroke-width:2px,color:#fff
    classDef messaging fill:#ffd43b,stroke:#f59f00,stroke-width:2px,color:#000
    classDef monitoring fill:#a78bfa,stroke:#7c3aed,stroke-width:2px,color:#fff
    classDef external fill:#ffa94d,stroke:#e67700,stroke-width:2px,color:#000
    
    class Traefik gateway
    class BFF,UserService,LessonService,ContentService,NotificationService,OrderService service
    class PostgreSQL,MongoDB,Redis database
    class RabbitMQ messaging
    class Prometheus,Grafana,Loki,PGExporter,RedisExporter,RMQExporter monitoring
    class S3,Stripe,SMTP external
```

### Legend
- **Solid Lines (→)**: Synchronous HTTP/gRPC communication
- **Dotted Lines (⋯→)**: Asynchronous event-driven communication via message broker
- **Colors**:
  - 🔴 **Red**: API Gateway
  - 🔵 **Blue**: Microservices
  - 🟢 **Green**: Data stores (PostgreSQL, MongoDB, Redis)
  - 🟡 **Yellow**: Message broker (RabbitMQ)
  - 🟣 **Purple**: Monitoring & Observability
  - 🟠 **Orange**: External services

> **Note:** The architecture currently uses REST/HTTP for service communication. For improved performance, see the [gRPC Implementation Guide](#grpc-implementation-guide) to convert internal service-to-service calls to gRPC while maintaining REST for external APIs.

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
  - Tracks email delivery and status
  - Integrations: SMTP or SendGrid/Mailgun
- **Purpose:** Keeps users engaged & informed through timely notifications.

### 5. Order Service

- **Features:**
  - Payment processing with Stripe integration
  - Order management and tracking
  - Subscription and course purchase handling
  - Webhook handling for payment events
- **Purpose:** Manages monetary transactions and premium content access.

### 6. BFF Service

- **Purpose:** The Aggregator Service (also called API Composition Layer / Backend-for-Frontend) is responsible for combining data from multiple domain services (User, Lesson, Progress, Content, Order) into a single API response. Instead of the client making multiple calls, the aggregator merges responses and optimizes communication.


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

## Communication Patterns

The architecture follows a hybrid communication pattern as shown in the [Architecture Diagram](#architecture-diagram) above:

### Synchronous Communication (HTTP/REST/GraphQL)
- **Client → Traefik → Services**: All external requests flow through the API gateway
- **BFF → Domain Services**: The BFF aggregates data by making HTTP calls to multiple services
- **Direct Service Access**: Services can be accessed directly through Traefik for specific operations

### Asynchronous Communication (Event-Driven)
- **RabbitMQ Message Broker**: Services publish events to RabbitMQ for decoupled, async operations
- **Event Examples**:
  - `UserCreated`: Published by User Service → Consumed by Notification Service
  - `LessonCompleted`: Published by Lesson Service → Consumed by Notification Service
  - `OrderCompleted`: Published by Order Service → Consumed by Lesson Service (unlock content)

### Data Flow Layers
1. **Gateway Layer**: Traefik handles security, routing, and TLS termination
2. **BFF Layer**: Aggregates and composes responses from multiple services
3. **Service Layer**: Domain services handle business logic independently
4. **Data Layer**: PostgreSQL, MongoDB, and Redis for persistence and caching
5. **Event Layer**: RabbitMQ for asynchronous inter-service communication

---

## gRPC Implementation Guide

### Why gRPC for Microservices?

While the current architecture uses REST/HTTP for synchronous communication, **gRPC** offers significant advantages for internal service-to-service communication:

- **Performance**: Binary protocol (Protocol Buffers) is faster than JSON
- **Type Safety**: Strong typing with .proto definitions
- **Streaming**: Bidirectional streaming support
- **Code Generation**: Auto-generate client/server code in multiple languages
- **Efficiency**: Smaller payload size and HTTP/2 multiplexing

### When to Use gRPC vs REST

| Use Case | Protocol | Reason |
|----------|----------|--------|
| **Client → API Gateway** | REST/HTTP | Better browser/mobile compatibility |
| **API Gateway → Services** | REST/HTTP | Current implementation, simpler debugging |
| **Service → Service** | **gRPC** | Better performance, type safety |
| **BFF → Services** | **gRPC** (Recommended) | High-frequency calls benefit from efficiency |
| **Public APIs** | REST/HTTP | Wider compatibility |

### Converting Services to gRPC

#### Step 1: Define Protocol Buffers

Create `.proto` files to define your service contracts:

**Example: `user-service.proto`**
```protobuf
syntax = "proto3";

package user;
option go_package = "github.com/yourorg/user-services/pb";

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUsersByIds(GetUsersByIdsRequest) returns (GetUsersByIdsResponse);
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  string user_id = 1;
  string email = 2;
  string name = 3;
  string avatar_url = 4;
  string created_at = 5;
  string updated_at = 6;
}

message ValidateTokenRequest {
  string token = 1;
}

message ValidateTokenResponse {
  bool valid = 1;
  string user_id = 2;
  string error_message = 3;
}

message GetUsersByIdsRequest {
  repeated string user_ids = 1;
}

message GetUsersByIdsResponse {
  repeated GetUserResponse users = 1;
}

// ... other messages
```

#### Step 2: Generate Code

**For Go Services (User, Content, Order, BFF):**
```bash
# Install protoc compiler and Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/user-service.proto
```

**For Python Service (Lesson):**
```bash
# Install gRPC tools
pip install grpcio grpcio-tools

# Generate code
python -m grpc_tools.protoc -I./proto \
       --python_out=./pb \
       --grpc_python_out=./pb \
       proto/lesson-service.proto
```

**For Node.js Service (Notification):**
```bash
# Install gRPC packages
npm install @grpc/grpc-js @grpc/proto-loader

# Use dynamic loading or generate with protoc
npm install -g grpc-tools
grpc_tools_node_protoc --js_out=import_style=commonjs,binary:./pb \
       --grpc_out=grpc_js:./pb \
       --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` \
       proto/notification-service.proto
```

#### Step 3: Implement gRPC Server

**Go Example (User Service):**
```go
package main

import (
    "context"
    "log"
    "net"
    
    "google.golang.org/grpc"
    pb "github.com/yourorg/user-services/pb"
)

type userServiceServer struct {
    pb.UnimplementedUserServiceServer
    // your dependencies (DB, Redis, etc.)
}

func (s *userServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // Implement your logic
    user, err := s.userRepo.GetByID(ctx, req.UserId)
    if err != nil {
        return nil, err
    }
    
    return &pb.GetUserResponse{
        UserId:    user.ID,
        Email:     user.Email,
        Name:      user.Name,
        AvatarUrl: user.AvatarURL,
        CreatedAt: user.CreatedAt.String(),
        UpdatedAt: user.UpdatedAt.String(),
    }, nil
}

func main() {
    // Start gRPC server alongside REST server
    grpcServer := grpc.NewServer()
    pb.RegisterUserServiceServer(grpcServer, &userServiceServer{})
    
    listener, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    log.Println("gRPC server listening on :50051")
    if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
```

**Python Example (Lesson Service):**
```python
import grpc
from concurrent import futures
import lesson_service_pb2
import lesson_service_pb2_grpc

class LessonServiceServicer(lesson_service_pb2_grpc.LessonServiceServicer):
    def GetLessonProgress(self, request, context):
        # Implement your logic
        progress = get_progress_from_db(request.user_id, request.lesson_id)
        
        return lesson_service_pb2.GetLessonProgressResponse(
            user_id=progress.user_id,
            lesson_id=progress.lesson_id,
            completed=progress.completed,
            score=progress.score
        )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    lesson_service_pb2_grpc.add_LessonServiceServicer_to_server(
        LessonServiceServicer(), server
    )
    server.add_insecure_port('[::]:50052')
    server.start()
    print("gRPC server listening on :50052")
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
```

#### Step 4: Implement gRPC Client

**BFF Service calling User Service:**
```go
package main

import (
    "context"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/yourorg/user-services/pb"
)

type UserServiceClient struct {
    client pb.UserServiceClient
}

func NewUserServiceClient(address string) (*UserServiceClient, error) {
    conn, err := grpc.Dial(address, 
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
        grpc.WithTimeout(5*time.Second),
    )
    if err != nil {
        return nil, err
    }
    
    return &UserServiceClient{
        client: pb.NewUserServiceClient(conn),
    }, nil
}

func (c *UserServiceClient) GetUser(ctx context.Context, userID string) (*pb.GetUserResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    
    return c.client.GetUser(ctx, &pb.GetUserRequest{
        UserId: userID,
    })
}

func (c *UserServiceClient) GetUsersByIds(ctx context.Context, userIDs []string) ([]*pb.GetUserResponse, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    resp, err := c.client.GetUsersByIds(ctx, &pb.GetUsersByIdsRequest{
        UserIds: userIDs,
    })
    if err != nil {
        return nil, err
    }
    
    return resp.Users, nil
}
```

### Docker Compose Configuration for gRPC

Update `infrastructure/docker-compose.yml` to expose gRPC ports:

```yaml
services:
  user-services:
    # ... existing config
    ports:
      - "8001:8001"  # REST API
      - "50051:50051"  # gRPC
    environment:
      - GRPC_PORT=50051
      - REST_PORT=8001
      # ... other env vars

  lesson-services:
    # ... existing config
    ports:
      - "8005:8005"  # REST API
      - "50052:50052"  # gRPC
    environment:
      - GRPC_PORT=50052
      - REST_PORT=8005

  content-services:
    ports:
      - "8004:8004"  # REST API
      - "50053:50053"  # gRPC
    environment:
      - GRPC_PORT=50053

  order-services:
    ports:
      - "8006:8006"  # REST API
      - "50054:50054"  # gRPC
    environment:
      - GRPC_PORT=50054

  bff-services:
    environment:
      # gRPC endpoints for internal calls
      - USER_SERVICE_GRPC=user-services:50051
      - LESSON_SERVICE_GRPC=lesson-services:50052
      - CONTENT_SERVICE_GRPC=content-services:50053
      - ORDER_SERVICE_GRPC=order-services:50054
      # REST endpoints for fallback/external
      - USER_SERVICE_URL=http://user-services:8001
```

### Migration Strategy

**Phase 1: Dual Protocol (Recommended)**
- Keep REST endpoints for external/client access
- Add gRPC servers alongside REST servers
- Gradually migrate internal service-to-service calls to gRPC

**Phase 2: BFF Migration**
- Update BFF service to use gRPC for internal calls
- Keep REST for external API gateway access
- Monitor performance improvements

**Phase 3: Full Internal gRPC**
- All service-to-service communication uses gRPC
- REST only exposed through API gateway for external clients
- GraphQL remains for flexible client queries (Content Service)

### Testing gRPC Services

**Using grpcurl (like curl for gRPC):**
```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Call a method
grpcurl -plaintext -d '{"user_id": "123"}' \
  localhost:50051 user.UserService/GetUser
```

**Using BloomRPC or Postman:**
- BloomRPC: GUI client for gRPC (like Postman for REST)
- Postman: Now supports gRPC in recent versions

### Benefits for This Architecture

1. **BFF Performance**: BFF makes multiple service calls - gRPC reduces latency by ~30-50%
2. **Type Safety**: Protocol Buffers prevent serialization errors
3. **Bandwidth**: Smaller payloads reduce network costs
4. **Streaming**: Future features like real-time progress updates
5. **Polyglot Support**: Go, Python, Node.js all have excellent gRPC support

### Service Port Allocation

| Service | REST Port | gRPC Port |
|---------|-----------|-----------|
| User Service | 8001 | 50051 |
| Lesson Service | 8005 | 50052 |
| Content Service | 8004 | 50053 |
| Order Service | 8006 | 50054 |
| Notification Service | 8003 | 50055 |
| BFF Service | 8010 | - (client only) |

---

## Getting Started

### Prerequisites
- Docker and Docker Compose installed
- At least 8GB of RAM available
- Ports 80, 443, 3000, 5432, 6379, 8001-8010, 9090, 15672, 27017 available

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd english-learning-app
   ```

2. **Start the infrastructure stack**
   ```bash
   cd infrastructure
   docker-compose up -d
   ```

3. **Verify services are running**
   ```bash
   docker-compose ps
   ```

4. **Access the services**
   - **API Gateway (Traefik)**: http://localhost:80
   - **Traefik Dashboard**: http://localhost:8080
   - **Grafana**: http://localhost:3000 (admin/admin)
   - **Prometheus**: http://localhost:9090
   - **RabbitMQ Management**: http://localhost:15672 (user/password)
   - **pgAdmin**: http://localhost:5050 (admin@example.com/password)

5. **View logs**
   ```bash
   docker-compose logs -f [service-name]
   ```

6. **Stop the stack**
   ```bash
   docker-compose down
   ```

### Service Endpoints

Once running, the services are accessible through Traefik at:
- **User Service**: http://localhost/api/user
- **Lesson Service**: http://localhost/api/lesson
- **Content Service**: http://localhost/api/content
- **Notification Service**: http://localhost/api/notification
- **Order Service**: http://localhost/api/order
- **BFF Service**: http://localhost/api/bff

### Development

For detailed development instructions, see [CLAUDE.md](./CLAUDE.md) which includes:
- Service-specific build and run commands
- Database migration instructions
- Testing guidelines
- Architecture patterns and best practices

---

## Additional Documentation

- **[Authentication Guide](./authentication.md)**: JWT, MFA, and session management
- **[Admin Features](./admin.md)**: Admin panel and management features
- **[Frontend Integration](./FRONTEND_INTEGRATION_README.md)**: Guide for integrating with frontend applications
- **[Security Report](./security-report.md)**: Security audit and recommendations

---

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]