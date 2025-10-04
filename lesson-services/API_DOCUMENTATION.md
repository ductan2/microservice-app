# Lesson Services API Documentation

## Base URL
```
http://localhost:8000/api/v1
```

## Authentication
Currently no authentication required (development mode).

---

## User Preferences API

### Get User Preferences
```http
GET /api/v1/users/{user_id}
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "locale": "en",
  "level_hint": "beginner",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Create User Preferences
```http
POST /api/v1/users
```

**Payload:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "locale": "en",
  "level_hint": "beginner"
}
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "locale": "en",
  "level_hint": "beginner",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Update User Preferences
```http
PUT /api/v1/users/{user_id}
```

**Payload:**
```json
{
  "locale": "vi",
  "level_hint": "intermediate"
}
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "locale": "vi",
  "level_hint": "intermediate",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### Update User Locale
```http
PATCH /api/v1/users/{user_id}/locale
```

**Payload:**
```json
{
  "locale": "fr"
}
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "locale": "fr",
  "level_hint": "intermediate",
  "updated_at": "2024-01-15T10:40:00Z"
}
```

### Delete User Preferences
```http
DELETE /api/v1/users/{user_id}
```

**Response:**
```
Status: 204 No Content
```

---

## Daily Activity API

### Get Today's Activity
```http
GET /api/v1/daily-activity/user/{user_id}/today
```

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "activity_dt": "2024-01-15",
  "lessons_completed": 3,
  "quizzes_completed": 2,
  "minutes": 45,
  "points": 150
}
```

### Get Activity by Date
```http
GET /api/v1/daily-activity/user/{user_id}/date/{activity_date}
```

**Parameters:**
- `activity_date`: Date in YYYY-MM-DD format

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "activity_dt": "2024-01-14",
  "lessons_completed": 2,
  "quizzes_completed": 1,
  "minutes": 30,
  "points": 100
}
```

### Get Activity Range
```http
GET /api/v1/daily-activity/user/{user_id}/range?date_from=2024-01-01&date_to=2024-01-15
```

**Query Parameters:**
- `date_from` (optional): Start date (defaults to 30 days ago)
- `date_to` (optional): End date (defaults to today)

**Response:**
```json
[
  {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "activity_dt": "2024-01-01",
    "lessons_completed": 1,
    "quizzes_completed": 0,
    "minutes": 15,
    "points": 50
  },
  {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "activity_dt": "2024-01-02",
    "lessons_completed": 2,
    "quizzes_completed": 1,
    "minutes": 30,
    "points": 100
  }
]
```

### Get Week Activity
```http
GET /api/v1/daily-activity/user/{user_id}/week
```

**Response:**
```json
[
  {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "activity_dt": "2024-01-08",
    "lessons_completed": 0,
    "quizzes_completed": 0,
    "minutes": 0,
    "points": 0
  },
  {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "activity_dt": "2024-01-09",
    "lessons_completed": 2,
    "quizzes_completed": 1,
    "minutes": 25,
    "points": 75
  }
]
```

### Get Month Activity
```http
GET /api/v1/daily-activity/user/{user_id}/month?year=2024&month=1
```

**Query Parameters:**
- `year` (optional): Year (defaults to current year)
- `month` (optional): Month 1-12 (defaults to current month)

**Response:**
```json
{
  "year": 2024,
  "month": 1,
  "totals": {
    "lessons_completed": 45,
    "quizzes_completed": 20,
    "minutes": 675,
    "points": 2250
  },
  "days": [
    {
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "activity_dt": "2024-01-01",
      "lessons_completed": 2,
      "quizzes_completed": 1,
      "minutes": 30,
      "points": 100
    }
  ]
}
```

### Get Activity Summary
```http
GET /api/v1/daily-activity/user/{user_id}/stats/summary
```

**Response:**
```json
{
  "lifetime": {
    "lessons_completed": 150,
    "quizzes_completed": 75,
    "minutes": 2250,
    "points": 7500
  },
  "last_7_days": {
    "lessons_completed": 12,
    "quizzes_completed": 6,
    "minutes": 180,
    "points": 600
  },
  "last_30_days": {
    "lessons_completed": 45,
    "quizzes_completed": 20,
    "minutes": 675,
    "points": 2250
  },
  "average_per_day": {
    "lessons_completed": 3,
    "quizzes_completed": 1,
    "minutes": 45,
    "points": 150
  },
  "total_active_days": 50,
  "most_active_day": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "activity_dt": "2024-01-10",
    "lessons_completed": 5,
    "quizzes_completed": 3,
    "minutes": 90,
    "points": 300
  }
}
```

### Increment Activity
```http
POST /api/v1/daily-activity/increment
```

**Payload:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "activityDate": "2024-01-15",
  "field": "lessons_completed",
  "amount": 1
}
```

**Valid fields:**
- `lessons_completed`
- `quizzes_completed`
- `minutes`
- `points`

**Response:**
```json
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "activity_dt": "2024-01-15",
  "lessons_completed": 4,
  "quizzes_completed": 2,
  "minutes": 45,
  "points": 150
}
```

---

## Health Check

### Health Status
```http
GET /api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "lesson-services"
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "detail": "date_from must be before or equal to date_to"
}
```

### 404 Not Found
```json
{
  "detail": "User not found"
}
```

### 409 Conflict
```json
{
  "detail": "User preferences already exist"
}
```

### 422 Validation Error
```json
{
  "detail": [
    {
      "loc": ["body", "user_id"],
      "msg": "field required",
      "type": "value_error.missing"
    }
  ]
}
```

---

## Data Models

### DimUser
```json
{
  "user_id": "UUID",
  "locale": "string",
  "level_hint": "string",
  "updated_at": "datetime"
}
```

### DailyActivity
```json
{
  "user_id": "UUID",
  "activity_dt": "date",
  "lessons_completed": "integer",
  "quizzes_completed": "integer",
  "minutes": "integer",
  "points": "integer"
}
```

### DailyTotals
```json
{
  "lessons_completed": "integer",
  "quizzes_completed": "integer",
  "minutes": "integer",
  "points": "integer"
}
```

---

## Development

### Running the Service
```bash
cd lesson-services
python main.py
```

### API Documentation
- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`

### Database
- PostgreSQL database required
- Connection string in `app/config.py`
- Run migrations with Alembic
