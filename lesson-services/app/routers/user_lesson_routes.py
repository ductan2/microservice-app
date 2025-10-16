from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, Query, Response, status
from sqlalchemy.orm import Session
from typing import List, Optional
from uuid import UUID

from app.database.connection import get_db
from app.schemas.user_lesson_schema import (
    LessonStatus,
    UserLessonCompletionRequest,
    UserLessonCreate,
    UserLessonResponse,
    UserLessonStats,
    UserLessonUpdate,
)
from app.services.user_lesson_service import UserLessonService
from app.middlewares.auth_middleware import get_current_user_id


router = APIRouter(prefix="/api/user-lessons", tags=["User Lesson Progress"])


def get_user_lesson_service(db: Session = Depends(get_db)) -> UserLessonService:
    return UserLessonService(db)


@router.get("", response_model=List[UserLessonResponse])
def list_user_lessons(
    status: Optional[LessonStatus] = Query(default=None),
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> List[UserLessonResponse]:
    return service.get_user_lessons(user_id, status=status)


@router.get("/lesson/{lesson_id}", response_model=UserLessonResponse)
def get_user_lesson(
    lesson_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> UserLessonResponse:
    lesson = service.get_user_lesson(user_id, lesson_id)
    if lesson is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User lesson not found")
    return lesson


@router.post("/start", response_model=UserLessonResponse, status_code=status.HTTP_201_CREATED)
def start_user_lesson(
    payload: UserLessonCreate,
    service: UserLessonService = Depends(get_user_lesson_service),
) -> UserLessonResponse:
    return service.start_lesson(payload)


@router.put("/lesson/{lesson_id}/progress", response_model=UserLessonResponse)
def update_lesson_progress(
    lesson_id: UUID,
    payload: UserLessonUpdate,
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> UserLessonResponse:
    lesson = service.update_progress(user_id, lesson_id, payload)
    if lesson is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User lesson not found")
    return lesson


@router.post("/lesson/{lesson_id}/complete", response_model=UserLessonResponse)
def complete_lesson(
    lesson_id: UUID,
    payload: UserLessonCompletionRequest,
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> UserLessonResponse:
    lesson = service.complete_lesson(user_id, lesson_id, payload)
    if lesson is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User lesson not found")
    return lesson


@router.post("/lesson/{lesson_id}/abandon", response_model=UserLessonResponse)
def abandon_lesson(
    lesson_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> UserLessonResponse:
    lesson = service.abandon_lesson(user_id, lesson_id)
    if lesson is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User lesson not found")
    return lesson


@router.get("/in-progress", response_model=List[UserLessonResponse])
def list_in_progress_lessons(
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> List[UserLessonResponse]:
    return service.get_in_progress_lessons(user_id)


@router.get("/completed", response_model=List[UserLessonResponse])
def list_completed_lessons(
    limit: int = Query(default=50, ge=1, le=200),
    offset: int = Query(default=0, ge=0),
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> List[UserLessonResponse]:
    return service.get_completed_lessons(user_id, limit=limit, offset=offset)


@router.get("/stats", response_model=UserLessonStats)
def get_user_lesson_stats(
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> UserLessonStats:
    return service.get_lesson_stats(user_id)


@router.delete("/lesson/{lesson_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user_lesson(
    lesson_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    service: UserLessonService = Depends(get_user_lesson_service),
) -> Response:
    deleted = service.delete_user_lesson(user_id, lesson_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User lesson not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
