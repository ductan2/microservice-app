from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from uuid import UUID
from datetime import datetime

# Import dependencies
# from app.database.connection import get_db
# from app.services.sr_card_service import SRCardService
# from app.schemas.sr_schema import SRCardCreate, SRCardResponse

router = APIRouter(prefix="/api/spaced-repetition/cards", tags=["Spaced Repetition Cards"])

# GET /api/spaced-repetition/cards/user/{user_id}
# Logic: Get all SR cards for user
# - Optional query params: suspended (true/false), due_only (true)
# - Fetch all sr_cards for user_id
# - Apply filters if provided
# - Return list of cards with SRS data

# GET /api/spaced-repetition/cards/user/{user_id}/due
# Logic: Get cards due for review today
# - Fetch cards where due_at <= current_time
# - Filter out suspended cards
# - Order by due_at ASC (most overdue first)
# - Return list of due cards for review session

# POST /api/spaced-repetition/cards
# Logic: Create new SR card for flashcard
# - Validate request body (user_id, flashcard_id)
# - Check if card already exists for this user+flashcard
# - If exists, return existing card
# - Initialize new card with default values:
#   - ease_factor = 2.5
#   - interval_d = 0
#   - repetition = 0
#   - due_at = current_time (due immediately)
#   - suspended = false
# - Insert record and return created card

# GET /api/spaced-repetition/cards/{card_id}
# Logic: Get specific SR card details
# - Validate card_id
# - Fetch sr_card record
# - Return card with SRS parameters

# PATCH /api/spaced-repetition/cards/{card_id}/suspend
# Logic: Suspend card from review rotation
# - Update suspended = true
# - Card won't appear in due queue
# - Return updated card

# PATCH /api/spaced-repetition/cards/{card_id}/unsuspend
# Logic: Reactivate suspended card
# - Update suspended = false
# - Card will appear in due queue based on due_at
# - Return updated card

# DELETE /api/spaced-repetition/cards/{card_id}
# Logic: Delete SR card (remove from user's deck)
# - Delete sr_card record
# - Related reviews remain for history
# - Return 204 No Content

# GET /api/spaced-repetition/cards/user/{user_id}/stats
# Logic: Get SRS statistics for user
# - Count total cards
# - Count due cards
# - Count suspended cards
# - Calculate average ease_factor
# - Group cards by interval ranges
# - Return statistics object

