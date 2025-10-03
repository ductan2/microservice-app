from sqlalchemy.orm import Session
from typing import Optional
from uuid import UUID
from datetime import datetime

# from app.models.progress_models import DimUser
# from app.schemas.user_schema import DimUserCreate, DimUserUpdate

class DimUserService:
    def __init__(self, db: Session):
        self.db = db
    
    # get_user_by_id(user_id: UUID) -> Optional[DimUser]
    # Logic: Query dim_users table by user_id
    # - Execute SELECT query with user_id filter
    # - Return user object or None if not found
    
    # create_user(user_data: DimUserCreate) -> DimUser
    # Logic: Create new user preference record
    # - Generate UUID if not provided
    # - Set default locale if not provided (e.g., 'en')
    # - Set level_hint if provided (beginner, intermediate, advanced)
    # - Set updated_at to current timestamp
    # - Insert into database
    # - Commit transaction
    # - Return created user object
    
    # update_user(user_id: UUID, user_data: DimUserUpdate) -> Optional[DimUser]
    # Logic: Update existing user preferences
    # - Find user by user_id
    # - If not found, return None
    # - Update locale if provided
    # - Update level_hint if provided
    # - Update updated_at to current timestamp
    # - Commit transaction
    # - Return updated user object
    
    # update_locale(user_id: UUID, locale: str) -> Optional[DimUser]
    # Logic: Update only user locale
    # - Find user by user_id
    # - Update locale field
    # - Update updated_at timestamp
    # - Commit and return updated user
    
    # delete_user(user_id: UUID) -> bool
    # Logic: Delete user preference record
    # - Find user by user_id
    # - Check for dependencies (active lessons, progress)
    # - If safe, delete record
    # - Commit transaction
    # - Return True if successful, False otherwise
    
    # user_exists(user_id: UUID) -> bool
    # Logic: Check if user exists in dim_users
    # - Execute COUNT query with user_id filter
    # - Return True if count > 0, else False

