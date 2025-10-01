# Outbox Pattern Implementation

## Tổng quan

Hệ thống sử dụng **Transactional Outbox Pattern** để đảm bảo message được gửi đến RabbitMQ một cách đáng tin cậy khi user đăng ký.

## Luồng hoạt động (Flow)

```
┌──────────────┐
│   Client     │
└──────┬───────┘
       │ POST /register
       ▼
┌─────────────────────────────────────────────┐
│         User Service (Go)                   │
│                                             │
│  1. Validate & Hash Password                │
│  2. Create User                             │
│  3. Create User Profile                     │
│  4. Create Audit Log                        │
│  5. ✅ Create Outbox Event (user.created)   │
│     - Save to outbox table (PostgreSQL)     │
│     - Topic: "user.created"                 │
│     - Payload: {user_id, email, name}       │
│                                             │
│  ┌──────────────────────────────────┐       │
│  │  Outbox Processor (Background)   │       │
│  │  - Runs every 5 seconds          │       │
│  │  - Fetch unpublished events      │       │
│  │  - Publish to RabbitMQ           │       │
│  │  - Mark as published             │       │
│  └───────────┬──────────────────────┘       │
└──────────────┼──────────────────────────────┘
               │
               │ Publish to RabbitMQ
               ▼
┌─────────────────────────────────────────────┐
│          RabbitMQ Broker                    │
│                                             │
│  Exchange: "notifications" (topic)          │
│  Routing Key: "user.created"                │
│                                             │
│  ┌─────────────────────────────────┐        │
│  │  Queue: notifications.user_events│        │
│  └───────────┬─────────────────────┘        │
└──────────────┼──────────────────────────────┘
               │
               │ Consume
               ▼
┌─────────────────────────────────────────────┐
│    Notification Service (Node.js)           │
│                                             │
│  1. Receive user.created event              │
│  2. Parse payload {user_id, email, name}    │
│  3. Create welcome email                    │
│  4. Send email via EmailService             │
│     - SendGrid or SMTP                      │
│  5. Ack message                             │
│                                             │
└──────────────┬──────────────────────────────┘
               │
               ▼
          ┌────────┐
          │  User  │
          │  📧    │
          └────────┘
```

## Chi tiết implementation

### 1. User Service (Go)

#### Outbox Model
```go
type Outbox struct {
    ID          int64          // Auto increment
    AggregateID uuid.UUID      // User ID
    Topic       string         // "user.created"
    Type        string         // "UserCreated"
    Payload     map[string]any // JSON data
    CreatedAt   time.Time
    PublishedAt sql.NullTime   // NULL = chưa publish
}
```

#### Outbox Service
- `ProcessUnpublishedEvents()`: Đọc events chưa publish, gửi lên RabbitMQ
- `publishToRabbitMQ()`: Publish event với routing key = topic

#### Outbox Processor (Worker)
- Chạy background goroutine
- Polling mỗi 5 giây
- Batch size: 10 events/lần
- Graceful shutdown support

#### Configuration
```env
RABBITMQ_URL=amqp://user:password@rabbitmq:5672/
RABBITMQ_EXCHANGE=notifications
```

### 2. Notification Service (Node.js)

#### RabbitMQ Consumer
- Queue: `notifications.user_events`
- Routing Key: `user.created`
- Binding to exchange: `notifications` (topic)

#### Welcome Email Template
- Subject: "Welcome to English Learning App! 🎉"
- HTML formatted với personalization
- Fallback text version

#### Configuration
```env
RABBITMQ_URL=amqp://localhost:5672
RABBITMQ_EXCHANGE=notifications
RABBITMQ_USER_EVENTS_QUEUE=notifications.user_events
RABBITMQ_USER_EVENTS_ROUTING_KEY=user.created
RABBITMQ_PREFETCH=10
```

## Lợi ích của Outbox Pattern

### 1. **Eventual Consistency**
- User được tạo thành công ngay cả khi RabbitMQ down
- Event sẽ được gửi khi RabbitMQ recover

### 2. **At-least-once Delivery**
- Đảm bảo message không bị mất
- Có thể duplicate (consumer phải idempotent)

### 3. **Transactional Guarantee**
- Outbox event được lưu cùng transaction với user
- Hoặc cả 2 thành công, hoặc cả 2 rollback

### 4. **Decoupling**
- User Service không phụ thuộc vào RabbitMQ availability
- Notification Service có thể scale độc lập

### 5. **Observability**
- Track tất cả events trong outbox table
- Dễ debug và monitor

## Testing Flow

### 1. Start services
```bash
# Infrastructure
cd infrastructure
docker-compose up -d postgres rabbitmq

# User Service
cd user-services
go run cmd/server/main.go

# Notification Service
cd notification-services
npm install
npm run dev
```

### 2. Register a user
```bash
curl -X POST http://localhost:8001/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "name": "Test User"
  }'
```

### 3. Check logs

**User Service:**
```
Processing 1 unpublished events
Published event 1 (type=UserCreated) to exchange=notifications with routing_key=user.created
```

**Notification Service:**
```
Received user.created event {"user_id":"...", "email":"test@example.com", "name":"Test User"}
Welcome email sent successfully {"email":"test@example.com"}
```

### 4. Verify database
```sql
-- Check outbox table
SELECT * FROM outbox WHERE topic = 'user.created';
-- published_at should be set after processor runs
```

## Monitoring & Troubleshooting

### Check unpublished events
```sql
SELECT COUNT(*) FROM outbox WHERE published_at IS NULL;
```

### Check RabbitMQ
```bash
# Management UI: http://localhost:15672
# Default credentials: guest/guest

# Check queues
docker exec rabbitmq rabbitmqctl list_queues
```

### Common Issues

1. **Events not being published**
   - Check RabbitMQ connection
   - Check outbox processor logs
   - Verify RABBITMQ_EXCHANGE config

2. **Emails not sent**
   - Check notification service logs
   - Verify EMAIL_PROVIDER config (sendgrid/smtp)
   - Check SMTP credentials

3. **Duplicate emails**
   - Normal for at-least-once delivery
   - Add idempotency key if needed

## Environment Variables

### User Service
```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=english_app

# RabbitMQ
RABBITMQ_URL=amqp://user:password@rabbitmq:5672/
RABBITMQ_EXCHANGE=notifications

# Server
PORT=8001
```

### Notification Service
```env
# Email Provider
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_DEFAULT_FROM=noreply@example.com

# RabbitMQ
RABBITMQ_URL=amqp://user:password@rabbitmq:5672
RABBITMQ_EXCHANGE=notifications
RABBITMQ_USER_EVENTS_QUEUE=notifications.user_events
RABBITMQ_USER_EVENTS_ROUTING_KEY=user.created
RABBITMQ_PREFETCH=10

# Server
PORT=3000
```

## Next Steps

- [ ] Add retry mechanism với exponential backoff
- [ ] Add dead letter queue cho failed messages
- [ ] Add metrics (Prometheus) cho outbox lag
- [ ] Add email template engine
- [ ] Add email delivery status tracking

