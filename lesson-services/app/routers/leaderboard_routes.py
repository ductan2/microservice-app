from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime

# Import dependencies
# from app.database.connection import get_db
# from app.services.leaderboard_service import LeaderboardService
# from app.schemas.leaderboard_schema import LeaderboardSnapshotResponse

router = APIRouter(prefix="/api/leaderboards", tags=["Leaderboards"])

# GET /api/leaderboards/weekly/current
# Logic: Get current week's leaderboard
# - Calculate current week key (e.g., "2025-W14")
# - Fetch latest leaderboard_snapshots for period='weekly' and current week
# - Order by rank ASC
# - Return ranked list of users with points

# GET /api/leaderboards/monthly/current
# Logic: Get current month's leaderboard
# - Calculate current month key (e.g., "2025-03")
# - Fetch latest snapshots for period='monthly' and current month
# - Order by rank ASC
# - Return ranked list

# GET /api/leaderboards/weekly/history
# Logic: Get historical weekly leaderboards
# - Query params: weeks_back (default 4)
# - Fetch snapshots for last N weeks
# - Group by period_key
# - Return historical leaderboard data

# GET /api/leaderboards/monthly/history
# Logic: Get historical monthly leaderboards
# - Query params: months_back (default 6)
# - Fetch snapshots for last N months
# - Group by period_key
# - Return historical data

# POST /api/leaderboards/snapshot/weekly
# Logic: Create weekly leaderboard snapshot (cron job)
# - Get current week key
# - Query user_points, order by weekly DESC
# - Take top 100 users
# - For each user, create leaderboard_snapshot record:
#   - period = 'weekly'
#   - period_key = current week
#   - rank = their position (1-100)
#   - user_id, points = weekly points
#   - taken_at = current timestamp
# - Bulk insert snapshots
# - Return count of snapshots created
# - Called by scheduler every Sunday night

# POST /api/leaderboards/snapshot/monthly
# Logic: Create monthly leaderboard snapshot (cron job)
# - Get current month key
# - Query user_points, order by monthly DESC
# - Take top 100 users
# - Create snapshot records for period='monthly'
# - Same structure as weekly
# - Bulk insert
# - Return count
# - Called by scheduler on last day of month

# GET /api/leaderboards/user/{user_id}/history
# Logic: Get user's leaderboard history
# - Fetch all leaderboard_snapshots for user_id
# - Order by taken_at DESC
# - Group by period and period_key
# - Return user's historical ranks and points

# GET /api/leaderboards/week/{week_key}
# Logic: Get specific week's leaderboard
# - week_key format: "2025-W14"
# - Fetch snapshots for period='weekly' and period_key=week_key
# - Take latest snapshot per user (in case multiple)
# - Order by rank ASC
# - Return historical leaderboard for that week

# GET /api/leaderboards/month/{month_key}
# Logic: Get specific month's leaderboard
# - month_key format: "2025-03"
# - Fetch snapshots for period='monthly' and period_key=month_key
# - Take latest snapshot per user
# - Order by rank ASC
# - Return historical leaderboard for that month

