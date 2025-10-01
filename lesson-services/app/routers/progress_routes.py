from fastapi import APIRouter, HTTPException, status, Depends, Query
from typing import List, Optional
from uuid import UUID
from sqlalchemy.orm import Session
from app.database.connection import get_db
from app.services.progress_service import ProgressService
from app.schemas.progress_schema import (
    UserLessonCreate, UserLessonUpdate, UserLessonResponse,
    QuizAttemptCreate, QuizAttemptUpdate, QuizAttemptResponse,
    QuizAnswerCreate, QuizAnswerResponse,
    SRCardCreate, SRCardUpdate, SRCardResponse,
    SRReviewCreate, SRReviewResponse,
    DailyActivityResponse, UserStreakResponse, UserPointsResponse,
    LeaderboardResponse, ProgressEventResponse,
    LessonStatus, LeaderboardPeriod
)

router = APIRouter()

def get_progress_service(db: Session = Depends(get_db)) -> ProgressService:
    return ProgressService(db)

# User Lesson Progress Endpoints
@router.post("/lessons/start", status_code=status.HTTP_201_CREATED)
async def start_lesson(
    lesson_data: UserLessonCreate,
    service: ProgressService = Depends(get_progress_service)
) -> UserLessonResponse:
    return await service.start_lesson(lesson_data)

@router.put("/lessons/{user_id}/{lesson_id}/progress", status_code=status.HTTP_200_OK)
async def update_lesson_progress(
    user_id: UUID,
    lesson_id: UUID,
    update_data: UserLessonUpdate,
    service: ProgressService = Depends(get_progress_service)
) -> UserLessonResponse:
    result = await service.update_lesson_progress(user_id, lesson_id, update_data)
    if not result:
        raise HTTPException(status_code=404, detail="User lesson not found")
    return result

@router.get("/lessons/user/{user_id}", status_code=status.HTTP_200_OK)
async def get_user_lessons(
    user_id: UUID,
    status_filter: Optional[LessonStatus] = Query(None, alias="status"),
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    service: ProgressService = Depends(get_progress_service)
) -> List[UserLessonResponse]:
    return await service.get_user_lessons(user_id, status_filter, skip, limit)

# Quiz Attempt Endpoints
@router.post("/quiz/attempts", status_code=status.HTTP_201_CREATED)
async def start_quiz_attempt(
    attempt_data: QuizAttemptCreate,
    service: ProgressService = Depends(get_progress_service)
) -> QuizAttemptResponse:
    return await service.start_quiz_attempt(attempt_data)

@router.post("/quiz/attempts/{attempt_id}/submit", status_code=status.HTTP_200_OK)
async def submit_quiz_attempt(
    attempt_id: UUID,
    answers: List[QuizAnswerCreate],
    service: ProgressService = Depends(get_progress_service)
) -> QuizAttemptResponse:
    try:
        return await service.submit_quiz_attempt(attempt_id, answers)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))

# Spaced Repetition Endpoints
@router.post("/spaced-repetition/cards", status_code=status.HTTP_201_CREATED)
async def create_sr_card(
    card_data: SRCardCreate,
    service: ProgressService = Depends(get_progress_service)
) -> SRCardResponse:
    return await service.create_sr_card(card_data)

@router.post("/spaced-repetition/review", status_code=status.HTTP_200_OK)
async def review_flashcard(
    user_id: UUID,
    flashcard_id: UUID,
    quality: int = Query(..., ge=0, le=5),
    service: ProgressService = Depends(get_progress_service)
) -> SRCardResponse:
    try:
        return await service.review_flashcard(user_id, flashcard_id, quality)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))

@router.get("/spaced-repetition/due/{user_id}", status_code=status.HTTP_200_OK)
async def get_due_cards(
    user_id: UUID,
    limit: int = Query(50, ge=1, le=200),
    service: ProgressService = Depends(get_progress_service)
) -> List[SRCardResponse]:
    return await service.get_due_cards(user_id, limit)

# Leaderboard Endpoints
@router.get("/leaderboard/{period}/{period_key}", status_code=status.HTTP_200_OK)
async def get_leaderboard(
    period: LeaderboardPeriod,
    period_key: str,
    limit: int = Query(100, ge=1, le=500),
    service: ProgressService = Depends(get_progress_service)
) -> LeaderboardResponse:
    result = await service.get_leaderboard(period, period_key, limit)
    if not result:
        raise HTTPException(status_code=404, detail="Leaderboard not found")
    return result

# User Statistics Endpoints
@router.get("/users/{user_id}/stats", status_code=status.HTTP_200_OK)
async def get_user_stats(
    user_id: UUID,
    service: ProgressService = Depends(get_progress_service)
) -> dict:
    return await service.get_user_stats(user_id)

@router.get("/users/{user_id}/points", status_code=status.HTTP_200_OK)
async def get_user_points(
    user_id: UUID,
    service: ProgressService = Depends(get_progress_service)
) -> Optional[UserPointsResponse]:
    stats = await service.get_user_stats(user_id)
    return stats.get("points")

@router.get("/users/{user_id}/streak", status_code=status.HTTP_200_OK)
async def get_user_streak(
    user_id: UUID,
    service: ProgressService = Depends(get_progress_service)
) -> Optional[UserStreakResponse]:
    stats = await service.get_user_stats(user_id)
    return stats.get("streak")

@router.get("/users/{user_id}/daily-activity", status_code=status.HTTP_200_OK)
async def get_user_daily_activity(
    user_id: UUID,
    service: ProgressService = Depends(get_progress_service)
) -> List[DailyActivityResponse]:
    stats = await service.get_user_stats(user_id)
    return stats.get("recent_activity", [])