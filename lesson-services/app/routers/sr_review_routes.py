from datetime import date
from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.spaced_repetition_schema import (
    SRReviewCreate,
    SRReviewResponse,
    SRReviewStatsResponse,
    SRReviewTodayStatsResponse,
)
from app.services.sr_review_service import SRReviewService
from app.middlewares.auth_middleware import get_current_user_id

router = APIRouter(prefix="/api/spaced-repetition/reviews", tags=["Spaced Repetition Reviews"])


def get_sr_review_service(db: Session = Depends(get_db)) -> SRReviewService:
    """Dependency to get SRReviewService instance."""
    return SRReviewService(db)


@router.post("", response_model=SRReviewResponse, status_code=status.HTTP_201_CREATED)
def create_review(
    payload: SRReviewCreate, 
    user_id: UUID = Depends(get_current_user_id),
    service: SRReviewService = Depends(get_sr_review_service)
) -> SRReviewResponse:
    try:
        review = service.create_review(user_id, payload)
    except ValueError as exc:  # pragma: no cover - defensive
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc))
    return SRReviewResponse.model_validate(review, from_attributes=True)


@router.get("/user/me", response_model=List[SRReviewResponse])
def get_user_reviews(
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    date_from: Optional[date] = Query(None),
    date_to: Optional[date] = Query(None),
    user_id: UUID = Depends(get_current_user_id),
    service: SRReviewService = Depends(get_sr_review_service),
) -> List[SRReviewResponse]:
    reviews = service.get_user_reviews(
        user_id,
        limit=limit,
        offset=offset,
        date_from=date_from,
        date_to=date_to,
    )
    return [SRReviewResponse.model_validate(review, from_attributes=True) for review in reviews]


@router.get("/user/me/flashcard/{flashcard_id}", response_model=List[SRReviewResponse])
def get_flashcard_reviews(
    flashcard_id: UUID,
    user_id: UUID = Depends(get_current_user_id),
    service: SRReviewService = Depends(get_sr_review_service),
) -> List[SRReviewResponse]:
    reviews = service.get_flashcard_reviews(user_id, flashcard_id)
    return [SRReviewResponse.model_validate(review, from_attributes=True) for review in reviews]


@router.get("/user/me/today", response_model=SRReviewTodayStatsResponse)
def get_today_review_stats(
    user_id: UUID = Depends(get_current_user_id),
    service: SRReviewService = Depends(get_sr_review_service)
) -> SRReviewTodayStatsResponse:
    stats = service.get_today_stats(user_id)
    return SRReviewTodayStatsResponse(**stats)


@router.get("/user/me/stats", response_model=SRReviewStatsResponse)
def get_user_review_stats(
    user_id: UUID = Depends(get_current_user_id),
    service: SRReviewService = Depends(get_sr_review_service)
) -> SRReviewStatsResponse:
    stats = service.get_user_review_stats(user_id)
    return SRReviewStatsResponse(**stats)


@router.delete("/{review_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_review(
    review_id: UUID, 
    service: SRReviewService = Depends(get_sr_review_service)
) -> Response:
    deleted = service.delete_review(review_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="SR review not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
