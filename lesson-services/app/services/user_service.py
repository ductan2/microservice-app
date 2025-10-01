from typing import List, Optional
from app.schemas.user_schema import UserCreate, UserUpdate, UserResponse

class UserService:
    def __init__(self):
        pass
    
    async def create_user(self, user: UserCreate) -> UserResponse:
        pass
    
    async def get_users(self, skip: int = 0, limit: int = 100) -> List[UserResponse]:
        pass
    
    async def get_user(self, user_id: int) -> Optional[UserResponse]:
        pass
    
    async def update_user(self, user_id: int, user: UserUpdate) -> Optional[UserResponse]:
        pass
    
    async def delete_user(self, user_id: int) -> bool:
        pass