# Notification Service API Documentation

## Overview
Notification Service cung cấp API để quản lý notification templates và user notifications.

## Base URL
```
http://localhost:3000/api/notifications
```

## Database Schema

### notification_templates
- `id` (UUID, Primary Key)
- `type` (TEXT) - Loại notification (email, push, sms, etc.)
- `title` (TEXT) - Tiêu đề notification
- `body` (TEXT) - Nội dung notification
- `data` (JSONB) - Dữ liệu bổ sung
- `created_at` (TIMESTAMPTZ)

### user_notifications
- `id` (UUID, Primary Key)
- `user_id` (UUID) - ID của user
- `notification_id` (UUID) - Reference đến notification_templates
- `is_read` (BOOLEAN) - Trạng thái đã đọc
- `created_at` (TIMESTAMPTZ)
- `read_at` (TIMESTAMPTZ)

## API Endpoints

### Notification Templates

#### 1. Create Notification Template
```http
POST /api/notifications/templates
Content-Type: application/json

{
  "type": "email",
  "title": "Welcome to our app",
  "body": "Thank you for joining us!",
  "data": {
    "priority": "high",
    "category": "welcome"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "type": "email",
    "title": "Welcome to our app",
    "body": "Thank you for joining us!",
    "data": {
      "priority": "high",
      "category": "welcome"
    },
    "created_at": "2024-01-01T00:00:00.000Z"
  }
}
```

#### 2. Get All Templates
```http
GET /api/notifications/templates
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "type": "email",
      "title": "Welcome to our app",
      "body": "Thank you for joining us!",
      "data": {...},
      "created_at": "2024-01-01T00:00:00.000Z",
      "user_count": 5
    }
  ]
}
```

#### 3. Get Template by ID
```http
GET /api/notifications/templates/{id}
```

#### 4. Update Template
```http
PUT /api/notifications/templates/{id}
Content-Type: application/json

{
  "title": "Updated title",
  "body": "Updated body"
}
```

#### 5. Delete Template
```http
DELETE /api/notifications/templates/{id}
```

### User Notifications

#### 1. Create User Notification
```http
POST /api/notifications/users/{userId}/notifications
Content-Type: application/json

{
  "notification_id": "template-uuid"
}
```

#### 2. Get User Notifications
```http
GET /api/notifications/users/{userId}/notifications?limit=50&offset=0&is_read=false
```

**Query Parameters:**
- `limit` (optional): Số lượng notifications trả về (default: 50, max: 100)
- `offset` (optional): Số lượng bỏ qua (default: 0)
- `is_read` (optional): Lọc theo trạng thái đã đọc (true/false)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "user_id": "user-uuid",
      "notification_id": "template-uuid",
      "is_read": false,
      "created_at": "2024-01-01T00:00:00.000Z",
      "read_at": null,
      "template": {
        "id": "template-uuid",
        "type": "email",
        "title": "Welcome to our app",
        "body": "Thank you for joining us!",
        "data": {...},
        "created_at": "2024-01-01T00:00:00.000Z"
      }
    }
  ]
}
```

#### 3. Mark Notifications as Read
```http
PUT /api/notifications/users/{userId}/notifications/read
Content-Type: application/json

{
  "notification_ids": ["notification-uuid-1", "notification-uuid-2"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "updated_count": 2
  }
}
```

#### 4. Get Unread Count
```http
GET /api/notifications/users/{userId}/notifications/unread-count
```

**Response:**
```json
{
  "success": true,
  "data": {
    "unread_count": 5
  }
}
```

#### 5. Delete User Notification
```http
DELETE /api/notifications/users/{userId}/notifications/{notificationId}
```

### Bulk Operations

#### Send Notification to Multiple Users
```http
POST /api/notifications/templates/{templateId}/send
Content-Type: application/json

{
  "user_ids": ["user-uuid-1", "user-uuid-2", "user-uuid-3"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "notifications_created": 3,
    "notifications": [...]
  }
}
```

## Error Responses

All endpoints return errors in the following format:

```json
{
  "success": false,
  "error": "Error message",
  "details": "Additional error details (for validation errors)"
}
```

## Environment Variables

```env
# PostgreSQL
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=userdb
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_SSLMODE=disable

# Server
PORT=3000
```

## Installation & Setup

1. Install dependencies:
```bash
npm install
```

2. Set up environment variables in `.env` file

3. Start the service:
```bash
npm run dev
```

## Database Setup

The service will automatically create the required tables and indexes when it starts up. Make sure PostgreSQL is running and accessible with the provided credentials.
