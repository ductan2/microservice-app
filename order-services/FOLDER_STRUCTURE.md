# Cấu trúc Folder - Order Services

Tài liệu này giải thích mục đích của từng folder trong project `order-services`.

## 📁 Cấu trúc tổng quan

```
order-services/
├── cmd/                    # Entry point của application
├── internal/               # Code nội bộ (không export ra ngoài)
│   ├── api/                # Layer xử lý HTTP API
│   ├── cache/              # Redis cache logic
│   ├── config/             # Configuration management
│   ├── db/                 # Database connection & migration
│   ├── models/             # Database models (GORM)
│   ├── queue/              # RabbitMQ message queue
│   ├── server/             # HTTP server setup & routing
│   ├── utils/              # Utility functions
│   └── worker/             # Background workers
├── Dockerfile              # Production Docker image
├── Dockerfile.dev         # Development Docker image
├── Makefile               # Build commands
└── README.md              # Documentation
```

---

## 📂 Chi tiết từng folder

### `cmd/server/`
**Mục đích:** Entry point của application, chứa file `main.go`
- Khởi tạo database connection
- Khởi tạo Redis, RabbitMQ
- Setup router và middleware
- Start HTTP server
- Graceful shutdown handling

**Ví dụ:** `main.go` - hàm `main()` chạy đầu tiên khi start service

---

### `internal/api/controllers/`
**Mục đích:** HTTP Controllers - xử lý HTTP requests/responses
- Nhận request từ client
- Validate input
- Gọi service layer để xử lý business logic
- Trả về response (JSON, status code)

**Ví dụ:** 
- `order_controller.go` - xử lý các endpoint `/api/v1/orders`
- `payment_controller.go` - xử lý payment endpoints

---

### `internal/api/dto/`
**Mục đích:** Data Transfer Objects - định nghĩa cấu trúc data cho API
- Request DTOs (input từ client)
- Response DTOs (output trả về client)
- Khác với models ở chỗ DTOs chỉ dùng cho API layer

**Ví dụ:**
- `create_order_dto.go` - struct cho request tạo order
- `order_response_dto.go` - struct cho response trả về

---

### `internal/api/helpers/`
**Mục đích:** Helper functions hỗ trợ cho API layer
- Format response
- Parse request
- Validation helpers
- Common utilities cho controllers

**Ví dụ:** `response_helper.go` - format JSON response chuẩn

---

### `internal/api/middleware/`
**Mục đích:** HTTP Middleware - xử lý trước/sau request
- Authentication middleware (check JWT token)
- Authorization middleware (check permissions)
- Logging middleware
- CORS middleware
- Rate limiting

**Ví dụ:** `auth_middleware.go` - verify JWT token trước khi vào controller

---

### `internal/api/repositories/`
**Mục đích:** Data Access Layer - tương tác trực tiếp với database
- CRUD operations (Create, Read, Update, Delete)
- Database queries
- Transaction handling
- Sử dụng GORM để query database

**Ví dụ:**
- `order_repository.go` - các hàm `CreateOrder()`, `GetOrderByID()`, `UpdateOrder()`
- `payment_repository.go` - các hàm liên quan đến payment data

---

### `internal/api/routes/`
**Mục đích:** Định nghĩa HTTP routes và mapping với controllers
- Đăng ký routes (GET, POST, PUT, DELETE)
- Gán middleware cho routes
- Group routes theo prefix (ví dụ: `/api/v1/orders`)

**Ví dụ:** `order_routes.go` - định nghĩa tất cả routes liên quan đến orders

---

### `internal/api/services/`
**Mục đích:** Business Logic Layer - chứa logic nghiệp vụ
- Xử lý business rules
- Orchestrate nhiều repositories
- Validate business logic
- Gọi external services nếu cần
- Không biết về HTTP, chỉ xử lý logic

**Ví dụ:**
- `order_service.go` - logic tạo order, validate, tính toán giá
- `payment_service.go` - logic xử lý payment, integration với payment gateway

---

### `internal/cache/`
**Mục đích:** Redis cache logic
- Cache operations (get, set, delete)
- Session cache
- Cache strategies (TTL, invalidation)
- Redis client wrapper

**Ví dụ:** `redis.go` - Redis client connection, `order_cache.go` - cache orders

---

### `internal/config/`
**Mục đích:** Configuration management
- Load environment variables
- Database config
- JWT config
- Service configs (ports, timeouts)
- Connection strings

**Ví dụ:** `config.go` - load config từ `.env`, `jwt.go` - JWT settings

---

### `internal/db/`
**Mục đích:** Database connection và migration
- PostgreSQL connection setup
- GORM initialization
- Auto-migration (tạo tables từ models)
- Connection pool management

**Ví dụ:** `postgres.go` - hàm `ConnectPostgres()`, `AutoMigrate()`

---

### `internal/models/`
**Mục đích:** Database models (GORM structs)
- Định nghĩa database tables
- Relationships (has many, belongs to)
- GORM tags (primary key, foreign key, indexes)
- Model structs map trực tiếp với database tables

**Ví dụ:**
- `order_model.go` - struct `Order` với các fields: ID, UserID, Total, Status
- `order_item_model.go` - struct `OrderItem` với relationship đến `Order`

---

### `internal/queue/`
**Mục đích:** Message Queue (RabbitMQ) integration
- RabbitMQ connection
- Publish messages
- Consume messages
- Queue/exchange declarations
- Event publishing cho cross-service communication

**Ví dụ:** `rabbitmq.go` - connection và publish events như `OrderCreated`, `OrderPaid`

---

### `internal/server/`
**Mục đích:** HTTP Server setup và routing
- Gin router initialization
- Register all routes
- Setup middleware chain
- Dependency injection (repositories, services, controllers)
- Server configuration

**Ví dụ:** `router.go` - hàm `NewRouter()` setup toàn bộ routes và dependencies

---

### `internal/utils/`
**Mục đích:** Utility functions dùng chung
- Helper functions không thuộc layer cụ thể
- String manipulation
- Date/time utilities
- Validation helpers
- Error handling utilities

**Ví dụ:**
- `env.go` - read environment variables
- `validation.go` - validate email, phone number
- `response.go` - format HTTP responses

---

### `internal/worker/`
**Mục đích:** Background workers xử lý async tasks
- Outbox pattern processor
- Scheduled jobs
- Event processors
- Background tasks không block HTTP requests

**Ví dụ:** `outbox_processor.go` - worker đọc từ outbox table và publish events

---

## 🔄 Flow xử lý request

```
Client Request
    ↓
[Routes] → định nghĩa endpoint
    ↓
[Middleware] → auth, logging, CORS
    ↓
[Controller] → nhận request, validate input
    ↓
[Service] → business logic
    ↓
[Repository] → database operations
    ↓
[Models] → database tables
    ↓
Response ← Controller trả về
```

---

## 📝 Notes

- **internal/**: Code không export ra ngoài, chỉ dùng trong service này
- **cmd/**: Entry point, chỉ có `main.go`
- **api/**: Tất cả code liên quan đến HTTP API
- **models/**: Database schema, không phải DTOs
- **services/**: Business logic, không biết về HTTP
- **repositories/**: Data access, chỉ biết về database

