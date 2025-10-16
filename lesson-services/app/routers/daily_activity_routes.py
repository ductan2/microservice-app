from datetime import date, timedelta
from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.daily_activity_schema import (
    DailyActivityIncrementRequest,
    DailyActivityMonthSummary,
    DailyActivityResponse,
    DailyActivitySummary,
    DailyTotals,
)
from app.services.daily_activity_service import DailyActivityService
from app.middlewares.auth_middleware import get_current_user_id


router = APIRouter(prefix="/api/daily-activity", tags=["Daily Activity"])


def get_daily_activity_service(db: Session = Depends(get_db)) -> DailyActivityService:
    """Dependency to get DailyActivityService instance."""
    return DailyActivityService(db)


def _empty_activity(user_id: UUID, activity_dt: date) -> DailyActivityResponse:
    return DailyActivityResponse(
        user_id=user_id,
        activity_dt=activity_dt,
        lessons_completed=0,
        quizzes_completed=0,
        minutes=0,
        points=0,
    )


@router.get("/user/me/today", response_model=DailyActivityResponse)
def get_today_activity(
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service)
) -> DailyActivityResponse:
    activity = service.get_today_activity(user_id)
    if activity is None:
        return _empty_activity(user_id, date.today())
    return DailyActivityResponse.model_validate(activity)


@router.get(
    "/user/me/date/{activity_date}",
    response_model=DailyActivityResponse
)
def get_activity_by_date(
    activity_date: date,
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service)
) -> DailyActivityResponse:
    activity = service.get_activity_by_date(user_id, activity_date)
    if activity is None:
        return _empty_activity(user_id, activity_date)
    return DailyActivityResponse.model_validate(activity)


@router.get("/user/me/range", response_model=List[DailyActivityResponse])
def get_activity_range(
    date_from: Optional[date] = Query(default=None),
    date_to: Optional[date] = Query(default=None),
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service),
) -> List[DailyActivityResponse]:
    end_date = date_to or date.today()
    start_date = date_from or (end_date - timedelta(days=29))
    if start_date > end_date:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="date_from must be before or equal to date_to",
        )
    activities = service.get_activity_range(user_id, start_date, end_date)
    return [DailyActivityResponse.model_validate(activity) for activity in activities]


@router.get("/user/me/week", response_model=List[DailyActivityResponse])
def get_week_activity(
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service)
) -> List[DailyActivityResponse]:
    activities = service.get_week_activity(user_id)
    return [DailyActivityResponse.model_validate(activity) for activity in activities]


@router.get("/user/me/month", response_model=DailyActivityMonthSummary)
def get_month_activity(
    year: Optional[int] = Query(default=None),
    month: Optional[int] = Query(default=None, ge=1, le=12),
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service),
) -> DailyActivityMonthSummary:
    today = date.today()
    target_year = year or today.year
    target_month = month or today.month
    summary = service.get_month_activity(user_id, target_year, target_month)
    return DailyActivityMonthSummary(
        year=summary["year"],
        month=summary["month"],
        totals=DailyTotals(**summary["totals"]),
        days=[
            DailyActivityResponse.model_validate(activity)
            for activity in summary["days"]
        ],
    )


@router.get("/user/me/stats/summary", response_model=DailyActivitySummary)
def get_activity_summary(
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service)
) -> DailyActivitySummary:
    summary = service.get_activity_summary(user_id)
    return DailyActivitySummary(
        lifetime=DailyTotals(**summary["lifetime"]),
        last_7_days=DailyTotals(**summary["last_7_days"]),
        last_30_days=DailyTotals(**summary["last_30_days"]),
        average_per_day=DailyTotals(**summary["average_per_day"]),
        total_active_days=summary["total_active_days"],
        most_active_day=(
            DailyActivityResponse.model_validate(summary["most_active_day"])
            if summary["most_active_day"]
            else None
        ),
    )


@router.post("/increment", response_model=DailyActivityResponse)
def increment_activity(
    payload: DailyActivityIncrementRequest, 
    user_id: UUID = Depends(get_current_user_id),
    service: DailyActivityService = Depends(get_daily_activity_service)
) -> DailyActivityResponse:
    try:
        activity = service.increment_activity(
            user_id=user_id,
            activity_date=payload.activity_dt or date.today(),
            field=payload.field.lower(),
            amount=payload.amount,
        )
    except ValueError as exc:  # pragma: no cover - defensive, ensures 400 for invalid field
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc)
        ) from exc

    return DailyActivityResponse.model_validate(activity)

