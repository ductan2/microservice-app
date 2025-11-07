from __future__ import annotations

from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.middlewares.auth_middleware import get_current_user_id
from app.schemas.course_enrollment_schema import (
    CourseEnrollmentCreate,
    CourseEnrollmentResponse,
    CourseEnrollmentUpdate,
    EnrollmentStatus,
)
from app.services.course_enrollment_service import CourseEnrollmentService
from app.routers.base import ApiResponseRoute


router = APIRouter(
    prefix="/api/course-enrollments",
    tags=["Course Enrollments"],
    route_class=ApiResponseRoute,
)


def get_service(db: Session = Depends(get_db)) -> CourseEnrollmentService:
    return CourseEnrollmentService(db)


@router.get("/me", response_model=List[CourseEnrollmentResponse])
def list_my_enrollments(
    status: Optional[EnrollmentStatus] = Query(default=None),
    limit: int = Query(default=100, ge=1, le=500),
    offset: int = Query(default=0, ge=0),
    user_id: UUID = Depends(get_current_user_id),
    service: CourseEnrollmentService = Depends(get_service),
) -> List[CourseEnrollmentResponse]:
    return service.get_for_user(user_id, status=status, limit=limit, offset=offset)


@router.post("", response_model=CourseEnrollmentResponse, status_code=status.HTTP_201_CREATED)
def enroll_course(
    payload: CourseEnrollmentCreate,
    user_id: UUID = Depends(get_current_user_id),
    service: CourseEnrollmentService = Depends(get_service),
) -> CourseEnrollmentResponse:
    return service.enroll(user_id, payload)


@router.get("/{enrollment_id}", response_model=CourseEnrollmentResponse)
def get_enrollment(
    enrollment_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    service: CourseEnrollmentService = Depends(get_service),
) -> CourseEnrollmentResponse:
    row = service.get_by_id(enrollment_id, user_id=user_id)
    if row is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Enrollment not found")
    return row


@router.put("/{enrollment_id}", response_model=CourseEnrollmentResponse)
def update_enrollment(
    enrollment_id: UUID,
    payload: CourseEnrollmentUpdate,
    user_id: UUID = Depends(get_current_user_id),
    service: CourseEnrollmentService = Depends(get_service),
) -> CourseEnrollmentResponse:
    row = service.update(enrollment_id, user_id, payload)
    if row is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Enrollment not found")
    return row


@router.post("/{enrollment_id}/cancel", response_model=CourseEnrollmentResponse)
def cancel_enrollment(
    enrollment_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    service: CourseEnrollmentService = Depends(get_service),
) -> CourseEnrollmentResponse:
    row = service.cancel(enrollment_id, user_id)
    if row is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Enrollment not found")
    return row

