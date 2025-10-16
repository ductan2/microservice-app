from typing import Optional
from pydantic import BaseModel, Field
from datetime import date
from uuid import UUID


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


class DailyTotals(BaseModel):
    lessons_completed: int = 0
    quizzes_completed: int = 0
    minutes: int = 0
    points: int = 0


class DailyActivityMonthSummary(BaseModel):
    year: int
    month: int
    totals: DailyTotals
    days: list[DailyActivityResponse]


class DailyActivitySummary(BaseModel):
    lifetime: DailyTotals
    last_7_days: DailyTotals
    last_30_days: DailyTotals
    average_per_day: DailyTotals
    total_active_days: int
    most_active_day: DailyActivityResponse | None = None


class DailyActivityIncrementRequest(BaseModel):
    activity_dt: Optional[date] = None
    field: str  # 'lessons_completed', 'quizzes_completed', 'minutes', 'points'
    amount: int = Field(default=1, ge=1)