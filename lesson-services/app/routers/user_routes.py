from fastapi import APIRouter, HTTPException, status
from typing import List
from app.schemas.user_schema import UserCreate, UserUpdate, UserResponse
from app.services.user_service import UserService

router = APIRouter()
user_service = UserService()

@router.post("/users", status_code=status.HTTP_201_CREATED)
async def create_user(user: UserCreate) -> UserResponse:
    return await user_service.create_user(user)

@router.get("/users", status_code=status.HTTP_200_OK)
async def get_users(skip: int = 0, limit: int = 100) -> List[UserResponse]:
    return await user_service.get_users(skip=skip, limit=limit)

@router.get("/users/{user_id}", status_code=status.HTTP_200_OK)
async def get_user(user_id: int) -> UserResponse:
    user = await user_service.get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

@router.put("/users/{user_id}", status_code=status.HTTP_200_OK)
async def update_user(user_id: int, user: UserUpdate) -> UserResponse:
    updated_user = await user_service.update_user(user_id, user)
    if not updated_user:
        raise HTTPException(status_code=404, detail="User not found")
    return updated_user

@router.delete("/users/{user_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_user(user_id: int):
    success = await user_service.delete_user(user_id)
    if not success:
        raise HTTPException(status_code=404, detail="User not found")