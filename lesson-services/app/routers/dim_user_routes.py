from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.dim_user_schema import (
    DimUserCreate,
    DimUserLocaleUpdate,
    DimUserResponse,
    DimUserUpdate,
)
from app.services.dim_user_service import DimUserService


router = APIRouter(prefix="/api/users", tags=["User Preferences"])


def get_dim_user_service(db: Session = Depends(get_db)) -> DimUserService:
    """Dependency to get DimUserService instance."""
    return DimUserService(db)


@router.get("/{user_id}", response_model=DimUserResponse)
def get_user_preferences(
    user_id: UUID, 
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    user = service.get_user_by_id(user_id)
    if not user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return user


@router.post("", response_model=DimUserResponse, status_code=status.HTTP_201_CREATED)
def create_user_preferences(
    payload: DimUserCreate, 
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    if service.user_exists(payload.user_id):
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="User preferences already exist",
        )
    return service.create_user(payload)


@router.put("/{user_id}", response_model=DimUserResponse)
def update_user_preferences(
    user_id: UUID, 
    payload: DimUserUpdate, 
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    user = service.update_user(user_id, payload)
    if not user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return user


@router.patch("/{user_id}/locale", response_model=DimUserResponse)
def update_user_locale(
    user_id: UUID, 
    payload: DimUserLocaleUpdate, 
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    user = service.update_locale(user_id, payload.locale)
    if not user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return user


@router.delete("/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_user_preferences(
    user_id: UUID, 
    service: DimUserService = Depends(get_dim_user_service)
) -> Response:
    deleted = service.delete_user(user_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
