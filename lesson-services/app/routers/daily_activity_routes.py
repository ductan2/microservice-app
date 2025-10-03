from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from uuid import UUID
from datetime import date, datetime

# Import dependencies
# from app.database.connection import get_db
# from app.services.daily_activity_service import DailyActivityService
# from app.schemas.activity_schema import DailyActivityResponse, DailyActivityStats

router = APIRouter(prefix="/api/daily-activity", tags=["Daily Activity"])

# GET /api/daily-activity/user/{user_id}/today
# Logic: Get today's activity for user
# - Get current date
# - Fetch daily_activity for user_id and today's date
# - Return today's stats (lessons, quizzes, minutes, points)
# - Return empty/zero stats if no activity today

# GET /api/daily-activity/user/{user_id}/date/{activity_date}
# Logic: Get activity for specific date
# - Validate date format
# - Fetch daily_activity for user_id and specified date
# - Return activity stats for that date

# GET /api/daily-activity/user/{user_id}/range
# Logic: Get activity for date range
# - Query params: date_from, date_to (default last 30 days)
# - Fetch all daily_activity records in range
# - Order by activity_dt ASC
# - Return list of daily activity records

# GET /api/daily-activity/user/{user_id}/week
# Logic: Get current week's activity (Mon-Sun)
# - Calculate start and end of current week
# - Fetch daily_activity for week range
# - Return 7 days of data (fill missing days with zeros)

# GET /api/daily-activity/user/{user_id}/month
# Logic: Get current month's activity
# - Calculate start and end of current month
# - Fetch daily_activity for month
# - Aggregate totals: total lessons, quizzes, minutes, points
# - Return monthly summary with daily breakdown

# GET /api/daily-activity/user/{user_id}/stats/summary
# Logic: Get aggregated activity statistics
# - Calculate total lifetime stats (all-time totals)
# - Calculate last 7 days totals
# - Calculate last 30 days totals
# - Calculate averages (daily average)
# - Find most active day
# - Return comprehensive summary

# POST /api/daily-activity/increment
# Logic: Increment activity counters (internal use)
# - Validate request body (user_id, date, field to increment, amount)
# - Find or create daily_activity record for user+date
# - Increment specified field (lessons_completed, quizzes_completed, minutes, points)
# - Commit transaction
# - Return updated activity record
# - Note: This is called by other services when activities complete

