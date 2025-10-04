from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.orm import Session
from typing import List
from uuid import UUID

from app.database.connection import get_db
from app.schemas.progress_schema import (
    PointsAdjustmentRequest,
    PointsLeaderboardEntry,
    UserPointsRankResponse,
    UserPointsResponse,
)
from app.services.user_points_service import UserPointsService


router = APIRouter(prefix="/api/points", tags=["User Points"])


def get_user_points_service(db: Session = Depends(get_db)) -> UserPointsService:
    return UserPointsService(db)


@router.get("/user/{user_id}", response_model=UserPointsResponse)
def get_user_points(
    user_id: UUID,
    service: UserPointsService = Depends(get_user_points_service),
) -> UserPointsResponse:
    return service.get_or_create_points(user_id)


@router.post("/user/{user_id}/add", response_model=UserPointsResponse)
def add_user_points(
    user_id: UUID,
    payload: PointsAdjustmentRequest,
    service: UserPointsService = Depends(get_user_points_service),
) -> UserPointsResponse:
    try:
        return service.add_points(user_id, payload.points)
    except ValueError as exc:  # pragma: no cover - defensive validation
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


@router.post("/user/{user_id}/subtract", response_model=UserPointsResponse)
def subtract_user_points(
    user_id: UUID,
    payload: PointsAdjustmentRequest,
    service: UserPointsService = Depends(get_user_points_service),
) -> UserPointsResponse:
    try:
        return service.subtract_points(user_id, payload.points)
    except ValueError as exc:  # pragma: no cover - defensive validation
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)) from exc


def _build_leaderboard(records, attribute: str) -> List[PointsLeaderboardEntry]:
    entries: List[PointsLeaderboardEntry] = []
    for index, record in enumerate(records, start=1):
        entries.append(
            PointsLeaderboardEntry(
                rank=index,
                user_id=record.user_id,
                points=getattr(record, attribute, 0),
            )
        )
    return entries


@router.get("/leaderboard/lifetime", response_model=List[PointsLeaderboardEntry])
def get_lifetime_leaderboard(
    limit: int = Query(default=100, ge=1, le=500),
    offset: int = Query(default=0, ge=0),
    service: UserPointsService = Depends(get_user_points_service),
) -> List[PointsLeaderboardEntry]:
    records = service.get_lifetime_leaderboard(limit=limit, offset=offset)
    return _build_leaderboard(records, "lifetime")


@router.get("/leaderboard/weekly", response_model=List[PointsLeaderboardEntry])
def get_weekly_leaderboard(
    limit: int = Query(default=100, ge=1, le=500),
    offset: int = Query(default=0, ge=0),
    service: UserPointsService = Depends(get_user_points_service),
) -> List[PointsLeaderboardEntry]:
    records = service.get_weekly_leaderboard(limit=limit, offset=offset)
    return _build_leaderboard(records, "weekly")


@router.get("/leaderboard/monthly", response_model=List[PointsLeaderboardEntry])
def get_monthly_leaderboard(
    limit: int = Query(default=100, ge=1, le=500),
    offset: int = Query(default=0, ge=0),
    service: UserPointsService = Depends(get_user_points_service),
) -> List[PointsLeaderboardEntry]:
    records = service.get_monthly_leaderboard(limit=limit, offset=offset)
    return _build_leaderboard(records, "monthly")


@router.post("/reset/weekly")
def reset_weekly_points(
    service: UserPointsService = Depends(get_user_points_service),
) -> dict:
    updated = service.reset_weekly_points()
    return {"updated": updated}


@router.post("/reset/monthly")
def reset_monthly_points(
    service: UserPointsService = Depends(get_user_points_service),
) -> dict:
    updated = service.reset_monthly_points()
    return {"updated": updated}


@router.get("/user/{user_id}/rank", response_model=UserPointsRankResponse)
def get_user_rank(
    user_id: UUID,
    service: UserPointsService = Depends(get_user_points_service),
) -> UserPointsRankResponse:
    ranks = service.get_user_ranks(user_id)
    if ranks is None:
        service.initialize_user_points(user_id)
        ranks = service.get_user_ranks(user_id)
    if ranks is None:  # pragma: no cover - should not happen, but defensive
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User points not found")
    return UserPointsRankResponse(**ranks)
