from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime, date
from uuid import UUID
from enum import Enum

class LessonStatus(str, Enum):
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    ABANDONED = "abandoned"

class LeaderboardPeriod(str, Enum):
    WEEKLY = "weekly"
    MONTHLY = "monthly"

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

# Quiz Answer Schemas
class QuizAnswerBase(BaseModel):
    question_id: UUID
    selected_ids: List[UUID] = []
    text_answer: Optional[str] = None
    is_correct: Optional[bool] = None
    points_earned: int = 0

class QuizAnswerCreate(QuizAnswerBase):
    attempt_id: UUID

class QuizAnswerResponse(QuizAnswerBase):
    id: UUID
    attempt_id: UUID
    answered_at: datetime
    
    class Config:
        from_attributes = True

# Spaced Repetition Schemas
class SRCardBase(BaseModel):
    user_id: UUID
    flashcard_id: UUID
    ease_factor: float = 2.5
    interval_d: int = 0
    repetition: int = 0
    suspended: bool = False

class SRCardCreate(SRCardBase):
    pass

class SRCardUpdate(BaseModel):
    ease_factor: Optional[float] = None
    interval_d: Optional[int] = None
    repetition: Optional[int] = None
    due_at: Optional[datetime] = None
    suspended: Optional[bool] = None

class SRCardResponse(SRCardBase):
    id: UUID
    due_at: datetime
    
    class Config:
        from_attributes = True

class SRReviewBase(BaseModel):
    user_id: UUID
    flashcard_id: UUID
    quality: int = Field(..., ge=0, le=5)
    prev_interval: Optional[int] = None
    new_interval: Optional[int] = None
    new_ef: Optional[float] = None

class SRReviewCreate(SRReviewBase):
    pass

class SRReviewResponse(SRReviewBase):
    id: UUID
    reviewed_at: datetime
    
    class Config:
        from_attributes = True

# Daily Activity Schemas
class DailyActivityBase(BaseModel):
    user_id: UUID
    activity_dt: date
    lessons_completed: int = 0
    quizzes_completed: int = 0
    minutes: int = 0
    points: int = 0

class DailyActivityCreate(DailyActivityBase):
    pass

class DailyActivityUpdate(BaseModel):
    lessons_completed: Optional[int] = None
    quizzes_completed: Optional[int] = None
    minutes: Optional[int] = None
    points: Optional[int] = None

class DailyActivityResponse(DailyActivityBase):
    class Config:
        from_attributes = True

# User Streak Schemas
class UserStreakBase(BaseModel):
    user_id: UUID
    current_len: int = 0
    longest_len: int = 0
    last_day: Optional[date] = None

class UserStreakCreate(UserStreakBase):
    pass

class UserStreakUpdate(BaseModel):
    current_len: Optional[int] = None
    longest_len: Optional[int] = None
    last_day: Optional[date] = None

class UserStreakResponse(UserStreakBase):
    class Config:
        from_attributes = True

# User Points Schemas
class UserPointsBase(BaseModel):
    user_id: UUID
    lifetime: int = 0
    weekly: int = 0
    monthly: int = 0

class UserPointsCreate(UserPointsBase):
    pass

class UserPointsUpdate(BaseModel):
    lifetime: Optional[int] = None
    weekly: Optional[int] = None
    monthly: Optional[int] = None

class UserPointsResponse(UserPointsBase):
    updated_at: datetime
    
    class Config:
        from_attributes = True

# Leaderboard Schemas
class LeaderboardEntry(BaseModel):
    rank: int
    user_id: UUID
    points: int

class LeaderboardResponse(BaseModel):
    period: LeaderboardPeriod
    period_key: str
    entries: List[LeaderboardEntry]
    taken_at: datetime

# Progress Event Schemas
class ProgressEventBase(BaseModel):
    user_id: UUID
    type: str
    payload: dict

class ProgressEventCreate(ProgressEventBase):
    pass

class ProgressEventResponse(ProgressEventBase):
    id: int
    created_at: datetime
    
    class Config:
        from_attributes = True