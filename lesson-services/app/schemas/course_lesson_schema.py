from __future__ import annotations

from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import BaseModel


class CourseLessonBase(BaseModel):
    course_id: UUID
    lesson_id: UUID
    ord: int
    is_required: bool = True


class CourseLessonCreate(CourseLessonBase):
    pass


class CourseLessonUpdate(BaseModel):
    ord: Optional[int] = None
    is_required: Optional[bool] = None


class CourseLessonResponse(CourseLessonBase):
    id: UUID
    created_at: datetime

    class Config:
        from_attributes = True


