from __future__ import annotations

from typing import List
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.course_lesson_schema import (
    CourseLessonCreate,
    CourseLessonResponse,
    CourseLessonUpdate,
)
from app.services.course_lesson_service import CourseLessonService
from app.routers.base import ApiResponseRoute


router = APIRouter(
    prefix="/api/course-lessons",
    tags=["Course Lessons"],
    route_class=ApiResponseRoute,
)


def get_service(db: Session = Depends(get_db)) -> CourseLessonService:
    return CourseLessonService(db)


@router.get("/by-course/{course_id}", response_model=List[CourseLessonResponse])
def list_course_lessons(
    course_id: UUID,
    service: CourseLessonService = Depends(get_service),
) -> List[CourseLessonResponse]:
    return service.list_by_course(course_id)


@router.post("", response_model=CourseLessonResponse, status_code=status.HTTP_201_CREATED)
def create_course_lesson(
    payload: CourseLessonCreate,
    service: CourseLessonService = Depends(get_service),
) -> CourseLessonResponse:
    return service.create(payload)


@router.put("/{row_id}", response_model=CourseLessonResponse)
def update_course_lesson(
    row_id: UUID,
    payload: CourseLessonUpdate,
    service: CourseLessonService = Depends(get_service),
) -> CourseLessonResponse:
    row = service.update(row_id, payload)
    if row is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Course lesson not found")
    return row


@router.delete("/{row_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_course_lesson(
    row_id: UUID,
    service: CourseLessonService = Depends(get_service),
):
    deleted = service.delete(row_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Course lesson not found")
    return None

