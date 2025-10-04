from typing import Optional
from pydantic import BaseModel, Field
from datetime import datetime
from uuid import UUID


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


class PointsAdjustmentRequest(BaseModel):
    points: int = Field(..., ge=1)


class UserPointsRankResponse(BaseModel):
    lifetime_rank: Optional[int] = None
    weekly_rank: Optional[int] = None
    monthly_rank: Optional[int] = None


class PointsLeaderboardEntry(BaseModel):
    rank: int
    user_id: UUID
    points: int
