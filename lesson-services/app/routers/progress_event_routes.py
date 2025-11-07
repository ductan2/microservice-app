from datetime import date
from typing import Dict, List, Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.progress_event_schema import (
    ProgressEventCreate,
    ProgressEventResponse,
)
from app.services.progress_event_service import ProgressEventService
from app.middlewares.auth_middleware import get_current_user_id
from app.routers.base import ApiResponseRoute


router = APIRouter(
    prefix="/api/progress-events",
    tags=["Progress Events"],
    route_class=ApiResponseRoute,
)


def get_progress_event_service(db: Session = Depends(get_db)) -> ProgressEventService:
    """Dependency to get ProgressEventService instance."""
    return ProgressEventService(db)


@router.get("/user/me", response_model=List[ProgressEventResponse])
def get_user_events(
    event_type: Optional[str] = Query(None, alias="type"),
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    date_from: Optional[date] = Query(None),
    date_to: Optional[date] = Query(None),
    user_id: UUID = Depends(get_current_user_id),
    service: ProgressEventService = Depends(get_progress_event_service),
) -> List[ProgressEventResponse]:
    if date_from and date_to and date_from > date_to:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="date_from must be before or equal to date_to",
        )
    return service.get_user_events(
        user_id=user_id,
        event_type=event_type,
        limit=limit,
        offset=offset,
        date_from=date_from,
        date_to=date_to,
    )


@router.get("/{event_id}", response_model=ProgressEventResponse)
def get_event(
    event_id: int, 
    service: ProgressEventService = Depends(get_progress_event_service)
) -> ProgressEventResponse:
    event = service.get_event(event_id)
    if not event:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Event not found")
    return event


@router.post("", response_model=ProgressEventResponse, status_code=status.HTTP_201_CREATED)
def create_event(
    payload: ProgressEventCreate,
    service: ProgressEventService = Depends(get_progress_event_service),
) -> ProgressEventResponse:
    return service.create_event(payload)


@router.get("/user/me/type/{event_type}", response_model=List[ProgressEventResponse])
def get_events_by_type(
    event_type: str,
    limit: int = Query(100, ge=1, le=500),
    offset: int = Query(0, ge=0),
    user_id: UUID = Depends(get_current_user_id),
    service: ProgressEventService = Depends(get_progress_event_service),
) -> List[ProgressEventResponse]:
    return service.get_events_by_type(
        user_id=user_id,
        event_type=event_type,
        limit=limit,
        offset=offset,
    )


@router.get("/user/me/recent", response_model=List[ProgressEventResponse])
def get_recent_events(
    limit: int = Query(50, ge=1, le=200),
    user_id: UUID = Depends(get_current_user_id),
    service: ProgressEventService = Depends(get_progress_event_service),
) -> List[ProgressEventResponse]:
    return service.get_recent_events(user_id=user_id, limit=limit)


@router.delete("/{event_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_event(
    event_id: int, 
    service: ProgressEventService = Depends(get_progress_event_service)
) -> Response:
    deleted = service.delete_event(event_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Event not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@router.get("/stats/types", response_model=Dict[str, int])
def get_event_type_stats(
    date_from: Optional[date] = Query(None),
    date_to: Optional[date] = Query(None),
    service: ProgressEventService = Depends(get_progress_event_service),
) -> Dict[str, int]:
    if date_from and date_to and date_from > date_to:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="date_from must be before or equal to date_to",
        )

    return service.get_event_type_stats(date_from=date_from, date_to=date_to)
