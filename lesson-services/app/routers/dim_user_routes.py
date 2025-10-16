from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Request, Response, status
from sqlalchemy.orm import Session

from app.database.connection import get_db
from app.schemas.dim_user_schema import (
    DimUserCreate,
    DimUserLocaleUpdate,
    DimUserResponse,
    DimUserUpdate,
)
from app.services.dim_user_service import DimUserService
from app.middlewares.auth_middleware import get_current_user_id

router = APIRouter(prefix="/api/users", tags=["User Preferences"])


def get_dim_user_service(db: Session = Depends(get_db)) -> DimUserService:
    """Dependency to get DimUserService instance."""
    return DimUserService(db)


@router.get("/me", response_model=DimUserResponse)
def get_user_preferences(
    user_id: UUID = Depends(get_current_user_id),
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    user = service.get_user_by_id(user_id)
    if not user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return user


@router.post("", response_model=DimUserResponse, status_code=status.HTTP_201_CREATED)
def create_user_preferences(
    payload: DimUserCreate,
    user_id: UUID = Depends(get_current_user_id),
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    # Override user_id from payload with authenticated user_id
    payload.user_id = user_id
    if service.user_exists(user_id):
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="User preferences already exist",
        )
    return service.create_user(payload)


@router.put("/me", response_model=DimUserResponse)
def update_user_preferences(
    payload: DimUserUpdate,
    user_id: UUID = Depends(get_current_user_id),
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    user = service.update_user(user_id, payload)
    if not user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return user


@router.patch("/me/locale", response_model=DimUserResponse)
def update_user_locale(
    payload: DimUserLocaleUpdate,
    user_id: UUID = Depends(get_current_user_id),
    service: DimUserService = Depends(get_dim_user_service)
) -> DimUserResponse:
    user = service.update_locale(user_id, payload.locale)
    if not user:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return user


@router.delete("/me", status_code=status.HTTP_204_NO_CONTENT)
def delete_user_preferences(
    user_id: UUID = Depends(get_current_user_id),
    service: DimUserService = Depends(get_dim_user_service)
) -> Response:
    deleted = service.delete_user(user_id)
    if not deleted:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
