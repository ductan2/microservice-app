from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.middlewares.auth_middleware import InternalAuthRequired
from app.routers import (
    daily_activity_routes,
    dim_user_routes,
    health_routes,
    leaderboard_routes,
    outbox_routes,
    progress_event_routes,
    quiz_answer_routes,
    quiz_attempt_routes,
    sr_card_routes,
    sr_review_routes,
    user_lesson_routes,
    user_points_routes,
    user_streak_routes,
)

app = FastAPI(
    title="Lesson Services API",
    description="A RESTful API for managing English learning lessons",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Add authentication middleware (except for health check)
app.add_middleware(InternalAuthRequired)

app.include_router(health_routes.router, prefix="/api/v1", tags=["health"])
app.include_router(daily_activity_routes.router, prefix="/api/v1", tags=["daily-activity"])
app.include_router(dim_user_routes.router, prefix="/api/v1", tags=["dim-user"])
app.include_router(leaderboard_routes.router, prefix="/api/v1", tags=["leaderboard"])
app.include_router(outbox_routes.router, prefix="/api/v1", tags=["outbox"])
app.include_router(progress_event_routes.router, prefix="/api/v1", tags=["progress-event"])
app.include_router(quiz_answer_routes.router, prefix="/api/v1", tags=["quiz-answer"])
app.include_router(quiz_attempt_routes.router, prefix="/api/v1", tags=["quiz-attempt"])
app.include_router(sr_card_routes.router, prefix="/api/v1", tags=["sr-card"])
app.include_router(sr_review_routes.router, prefix="/api/v1", tags=["sr-review"])
app.include_router(user_lesson_routes.router, prefix="/api/v1", tags=["user-lesson"])
app.include_router(user_points_routes.router, prefix="/api/v1", tags=["user-points"])
app.include_router(user_streak_routes.router, prefix="/api/v1", tags=["user-streak"])

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8005)