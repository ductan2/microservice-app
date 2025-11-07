from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Optional
from uuid import UUID

from pydantic import BaseModel


class EnrollmentStatus(str, Enum):
    ENROLLED = "enrolled"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class CourseEnrollmentBase(BaseModel):
    user_id: UUID
    course_id: UUID
    status: EnrollmentStatus = EnrollmentStatus.ENROLLED
    progress_percent: int = 0
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    last_accessed_at: Optional[datetime] = None


class CourseEnrollmentCreate(BaseModel):
    course_id: UUID


class CourseEnrollmentUpdate(BaseModel):
    status: Optional[EnrollmentStatus] = None
    progress_percent: Optional[int] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    last_accessed_at: Optional[datetime] = None


class CourseEnrollmentResponse(CourseEnrollmentBase):
    id: UUID
    enrolled_at: datetime

    class Config:
        from_attributes = True


