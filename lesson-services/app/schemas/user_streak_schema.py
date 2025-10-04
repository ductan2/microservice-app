from pydantic import BaseModel
from typing import Optional
from datetime import date
from uuid import UUID


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


class StreakCheckRequest(BaseModel):
    activity_date: Optional[date] = None


class UserStreakStatusResponse(BaseModel):
    user_id: UUID
    current_len: int
    longest_len: int
    last_day: Optional[date] = None
    status: str
    has_activity_today: bool
    days_since_last: Optional[int] = None


class StreakLeaderboardEntry(BaseModel):
    rank: int
    user_id: UUID
    current_len: int
    longest_len: int
    last_day: Optional[date] = None

    class Config:
        from_attributes = True
