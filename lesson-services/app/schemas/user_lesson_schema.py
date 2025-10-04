from pydantic import BaseModel
from typing import Optional
from datetime import datetime
from uuid import UUID
from enum import Enum


class LessonStatus(str, Enum):
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    ABANDONED = "abandoned"


# User Lesson Schemas
class UserLessonBase(BaseModel):
    user_id: UUID
    lesson_id: UUID
    status: LessonStatus = LessonStatus.IN_PROGRESS
    last_section_ord: Optional[int] = None
    score_total: int = 0


class UserLessonCreate(UserLessonBase):
    pass


class UserLessonUpdate(BaseModel):
    status: Optional[LessonStatus] = None
    last_section_ord: Optional[int] = None
    score_total: Optional[int] = None
    completed_at: Optional[datetime] = None


class UserLessonResponse(UserLessonBase):
    id: UUID
    started_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class UserLessonCompletionRequest(BaseModel):
    score_total: Optional[int] = None
    last_section_ord: Optional[int] = None
    completed_at: Optional[datetime] = None


class UserLessonStats(BaseModel):
    total_started: int
    in_progress: int
    completed: int
    abandoned: int
    completion_rate: float
    total_score: int
