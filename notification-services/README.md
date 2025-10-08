# Notification Service

This service handles email notifications and user notification management using SendGrid or SMTP providers, with PostgreSQL for data persistence.

## Features

- Email sending via SendGrid or SMTP
- RabbitMQ integration for async processing
- Template-based email system
- Configurable email providers
- **NEW**: User notification management
- **NEW**: Notification templates with CRUD operations
- **NEW**: Read/unread status tracking
- **NEW**: Bulk notification sending

## Environment Variables

```env
PORT=3000
EMAIL_PROVIDER=smtp # or sendgrid

# SendGrid
SENDGRID_API_KEY=your_api_key
SENDGRID_DEFAULT_FROM=noreply@yourapp.com

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_SECURE=false
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_password
SMTP_DEFAULT_FROM=noreply@yourapp.com

# RabbitMQ
RABBITMQ_URL=amqp://localhost:5672
RABBITMQ_EXCHANGE=notifications
RABBITMQ_EMAIL_QUEUE=notifications.email
RABBITMQ_EMAIL_ROUTING_KEY=email.send
RABBITMQ_USER_EVENTS_QUEUE=notifications.user_events
RABBITMQ_USER_EVENTS_ROUTING_KEY=user.created,user.password_reset,user.email_verification
RABBITMQ_PREFETCH=10

# PostgreSQL
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=userdb
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable
```

## Usage

1. Install dependencies:
```bash
npm install
```

2. Set up environment variables (copy from `env.example`)

3. Make sure PostgreSQL is running

4. Start the service:
```bash
npm run dev
```

## API Endpoints

### Email Routes (Legacy)
- `GET /health` -> `{ status: 'ok' }`
- `POST /email/send` - Send email
- `POST /email/send-template` - Send templated email
- `GET /email/templates` - List available templates
- `GET /email/templates/:name` - Get specific template

### Notification Management (NEW)

#### Notification Templates
- `POST /api/notifications/templates` - Create notification template
- `GET /api/notifications/templates` - Get all templates with user counts
- `GET /api/notifications/templates/:id` - Get template by ID
- `PUT /api/notifications/templates/:id` - Update template
- `DELETE /api/notifications/templates/:id` - Delete template

#### User Notifications
- `POST /api/notifications/users/:userId/notifications` - Create user notification
- `GET /api/notifications/users/:userId/notifications` - Get user notifications (with pagination and filtering)
- `PUT /api/notifications/users/:userId/notifications/read` - Mark notifications as read
- `GET /api/notifications/users/:userId/notifications/unread-count` - Get unread count
- `DELETE /api/notifications/users/:userId/notifications/:notificationId` - Delete user notification

#### Bulk Operations
- `POST /api/notifications/templates/:templateId/send` - Send notification to multiple users

## Database Schema

The service automatically creates the following tables:

### notification_templates
- `id` (UUID, Primary Key)
- `type` (TEXT) - Notification type (email, push, sms, etc.)
- `title` (TEXT) - Notification title
- `body` (TEXT) - Notification body
- `data` (JSONB) - Additional data
- `created_at` (TIMESTAMPTZ)

### user_notifications
- `id` (UUID, Primary Key)
- `user_id` (UUID) - User ID
- `notification_id` (UUID) - Reference to notification_templates
- `is_read` (BOOLEAN) - Read status
- `created_at` (TIMESTAMPTZ)
- `read_at` (TIMESTAMPTZ)

## Architecture

The service uses:
- Express.js for HTTP API
- PostgreSQL for data persistence
- RabbitMQ for message queuing
- SendGrid/SMTP for email delivery
- Pino for logging
- Zod for validation
- UUID for unique identifiers

## Documentation

See `API_DOCUMENTATION.md` for detailed API documentation with examples.

## Notes
- For Gmail, enable 2FA and create an App Password, then use that for `SMTP_PASS`.
- The service automatically creates database tables and indexes on startup.
- All API responses follow a consistent format with `success` and `data` fields.