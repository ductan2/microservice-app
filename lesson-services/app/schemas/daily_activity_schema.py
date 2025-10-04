from datetime import date
from typing import List, Optional
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class DailyActivityBase(BaseModel):
    user_id: UUID
    activity_dt: date
    lessons_completed: int = 0
    quizzes_completed: int = 0
    minutes: int = 0
    points: int = 0

    model_config = ConfigDict(from_attributes=True)


class DailyActivityResponse(DailyActivityBase):
    pass


class DailyActivityRangeRequest(BaseModel):
    date_from: Optional[date] = None
    date_to: Optional[date] = None


class DailyActivityIncrementRequest(BaseModel):
    user_id: UUID
    activity_dt: date = Field(alias="activityDate")
    field: str
    amount: int = Field(ge=1)

    model_config = ConfigDict(populate_by_name=True)


class DailyTotals(BaseModel):
    lessons_completed: int
    quizzes_completed: int
    minutes: int
    points: int


class DailyActivitySummary(BaseModel):
    lifetime: DailyTotals
    last_7_days: DailyTotals
    last_30_days: DailyTotals
    average_per_day: DailyTotals
    total_active_days: int
    most_active_day: Optional[DailyActivityResponse] = None


class DailyActivityMonthSummary(BaseModel):
    year: int
    month: int
    totals: DailyTotals
    days: List[DailyActivityResponse]
