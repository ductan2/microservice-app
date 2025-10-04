from typing import Dict, List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.progress_schema import (
    LeaderboardPeriod,
    LeaderboardResponse,
    LeaderboardSnapshotCreate,
)
from app.services.leaderboard_service import LeaderboardService


router = APIRouter(prefix="/api/leaderboards", tags=["Leaderboards"])


def _get_service(db: Session) -> LeaderboardService:
    return LeaderboardService(db)


@router.get("/weekly/current", response_model=LeaderboardResponse)
def get_current_weekly_leaderboard(
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> LeaderboardResponse:
    service = _get_service(db)
    leaderboard = service.get_current_weekly_leaderboard(limit=limit, offset=offset)
    if not leaderboard:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Leaderboard not found"
        )
    return leaderboard


@router.get("/monthly/current", response_model=LeaderboardResponse)
def get_current_monthly_leaderboard(
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> LeaderboardResponse:
    service = _get_service(db)
    leaderboard = service.get_current_monthly_leaderboard(limit=limit, offset=offset)
    if not leaderboard:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Leaderboard not found"
        )
    return leaderboard


@router.get("/weekly/history", response_model=List[LeaderboardResponse])
def get_weekly_history(
    limit: int = Query(10, ge=1, le=52),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> List[LeaderboardResponse]:
    service = _get_service(db)
    return service.get_weekly_history(limit=limit, offset=offset)


@router.get("/monthly/history", response_model=List[LeaderboardResponse])
def get_monthly_history(
    limit: int = Query(12, ge=1, le=60),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> List[LeaderboardResponse]:
    service = _get_service(db)
    return service.get_monthly_history(limit=limit, offset=offset)


@router.post("/snapshot/weekly", status_code=status.HTTP_201_CREATED)
def create_weekly_snapshot(
    payload: LeaderboardSnapshotCreate,
    db: Session = Depends(get_db),
) -> Dict[str, int]:
    service = _get_service(db)
    created = service.create_snapshot(LeaderboardPeriod.WEEKLY, payload)
    return {"created": created}


@router.post("/snapshot/monthly", status_code=status.HTTP_201_CREATED)
def create_monthly_snapshot(
    payload: LeaderboardSnapshotCreate,
    db: Session = Depends(get_db),
) -> Dict[str, int]:
    service = _get_service(db)
    created = service.create_snapshot(LeaderboardPeriod.MONTHLY, payload)
    return {"created": created}


@router.get("/user/{user_id}/history", response_model=Dict[str, List[LeaderboardResponse]])
def get_user_history(
    user_id: UUID,
    db: Session = Depends(get_db),
) -> Dict[str, List[LeaderboardResponse]]:
    service = _get_service(db)
    return service.get_user_leaderboard_history(user_id)


@router.get("/week/{week_key}", response_model=LeaderboardResponse)
def get_week_leaderboard(
    week_key: str,
    limit: Optional[int] = Query(None, ge=1, le=500),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> LeaderboardResponse:
    service = _get_service(db)
    leaderboard = service.get_leaderboard_by_week(week_key, limit=limit, offset=offset)
    if not leaderboard:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Leaderboard not found"
        )
    return leaderboard


@router.get("/month/{month_key}", response_model=LeaderboardResponse)
def get_month_leaderboard(
    month_key: str,
    limit: Optional[int] = Query(None, ge=1, le=500),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> LeaderboardResponse:
    service = _get_service(db)
    leaderboard = service.get_leaderboard_by_month(month_key, limit=limit, offset=offset)
    if not leaderboard:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND, detail="Leaderboard not found"
        )
    return leaderboard
