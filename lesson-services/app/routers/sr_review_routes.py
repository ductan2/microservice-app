from datetime import date
from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.progress_schema import (
    SRReviewCreate,
    SRReviewResponse,
    SRReviewStatsResponse,
    SRReviewTodayStatsResponse,
)
from app.services.sr_review_service import SRReviewService

router = APIRouter(prefix="/api/spaced-repetition/reviews", tags=["Spaced Repetition Reviews"])


def _get_service(db: Session) -> SRReviewService:
    return SRReviewService(db)


@router.post("", response_model=SRReviewResponse, status_code=status.HTTP_201_CREATED)
def create_review(payload: SRReviewCreate, db: Session = Depends(get_db)) -> SRReviewResponse:
    service = _get_service(db)
    try:
        review = service.create_review(payload)
    except ValueError as exc:  # pragma: no cover - defensive
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(exc))
    return SRReviewResponse.model_validate(review, from_attributes=True)


@router.get("/user/{user_id}", response_model=List[SRReviewResponse])
def get_user_reviews(
    user_id: UUID,
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    date_from: Optional[date] = Query(None),
    date_to: Optional[date] = Query(None),
    db: Session = Depends(get_db),
) -> List[SRReviewResponse]:
    service = _get_service(db)
    reviews = service.get_user_reviews(
        user_id,
        limit=limit,
        offset=offset,
        date_from=date_from,
        date_to=date_to,
    )
    return [SRReviewResponse.model_validate(review, from_attributes=True) for review in reviews]


@router.get("/user/{user_id}/flashcard/{flashcard_id}", response_model=List[SRReviewResponse])
def get_flashcard_reviews(
    user_id: UUID,
    flashcard_id: UUID,
    db: Session = Depends(get_db),
) -> List[SRReviewResponse]:
    service = _get_service(db)
    reviews = service.get_flashcard_reviews(user_id, flashcard_id)
    return [SRReviewResponse.model_validate(review, from_attributes=True) for review in reviews]


@router.get("/user/{user_id}/today", response_model=SRReviewTodayStatsResponse)
def get_today_review_stats(user_id: UUID, db: Session = Depends(get_db)) -> SRReviewTodayStatsResponse:
    service = _get_service(db)
    stats = service.get_today_stats(user_id)
    return SRReviewTodayStatsResponse(**stats)


@router.get("/user/{user_id}/stats", response_model=SRReviewStatsResponse)
def get_user_review_stats(user_id: UUID, db: Session = Depends(get_db)) -> SRReviewStatsResponse:
    service = _get_service(db)
    stats = service.get_user_review_stats(user_id)
    return SRReviewStatsResponse(**stats)


@router.delete("/{review_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_review(review_id: UUID, db: Session = Depends(get_db)) -> Response:
    service = _get_service(db)
    deleted = service.delete_review(review_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="SR review not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
