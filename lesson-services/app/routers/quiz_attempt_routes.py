from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Request, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.quiz_schema import (
    QuizAttemptCreate,
    QuizAttemptDetailResponse,
    QuizAttemptResponse,
    QuizAttemptSubmit,
)
from app.services.quiz_attempt_service import QuizAttemptService


router = APIRouter(prefix="/api/quiz-attempts", tags=["Quiz Attempts"])


def get_quiz_attempt_service(db: Session = Depends(get_db)) -> QuizAttemptService:
    """Dependency to get QuizAttemptService instance."""
    return QuizAttemptService(db)


@router.post("/start", response_model=QuizAttemptResponse, status_code=status.HTTP_201_CREATED)
def start_quiz_attempt(
    payload: QuizAttemptCreate,
    service: QuizAttemptService = Depends(get_quiz_attempt_service),
) -> QuizAttemptResponse:
    attempt = service.start_quiz(payload)
    return attempt


@router.get("/{attempt_id}", response_model=QuizAttemptDetailResponse)
def get_quiz_attempt(
    attempt_id: UUID, 
    service: QuizAttemptService = Depends(get_quiz_attempt_service)
) -> QuizAttemptDetailResponse:
    attempt = service.get_attempt(attempt_id)
    if not attempt:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz attempt not found")
    return QuizAttemptDetailResponse.model_validate(attempt, from_attributes=True)


@router.post("/{attempt_id}/submit", response_model=QuizAttemptDetailResponse)
def submit_quiz_attempt(
    attempt_id: UUID,
    payload: QuizAttemptSubmit,
    service: QuizAttemptService = Depends(get_quiz_attempt_service),
) -> QuizAttemptDetailResponse:
    attempt = service.submit_quiz(attempt_id, payload)
    if not attempt:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz attempt not found or already submitted")
    # Reload attempt with answers for complete response
    refreshed = service.get_attempt(attempt_id)
    if not refreshed:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz attempt not found")
    return QuizAttemptDetailResponse.model_validate(refreshed, from_attributes=True)


@router.get("/user/me/quiz/{quiz_id}", response_model=List[QuizAttemptResponse])
def get_user_quiz_attempts(
    request: Request,
    quiz_id: UUID,
    service: QuizAttemptService = Depends(get_quiz_attempt_service),
) -> List[QuizAttemptResponse]:
    user_id: UUID = request.state.user_id
    return service.get_user_quiz_attempts(user_id, quiz_id)


@router.get("/user/me/history", response_model=List[QuizAttemptResponse])
def get_user_quiz_history(
    request: Request,
    passed: Optional[bool] = Query(None),
    limit: int = Query(50, ge=1, le=200),
    offset: int = Query(0, ge=0),
    service: QuizAttemptService = Depends(get_quiz_attempt_service),
) -> List[QuizAttemptResponse]:
    user_id: UUID = request.state.user_id
    return service.get_user_quiz_history(user_id, passed=passed, limit=limit, offset=offset)


@router.get("/lesson/{lesson_id}/user/me", response_model=List[QuizAttemptResponse])
def get_lesson_quiz_attempts(
    lesson_id: UUID,
    request: Request,
    service: QuizAttemptService = Depends(get_quiz_attempt_service),
) -> List[QuizAttemptResponse]:
    user_id: UUID = request.state.user_id
    return service.get_lesson_quiz_attempts(lesson_id, user_id)


@router.delete("/{attempt_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_quiz_attempt(
    attempt_id: UUID, 
    service: QuizAttemptService = Depends(get_quiz_attempt_service)
) -> Response:
    deleted = service.delete_attempt(attempt_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Quiz attempt not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)

