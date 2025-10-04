from pydantic import BaseModel
from typing import List, Optional
from datetime import datetime
from uuid import UUID
from enum import Enum


class LeaderboardPeriod(str, Enum):
    WEEKLY = "weekly"
    MONTHLY = "monthly"


# Leaderboard Schemas
class LeaderboardEntry(BaseModel):
    rank: int
    user_id: UUID
    points: int


class LeaderboardEntryCreate(LeaderboardEntry):
    pass


class LeaderboardResponse(BaseModel):
    period: LeaderboardPeriod
    period_key: str
    entries: List[LeaderboardEntry]
    taken_at: datetime


class LeaderboardSnapshotCreate(BaseModel):
    period_key: str
    entries: List[LeaderboardEntryCreate]
    taken_at: Optional[datetime] = None
