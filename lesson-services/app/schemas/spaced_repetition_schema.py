from pydantic import BaseModel, Field
from typing import Dict, Optional
from datetime import datetime, date
from uuid import UUID


# Spaced Repetition Card Schemas
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


class SRCardStatsResponse(BaseModel):
    total_cards: int
    due_cards: int
    suspended_cards: int
    new_cards: int
    learning_cards: int
    mature_cards: int
    average_ease_factor: float
    average_interval: float


# Spaced Repetition Review Schemas
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


class SRReviewTodayStatsResponse(BaseModel):
    total_reviews: int
    average_quality: float
    quality_distribution: Dict[int, int]
    retention_rate: float


class SRReviewStatsResponse(BaseModel):
    total_reviews: int
    average_quality: float
    quality_distribution: Dict[int, int]
    retention_rate: float
    review_streak: int
    unique_flashcards: int
    busiest_day: Optional[date]
    busiest_day_count: int
    total_time_minutes: int
