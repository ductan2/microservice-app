from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime
from uuid import UUID


# Quiz Attempt Schemas
class QuizAttemptBase(BaseModel):
    user_id: UUID
    quiz_id: UUID
    lesson_id: Optional[UUID] = None
    duration_ms: Optional[int] = None
    total_points: int = 0
    max_points: int = 0
    passed: Optional[bool] = None
    attempt_no: int = 1


class QuizAttemptCreate(QuizAttemptBase):
    pass


class QuizAttemptUpdate(BaseModel):
    submitted_at: Optional[datetime] = None
    duration_ms: Optional[int] = None
    total_points: Optional[int] = None
    passed: Optional[bool] = None


class QuizAttemptResponse(QuizAttemptBase):
    id: UUID
    started_at: datetime
    submitted_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class QuizAttemptDetailResponse(QuizAttemptResponse):
    answers: List["QuizAnswerResponse"] = Field(default_factory=list)

    class Config:
        from_attributes = True


# Quiz Answer Schemas
class QuizAnswerBase(BaseModel):
    question_id: UUID
    selected_ids: List[UUID] = Field(default_factory=list)
    text_answer: Optional[str] = None
    is_correct: Optional[bool] = None
    points_earned: int = 0


class QuizAnswerCreate(QuizAnswerBase):
    attempt_id: UUID


class QuizAnswerSubmission(QuizAnswerBase):
    pass


class QuizAnswerUpdate(BaseModel):
    selected_ids: Optional[List[UUID]] = None
    text_answer: Optional[str] = None
    is_correct: Optional[bool] = None
    points_earned: Optional[int] = None


class QuizAnswerResponse(QuizAnswerBase):
    id: UUID
    attempt_id: UUID
    answered_at: datetime

    class Config:
        from_attributes = True


class QuizAnswerSummary(BaseModel):
    total_answers: int
    correct_answers: int
    accuracy: float
    points_earned: int


class QuizAttemptSubmit(BaseModel):
    total_points: int
    max_points: Optional[int] = None
    passed: Optional[bool] = None
    submitted_at: Optional[datetime] = None
    duration_ms: Optional[int] = None
    answers: Optional[List[QuizAnswerSubmission]] = None


# Update forward references
QuizAttemptDetailResponse.model_rebuild()
