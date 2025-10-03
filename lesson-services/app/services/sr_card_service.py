from sqlalchemy.orm import Session
from typing import List, Optional, Dict
from uuid import UUID
from datetime import datetime, timedelta

# from app.models.progress_models import SRCard
# from app.schemas.sr_schema import SRCardCreate

class SRCardService:
    def __init__(self, db: Session):
        self.db = db
    
    # get_user_cards(user_id: UUID, suspended: Optional[bool] = None, due_only: bool = False) -> List[SRCard]
    # Logic: Get user's SR cards with filters
    # - Query sr_cards by user_id
    # - Apply suspended filter if provided
    # - If due_only=True, filter where due_at <= now and suspended=False
    # - Return list of cards
    
    # get_due_cards(user_id: UUID) -> List[SRCard]
    # Logic: Get cards due for review
    # - Query sr_cards by user_id
    # - Filter: due_at <= current_time AND suspended = false
    # - Order by due_at ASC (most overdue first)
    # - Return list of due cards ready for review
    
    # create_card(user_id: UUID, flashcard_id: UUID) -> SRCard
    # Logic: Create new SR card with default SRS parameters
    # - Check if card already exists for user+flashcard combination
    # - If exists, return existing card
    # - Generate new UUID for card id
    # - Initialize SM-2 algorithm defaults:
    #   - ease_factor = 2.5 (standard starting difficulty)
    #   - interval_d = 0 (new card, due today)
    #   - repetition = 0 (never reviewed)
    #   - due_at = current timestamp (immediately available)
    #   - suspended = false
    # - Insert into database and commit
    # - Return created card
    
    # get_card(card_id: UUID) -> Optional[SRCard]
    # Logic: Get specific SR card by ID
    # - Query sr_cards by card_id
    # - Return card object or None
    
    # suspend_card(card_id: UUID) -> Optional[SRCard]
    # Logic: Suspend card from review rotation
    # - Find card by card_id
    # - Update suspended = true
    # - Commit transaction
    # - Return updated card
    
    # unsuspend_card(card_id: UUID) -> Optional[SRCard]
    # Logic: Reactivate suspended card
    # - Find card by card_id
    # - Update suspended = false
    # - Commit transaction
    # - Return updated card
    
    # update_card_after_review(card_id: UUID, quality: int) -> SRCard
    # Logic: Update SR card after review using SM-2 algorithm
    # - Find card by card_id
    # - Apply SM-2 spaced repetition algorithm:
    #   - If quality >= 3 (correct):
    #     - If repetition == 0: interval = 1 day
    #     - If repetition == 1: interval = 6 days
    #     - If repetition >= 2: interval = previous_interval * ease_factor
    #     - Increment repetition
    #   - If quality < 3 (incorrect):
    #     - Reset repetition = 0
    #     - Reset interval = 0 (due today)
    #   - Update ease_factor:
    #     - new_ef = old_ef + (0.1 - (5 - quality) * (0.08 + (5 - quality) * 0.02))
    #     - Clamp ease_factor to minimum 1.3
    # - Calculate new due_at = current_time + interval_d days
    # - Update card with new values
    # - Commit transaction
    # - Return updated card
    
    # delete_card(card_id: UUID) -> bool
    # Logic: Delete SR card from user's deck
    # - Find and delete sr_card record
    # - Keep sr_reviews for historical data
    # - Commit transaction
    # - Return True if successful
    
    # get_user_stats(user_id: UUID) -> Dict
    # Logic: Calculate SRS statistics for user
    # - Count total cards (not suspended)
    # - Count due cards (due_at <= now, not suspended)
    # - Count suspended cards
    # - Count new cards (repetition == 0)
    # - Count learning cards (repetition < 3)
    # - Count mature cards (repetition >= 3)
    # - Calculate average ease_factor
    # - Calculate average interval
    # - Return comprehensive statistics dictionary
    
    # get_card_by_flashcard(user_id: UUID, flashcard_id: UUID) -> Optional[SRCard]
    # Logic: Get SR card for specific flashcard
    # - Query by user_id and flashcard_id
    # - Return card or None if not found

