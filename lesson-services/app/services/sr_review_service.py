from sqlalchemy.orm import Session
from typing import List, Optional, Dict
from uuid import UUID
from datetime import datetime, date, timedelta

# from app.models.progress_models import SRReview, SRCard
# from app.schemas.sr_schema import SRReviewCreate

class SRReviewService:
    def __init__(self, db: Session):
        self.db = db
    
    # create_review(user_id: UUID, flashcard_id: UUID, quality: int) -> SRReview
    # Logic: Record flashcard review and update SR card
    # - Validate quality is between 0-5
    # - Find SR card by user_id + flashcard_id
    # - If card doesn't exist, create new card first
    # - Capture current card state (prev_interval = card.interval_d)
    # - Generate new UUID for review id
    # - Set reviewed_at = current timestamp
    # - Call SR card service to update card based on quality
    # - Capture new card state (new_interval, new_ef)
    # - Create sr_review record with:
    #   - user_id, flashcard_id, quality
    #   - prev_interval (before update)
    #   - new_interval (after update)
    #   - new_ef (new ease factor)
    #   - reviewed_at timestamp
    # - Insert review record and commit
    # - Update daily_activity: increment minutes (estimate ~2 mins per review)
    # - Update daily_activity: add points (e.g., 5 points per review)
    # - Return created review record
    
    # get_user_reviews(user_id: UUID, limit: int = 100, offset: int = 0, date_from: Optional[date] = None, date_to: Optional[date] = None) -> List[SRReview]
    # Logic: Get paginated review history for user
    # - Query sr_reviews by user_id
    # - Apply date range filters if provided
    # - Order by reviewed_at DESC
    # - Apply pagination (limit, offset)
    # - Return list of review records
    
    # get_flashcard_reviews(user_id: UUID, flashcard_id: UUID) -> List[SRReview]
    # Logic: Get review history for specific flashcard
    # - Query sr_reviews by user_id and flashcard_id
    # - Order by reviewed_at DESC
    # - Return list showing learning progress over time
    
    # get_today_reviews(user_id: UUID) -> List[SRReview]
    # Logic: Get today's review activity
    # - Calculate start_of_today = midnight today
    # - Query reviews where reviewed_at >= start_of_today
    # - Filter by user_id
    # - Return list of today's reviews
    
    # get_today_stats(user_id: UUID) -> Dict
    # Logic: Calculate today's review statistics
    # - Get today's reviews
    # - Count total reviews
    # - Calculate average quality
    # - Group count by quality rating (0-5)
    # - Calculate retention rate (quality >= 3 count / total)
    # - Return statistics dictionary
    
    # get_user_review_stats(user_id: UUID) -> Dict
    # Logic: Get comprehensive review statistics
    # - Count total reviews all time
    # - Calculate average quality score
    # - Count reviews by quality (distribution 0-5)
    # - Calculate retention rate (quality >= 3)
    # - Count unique flashcards reviewed
    # - Calculate review streak:
    #   - Query distinct review dates
    #   - Count consecutive days from today backward
    # - Get busiest review day (date with most reviews)
    # - Calculate total time spent (reviews * avg_time_per_review)
    # - Return comprehensive statistics dictionary
    
    # delete_review(review_id: UUID) -> bool
    # Logic: Delete review record
    # - Note: This is for data cleanup only
    # - Warning: Does not revert SR card state
    # - Find and delete sr_review record
    # - Commit transaction
    # - Return True if successful
    
    # get_review_streak(user_id: UUID) -> int
    # Logic: Calculate consecutive days reviewed
    # - Query distinct dates from sr_reviews for user
    # - Order by reviewed_at DESC
    # - Start from today and count backward
    # - Stop when gap of more than 1 day found
    # - Return streak length in days
    
    # get_review_calendar(user_id: UUID, year: int, month: int) -> Dict[date, int]
    # Logic: Get review activity calendar for month
    # - Query reviews for specified month/year
    # - Group by date
    # - Count reviews per day
    # - Return dictionary: {date: review_count}

