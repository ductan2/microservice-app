from sqlalchemy.orm import Session
from typing import List, Optional, Dict
from uuid import UUID
from datetime import datetime

# from app.models.progress_models import UserPoints
# from app.schemas.points_schema import UserPointsCreate

class UserPointsService:
    def __init__(self, db: Session):
        self.db = db
    
    # get_user_points(user_id: UUID) -> Optional[UserPoints]
    # Logic: Get user's point record
    # - Query user_points by user_id (primary key)
    # - Return points record or None
    
    # add_points(user_id: UUID, points: int) -> UserPoints
    # Logic: Add points to all categories
    # - Find or create user_points record
    # - Increment lifetime += points
    # - Increment weekly += points
    # - Increment monthly += points
    # - Update updated_at = current timestamp
    # - Use upsert pattern if needed
    # - Commit transaction
    # - Return updated points record
    
    # get_lifetime_leaderboard(limit: int = 100, offset: int = 0) -> List[UserPoints]
    # Logic: Get top users by lifetime points
    # - Query user_points
    # - Order by lifetime DESC
    # - Apply limit and offset for pagination
    # - Return ranked list
    
    # get_weekly_leaderboard(limit: int = 100, offset: int = 0) -> List[UserPoints]
    # Logic: Get top users by weekly points
    # - Query user_points
    # - Order by weekly DESC
    # - Apply pagination
    # - Return ranked list for current week
    
    # get_monthly_leaderboard(limit: int = 100, offset: int = 0) -> List[UserPoints]
    # Logic: Get top users by monthly points
    # - Query user_points
    # - Order by monthly DESC
    # - Apply pagination
    # - Return ranked list for current month
    
    # reset_weekly_points() -> int
    # Logic: Reset weekly points for all users
    # - Execute UPDATE user_points SET weekly = 0
    # - Also update updated_at = current timestamp
    # - Commit transaction
    # - Return count of updated records
    # - Called by scheduler every Monday at midnight
    
    # reset_monthly_points() -> int
    # Logic: Reset monthly points for all users
    # - Execute UPDATE user_points SET monthly = 0
    # - Update updated_at = current timestamp
    # - Commit transaction
    # - Return count of updated records
    # - Called by scheduler on 1st day of month
    
    # get_user_ranks(user_id: UUID) -> Dict[str, int]
    # Logic: Calculate user's rank in all leaderboards
    # - Get user's points record
    # - Calculate lifetime rank:
    #   - Count users with lifetime > user's lifetime
    #   - Rank = count + 1
    # - Calculate weekly rank similarly
    # - Calculate monthly rank similarly
    # - Return dict: {
    #     'lifetime_rank': X,
    #     'weekly_rank': Y,
    #     'monthly_rank': Z
    #   }
    
    # initialize_user_points(user_id: UUID) -> UserPoints
    # Logic: Create initial points record for new user
    # - Create new user_points with:
    #   - user_id
    #   - lifetime = 0
    #   - weekly = 0
    #   - monthly = 0
    #   - updated_at = current timestamp
    # - Insert and commit
    # - Return created record
    
    # get_or_create_points(user_id: UUID) -> UserPoints
    # Logic: Get existing or create new points record
    # - Try to fetch user_points by user_id
    # - If not found, call initialize_user_points
    # - Return points record
    
    # subtract_points(user_id: UUID, points: int) -> UserPoints
    # Logic: Subtract points from all categories (for corrections)
    # - Get user_points record
    # - Decrement lifetime -= points (don't go below 0)
    # - Decrement weekly -= points (don't go below 0)
    # - Decrement monthly -= points (don't go below 0)
    # - Update updated_at
    # - Commit and return
    
    # get_top_users_with_details(period: str, limit: int = 10) -> List[Dict]
    # Logic: Get leaderboard with user details
    # - Query user_points with appropriate ordering
    # - Join with dim_users to get user info
    # - Optionally join to get user profile data
    # - Return enriched leaderboard with user names/avatars
    # - period: 'lifetime', 'weekly', or 'monthly'

