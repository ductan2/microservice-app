from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from uuid import UUID

# Import dependencies
# from app.database.connection import get_db
# from app.services.dim_user_service import DimUserService
# from app.schemas.user_schema import DimUserCreate, DimUserUpdate, DimUserResponse

router = APIRouter(prefix="/api/users", tags=["User Preferences"])

# GET /api/users/{user_id}
# Logic: Retrieve user preferences (locale, level_hint) by user_id
# - Validate user_id format
# - Call service to fetch user from database
# - Return 404 if not found
# - Return user data with locale and level_hint

# POST /api/users
# Logic: Create or initialize user preferences when first time user
# - Validate request body (user_id, locale, level_hint optional)
# - Check if user already exists - if yes, return 409 Conflict
# - Call service to create new user record
# - Return created user data with 201 status

# PUT /api/users/{user_id}
# Logic: Update user preferences (locale, learning level hint)
# - Validate user_id and request body
# - Call service to update user preferences
# - Update updated_at timestamp automatically
# - Return updated user data

# PATCH /api/users/{user_id}/locale
# Logic: Partially update only locale preference
# - Validate locale value (e.g., 'en', 'vi', 'ja')
# - Call service to update just locale field
# - Return updated user data

# DELETE /api/users/{user_id}
# Logic: Soft delete or remove user preferences
# - Check if user has any active progress data
# - If yes, warn or prevent deletion
# - If no, delete user record
# - Return 204 No Content on success
