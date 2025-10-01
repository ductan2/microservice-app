# Lesson/Progress Service

A comprehensive FastAPI service for managing learning progress, gamification, and spaced repetition in an English learning application.

## Features

### 📚 Lesson Progress Tracking
- Track user lesson enrollment and completion
- Monitor progress through lesson sections
- Calculate scores and completion status
- Support for abandoned lessons

### 🎯 Quiz Management
- Handle quiz attempts with multiple tries
- Track answers and scoring
- Calculate pass/fail status (70% threshold)
- Support for both multiple choice and text answers

### 🧠 Spaced Repetition (SM-2 Algorithm)
- Implement SM-2 algorithm for optimal learning intervals
- Track flashcard reviews and difficulty adjustments
- Automatic scheduling of due cards
- Support for card suspension

### 🏆 Gamification & Leaderboards
- Daily activity tracking (lessons, quizzes, minutes, points)
- Streak calculation and maintenance
- Point system (lifetime, weekly, monthly)
- Leaderboard snapshots for performance

### 📊 Analytics & Events
- Progress event tracking for all major actions
- Outbox pattern for event publishing
- User statistics aggregation
- Activity analytics

## Database Schema

The service uses PostgreSQL with the following main entities:

- **user_lessons**: Lesson enrollment and progress
- **quiz_attempts**: Quiz attempt tracking
- **quiz_answers**: Individual quiz responses
- **sr_cards**: Spaced repetition cards
- **sr_reviews**: Review history
- **daily_activity**: Daily learning metrics
- **user_streaks**: Learning streaks
- **user_points**: Point accumulation
- **leaderboard_snapshots**: Leaderboard data

## API Endpoints

### Lesson Progress
- `POST /api/v1/progress/lessons/start` - Start a new lesson
- `PUT /api/v1/progress/lessons/{user_id}/{lesson_id}/progress` - Update lesson progress
- `GET /api/v1/progress/lessons/user/{user_id}` - Get user's lessons

### Quiz Management
- `POST /api/v1/progress/quiz/attempts` - Start quiz attempt
- `POST /api/v1/progress/quiz/attempts/{attempt_id}/submit` - Submit quiz answers

### Spaced Repetition
- `POST /api/v1/progress/spaced-repetition/cards` - Create SR card
- `POST /api/v1/progress/spaced-repetition/review` - Review flashcard
- `GET /api/v1/progress/spaced-repetition/due/{user_id}` - Get due cards

### Statistics & Leaderboards
- `GET /api/v1/progress/users/{user_id}/stats` - User statistics
- `GET /api/v1/progress/users/{user_id}/points` - User points
- `GET /api/v1/progress/users/{user_id}/streak` - User streak
- `GET /api/v1/progress/leaderboard/{period}/{period_key}` - Leaderboard data

## Installation

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. Run database migrations:
```bash
alembic upgrade head
```

4. Start the application:
```bash
uvicorn main:app --reload
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string (for caching)
- `SECRET_KEY`: JWT secret key
- `ENVIRONMENT`: Application environment (development/production)

## Development

The application follows clean architecture principles:

- **Models**: SQLAlchemy ORM models (`app/models/`)
- **Schemas**: Pydantic request/response models (`app/schemas/`)
- **Services**: Business logic layer (`app/services/`)
- **Routes**: FastAPI endpoints (`app/routers/`)

## Database Migrations

Use Alembic for database schema management:

```bash
# Create new migration
alembic revision --autogenerate -m "Description"

# Apply migrations
alembic upgrade head

# Rollback migration
alembic downgrade -1
```

## Testing

Run the test suite:
```bash
pytest
```

## API Documentation

Interactive API documentation is available at:
- Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`