from __future__ import annotations

from typing import List, Optional
from uuid import UUID

from sqlalchemy import and_, asc
from sqlalchemy.orm import Session

from app.models.progress_models import CourseLesson
from app.schemas.course_lesson_schema import (
    CourseLessonCreate,
    CourseLessonResponse,
    CourseLessonUpdate,
)


class CourseLessonService:
    def __init__(self, db: Session):
        self.db = db

    def list_by_course(self, course_id: UUID) -> List[CourseLessonResponse]:
        rows = (
            self.db.query(CourseLesson)
            .filter(CourseLesson.course_id == course_id)
            .order_by(asc(CourseLesson.ord), asc(CourseLesson.created_at))
            .all()
        )
        return [CourseLessonResponse.from_orm(r) for r in rows]

    def create(self, payload: CourseLessonCreate) -> CourseLessonResponse:
        row = CourseLesson(**payload.dict())
        self.db.add(row)
        self.db.commit()
        self.db.refresh(row)
        return CourseLessonResponse.from_orm(row)

    def update(self, row_id: UUID, payload: CourseLessonUpdate) -> Optional[CourseLessonResponse]:
        row = self.db.query(CourseLesson).filter(CourseLesson.id == row_id).one_or_none()
        if not row:
            return None
        for field, value in payload.dict(exclude_unset=True).items():
            setattr(row, field, value)
        self.db.commit()
        self.db.refresh(row)
        return CourseLessonResponse.from_orm(row)

    def delete(self, row_id: UUID) -> bool:
        row = self.db.query(CourseLesson).filter(CourseLesson.id == row_id).one_or_none()
        if not row:
            return False
        self.db.delete(row)
        self.db.commit()
        return True


