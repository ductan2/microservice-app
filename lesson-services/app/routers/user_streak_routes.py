from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import Optional
from uuid import UUID
from datetime import date

# Import dependencies
# from app.database.connection import get_db
# from app.services.user_streak_service import UserStreakService
# from app.schemas.streak_schema import UserStreakResponse

router = APIRouter(prefix="/api/streaks", tags=["User Streaks"])

# GET /api/streaks/user/{user_id}
# Logic: Get current streak information for user
# - Fetch user_streaks record by user_id
# - Return current_len, longest_len, last_day
# - Return empty/zero if no streak record

# POST /api/streaks/user/{user_id}/check
# Logic: Check and update streak based on today's activity
# - Get today's date
# - Fetch user_streak record
# - Check if user has activity today (from daily_activity)
# - If activity today and last_day was yesterday: increment current_len
# - If activity today and last_day was today: no change
# - If activity today and last_day > 1 day ago: reset current_len to 1
# - If no activity today and last_day was yesterday: streak broken, reset to 0
# - Update longest_len if current_len > longest_len
# - Update last_day to today if activity
# - Return updated streak

# GET /api/streaks/user/{user_id}/status
# Logic: Get streak status and risk level
# - Get user_streak record
# - Check today's activity
# - Calculate risk level:
#   - Safe: activity today
#   - At risk: no activity today but last_day was yesterday
#   - Broken: no activity and last_day < yesterday
# - Return status with current streak and risk level

# GET /api/streaks/leaderboard
# Logic: Get top users by current streak
# - Query user_streaks
# - Order by current_len DESC
# - Limit to top 50 or query param
# - Return leaderboard of users with longest current streaks

