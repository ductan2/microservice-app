from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Request, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.spaced_repetition_schema import SRCardCreate, SRCardResponse, SRCardStatsResponse
from app.services.sr_card_service import SRCardService

router = APIRouter(prefix="/api/spaced-repetition/cards", tags=["Spaced Repetition Cards"])


def get_sr_card_service(db: Session = Depends(get_db)) -> SRCardService:
    """Dependency to get SRCardService instance."""
    return SRCardService(db)


@router.get("/user/me", response_model=List[SRCardResponse])
def get_user_cards(
    request: Request,
    suspended: Optional[bool] = Query(None),
    due_only: bool = Query(False),
    service: SRCardService = Depends(get_sr_card_service),
) -> List[SRCardResponse]:
    user_id: UUID = request.state.user_id
    cards = service.get_user_cards(user_id, suspended=suspended, due_only=due_only)
    return [SRCardResponse.model_validate(card, from_attributes=True) for card in cards]


@router.get("/user/me/due", response_model=List[SRCardResponse])
def get_due_cards(
    request: Request,
    service: SRCardService = Depends(get_sr_card_service)
) -> List[SRCardResponse]:
    user_id: UUID = request.state.user_id
    cards = service.get_due_cards(user_id)
    return [SRCardResponse.model_validate(card, from_attributes=True) for card in cards]


@router.post("", response_model=SRCardResponse, status_code=status.HTTP_201_CREATED)
def create_card(
    payload: SRCardCreate, 
    service: SRCardService = Depends(get_sr_card_service)
) -> SRCardResponse:
    card = service.create_card(payload)
    return SRCardResponse.model_validate(card, from_attributes=True)


@router.get("/{card_id}", response_model=SRCardResponse)
def get_card(
    card_id: UUID, 
    service: SRCardService = Depends(get_sr_card_service)
) -> SRCardResponse:
    card = service.get_card(card_id)
    if not card:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="SR card not found")
    return SRCardResponse.model_validate(card, from_attributes=True)


@router.patch("/{card_id}/suspend", response_model=SRCardResponse)
def suspend_card(
    card_id: UUID, 
    service: SRCardService = Depends(get_sr_card_service)
) -> SRCardResponse:
    card = service.suspend_card(card_id)
    if not card:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="SR card not found")
    return SRCardResponse.model_validate(card, from_attributes=True)


@router.patch("/{card_id}/unsuspend", response_model=SRCardResponse)
def unsuspend_card(
    card_id: UUID, 
    service: SRCardService = Depends(get_sr_card_service)
) -> SRCardResponse:
    card = service.unsuspend_card(card_id)
    if not card:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="SR card not found")
    return SRCardResponse.model_validate(card, from_attributes=True)


@router.delete("/{card_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_card(
    card_id: UUID, 
    service: SRCardService = Depends(get_sr_card_service)
) -> Response:
    deleted = service.delete_card(card_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="SR card not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@router.get("/user/me/stats", response_model=SRCardStatsResponse)
def get_user_card_stats(
    request: Request,
    service: SRCardService = Depends(get_sr_card_service)
) -> SRCardStatsResponse:
    user_id: UUID = request.state.user_id
    stats = service.get_user_stats(user_id)
    return SRCardStatsResponse(**stats)
