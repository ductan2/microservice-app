from sqlalchemy.orm import Session
from typing import Optional, List, Dict
from uuid import UUID
from datetime import date, timedelta

# from app.models.progress_models import UserStreak, DailyActivity
# from app.schemas.streak_schema import UserStreakResponse

class UserStreakService:
    def __init__(self, db: Session):
        self.db = db
    
    # get_user_streak(user_id: UUID) -> Optional[UserStreak]
    # Logic: Get user's streak record
    # - Query user_streaks by user_id (primary key)
    # - Return streak record or None
    
    # check_and_update_streak(user_id: UUID, activity_date: date) -> UserStreak
    # Logic: Update streak based on activity
    # - Find or create user_streak record for user
    # - Check daily_activity for activity_date to confirm activity
    # - If no activity that day, return current streak unchanged
    # - Calculate logic based on last_day:
    #   Case 1: last_day is None (first time)
    #     - Set current_len = 1
    #     - Set longest_len = 1
    #     - Set last_day = activity_date
    #   Case 2: activity_date == last_day
    #     - No change (already counted)
    #   Case 3: activity_date == last_day + 1 day
    #     - Increment current_len += 1
    #     - If current_len > longest_len: update longest_len
    #     - Set last_day = activity_date
    #   Case 4: activity_date > last_day + 1 day
    #     - Streak broken, reset current_len = 1
    #     - longest_len stays same
    #     - Set last_day = activity_date
    # - Commit transaction
    # - Return updated streak record
    
    # break_streak(user_id: UUID) -> UserStreak
    # Logic: Mark streak as broken (for missed day)
    # - Find user_streak record
    # - Set current_len = 0
    # - longest_len stays unchanged
    # - Keep last_day as is (for reference)
    # - Commit and return
    
    # get_streak_status(user_id: UUID) -> Dict
    # Logic: Get detailed streak status and risk
    # - Get user_streak record
    # - Get today's date
    # - Check if user has activity today (query daily_activity)
    # - Calculate streak status:
    #   - "active": has activity today and last_day == today or yesterday
    #   - "at_risk": no activity today but last_day == yesterday
    #   - "broken": last_day < yesterday
    # - Return status dict: {
    #     current_len, longest_len, last_day,
    #     status, has_activity_today, days_since_last
    #   }
    
    # get_streak_leaderboard(limit: int = 50) -> List[UserStreak]
    # Logic: Get top users by current streak
    # - Query user_streaks
    # - Filter: current_len > 0
    # - Order by current_len DESC, last_day DESC
    # - Limit to specified number
    # - Return list of top streaks
    
    # initialize_streak(user_id: UUID) -> UserStreak
    # Logic: Create initial streak record
    # - Create new user_streak with:
    #   - user_id
    #   - current_len = 0
    #   - longest_len = 0
    #   - last_day = None
    # - Insert and commit
    # - Return created record
    
    # get_or_create_streak(user_id: UUID) -> UserStreak
    # Logic: Get existing streak or create new one
    # - Try to fetch user_streak by user_id
    # - If not found, call initialize_streak
    # - Return streak record
    
    # recalculate_streak(user_id: UUID) -> UserStreak
    # Logic: Recalculate streak from daily_activity history
    # - Fetch all daily_activity for user ordered by activity_dt DESC
    # - Start from most recent date
    # - Count consecutive days with activity
    # - That's the current_len
    # - Scan all history to find longest consecutive streak
    # - Update user_streak record with calculated values
    # - Set last_day to most recent activity date
    # - Commit and return
    # - Use this for data correction or migration

