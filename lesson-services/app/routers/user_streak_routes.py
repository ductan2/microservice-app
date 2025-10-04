from __future__ import annotations

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from uuid import UUID

from app.database.connection import get_db
from app.schemas.user_streak_schema import (
    StreakCheckRequest,
    StreakLeaderboardEntry,
    UserStreakResponse,
    UserStreakStatusResponse,
)
from app.services.user_streak_service import UserStreakService


router = APIRouter(prefix="/api/streaks", tags=["User Streaks"])


def get_user_streak_service(db: Session = Depends(get_db)) -> UserStreakService:
    return UserStreakService(db)


@router.get("/user/{user_id}", response_model=UserStreakResponse)
def get_user_streak(
    user_id: UUID,
    service: UserStreakService = Depends(get_user_streak_service),
) -> UserStreakResponse:
    return service.get_or_create_streak(user_id)


@router.post("/user/{user_id}/check", response_model=UserStreakResponse)
def check_user_streak(
    user_id: UUID,
    payload: Optional[StreakCheckRequest] = None,
    service: UserStreakService = Depends(get_user_streak_service),
) -> UserStreakResponse:
    activity_date = payload.activity_date if payload else None
    return service.check_and_update_streak(user_id, activity_date=activity_date)


@router.get("/user/{user_id}/status", response_model=UserStreakStatusResponse)
def get_streak_status(
    user_id: UUID,
    service: UserStreakService = Depends(get_user_streak_service),
) -> UserStreakStatusResponse:
    status_data = service.get_streak_status(user_id)
    return UserStreakStatusResponse(**status_data)


@router.get("/leaderboard", response_model=List[StreakLeaderboardEntry])
def get_streak_leaderboard(
    limit: int = Query(default=50, ge=1, le=200),
    service: UserStreakService = Depends(get_user_streak_service),
) -> List[StreakLeaderboardEntry]:
    records = service.get_streak_leaderboard(limit=limit)
    return [
        StreakLeaderboardEntry(
            rank=index,
            user_id=record.user_id,
            current_len=record.current_len,
            longest_len=record.longest_len,
            last_day=record.last_day,
        )
        for index, record in enumerate(records, start=1)
    ]
