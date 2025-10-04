from datetime import date
from typing import Dict, List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.progress_schema import (
    ProgressEventCreate,
    ProgressEventResponse,
)
from app.services.progress_event_service import ProgressEventService


router = APIRouter(prefix="/api/progress-events", tags=["Progress Events"])


def _get_service(db: Session) -> ProgressEventService:
    return ProgressEventService(db)


@router.get("/user/{user_id}", response_model=List[ProgressEventResponse])
def get_user_events(
    user_id: UUID,
    event_type: Optional[str] = Query(None, alias="type"),
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    date_from: Optional[date] = Query(None),
    date_to: Optional[date] = Query(None),
    db: Session = Depends(get_db),
) -> List[ProgressEventResponse]:
    if date_from and date_to and date_from > date_to:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="date_from must be before or equal to date_to",
        )

    service = _get_service(db)
    return service.get_user_events(
        user_id=user_id,
        event_type=event_type,
        limit=limit,
        offset=offset,
        date_from=date_from,
        date_to=date_to,
    )


@router.get("/{event_id}", response_model=ProgressEventResponse)
def get_event(event_id: int, db: Session = Depends(get_db)) -> ProgressEventResponse:
    service = _get_service(db)
    event = service.get_event(event_id)
    if not event:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Event not found")
    return event


@router.post("", response_model=ProgressEventResponse, status_code=status.HTTP_201_CREATED)
def create_event(
    payload: ProgressEventCreate,
    db: Session = Depends(get_db),
) -> ProgressEventResponse:
    service = _get_service(db)
    return service.create_event(payload)


@router.get("/user/{user_id}/type/{event_type}", response_model=List[ProgressEventResponse])
def get_events_by_type(
    user_id: UUID,
    event_type: str,
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    db: Session = Depends(get_db),
) -> List[ProgressEventResponse]:
    service = _get_service(db)
    return service.get_events_by_type(
        user_id=user_id,
        event_type=event_type,
        limit=limit,
        offset=offset,
    )


@router.get("/user/{user_id}/recent", response_model=List[ProgressEventResponse])
def get_recent_events(
    user_id: UUID,
    limit: int = Query(50, ge=1, le=200),
    db: Session = Depends(get_db),
) -> List[ProgressEventResponse]:
    service = _get_service(db)
    return service.get_recent_events(user_id=user_id, limit=limit)


@router.delete("/{event_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_event(event_id: int, db: Session = Depends(get_db)) -> Response:
    service = _get_service(db)
    deleted = service.delete_event(event_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Event not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@router.get("/stats/types", response_model=Dict[str, int])
def get_event_type_stats(
    date_from: Optional[date] = Query(None),
    date_to: Optional[date] = Query(None),
    db: Session = Depends(get_db),
) -> Dict[str, int]:
    if date_from and date_to and date_from > date_to:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="date_from must be before or equal to date_to",
        )

    service = _get_service(db)
    return service.get_event_type_stats(date_from=date_from, date_to=date_to)

