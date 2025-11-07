from __future__ import annotations

from datetime import datetime
from typing import List, Optional
from uuid import UUID

from sqlalchemy import and_, desc
from sqlalchemy.orm import Session

from app.models.progress_models import CourseEnrollment
from app.schemas.course_enrollment_schema import (
    CourseEnrollmentCreate,
    CourseEnrollmentResponse,
    CourseEnrollmentUpdate,
    EnrollmentStatus,
)


class CourseEnrollmentService:
    def __init__(self, db: Session):
        self.db = db

    def get_by_id(self, enrollment_id: UUID, user_id: Optional[UUID] = None) -> Optional[CourseEnrollmentResponse]:
        query = self.db.query(CourseEnrollment).filter(CourseEnrollment.id == enrollment_id)
        if user_id is not None:
            query = query.filter(CourseEnrollment.user_id == user_id)
        row = query.one_or_none()
        return CourseEnrollmentResponse.from_orm(row) if row else None

    def get_for_user(
        self,
        user_id: UUID,
        status: Optional[EnrollmentStatus] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> List[CourseEnrollmentResponse]:
        query = self.db.query(CourseEnrollment).filter(CourseEnrollment.user_id == user_id)
        if status is not None:
            query = query.filter(CourseEnrollment.status == status)
        rows = (
            query.order_by(desc(CourseEnrollment.last_accessed_at), desc(CourseEnrollment.enrolled_at))
            .offset(offset)
            .limit(limit)
            .all()
        )
        return [CourseEnrollmentResponse.from_orm(r) for r in rows]

    def enroll(self, user_id: UUID, payload: CourseEnrollmentCreate) -> CourseEnrollmentResponse:
        existing = (
            self.db.query(CourseEnrollment)
            .filter(
                and_(
                    CourseEnrollment.user_id == user_id,
                    CourseEnrollment.course_id == payload.course_id,
                )
            )
            .one_or_none()
        )
        if existing:
            return CourseEnrollmentResponse.model_validate(existing)

        row = CourseEnrollment(
            user_id=user_id,
            course_id=payload.course_id,
            status=EnrollmentStatus.ENROLLED.value,
            progress_percent=0,
            enrolled_at=datetime.utcnow(),
            last_accessed_at=datetime.utcnow(),
        )
        self.db.add(row)
        self.db.commit()
        self.db.refresh(row)
        return CourseEnrollmentResponse.model_validate(row)

    def update(
        self, enrollment_id: UUID, user_id: UUID, payload: CourseEnrollmentUpdate
    ) -> Optional[CourseEnrollmentResponse]:
        row = (
            self.db.query(CourseEnrollment)
            .filter(
                and_(
                    CourseEnrollment.id == enrollment_id,
                    CourseEnrollment.user_id == user_id,
                )
            )
            .one_or_none()
        )
        if not row:
            return None

        data = payload.dict(exclude_unset=True)
        for field, value in data.items():
            setattr(row, field, value)

        # Auto timestamps maintenance
        row.last_accessed_at = datetime.utcnow()
        if getattr(row, "status", None) == EnrollmentStatus.IN_PROGRESS.value and row.started_at is None:
            row.started_at = datetime.utcnow()
        if getattr(row, "status", None) == EnrollmentStatus.COMPLETED.value and row.completed_at is None:
            row.completed_at = datetime.utcnow()

        self.db.commit()
        self.db.refresh(row)
        return CourseEnrollmentResponse.from_orm(row)

    def cancel(self, enrollment_id: UUID, user_id: UUID) -> Optional[CourseEnrollmentResponse]:
        row = (
            self.db.query(CourseEnrollment)
            .filter(
                and_(
                    CourseEnrollment.id == enrollment_id,
                    CourseEnrollment.user_id == user_id,
                )
            )
            .one_or_none()
        )
        if not row:
            return None
        row.status = EnrollmentStatus.CANCELLED.value
        row.last_accessed_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(row)
        return CourseEnrollmentResponse.from_orm(row)


