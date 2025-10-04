from datetime import datetime, timedelta
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy import func
from sqlalchemy.orm import Session

from app.models.progress_models import SRCard
from app.schemas.progress_schema import SRCardCreate


class SRCardService:
    def __init__(self, db: Session):
        self.db = db

    def get_user_cards(
        self,
        user_id: UUID,
        suspended: Optional[bool] = None,
        due_only: bool = False,
    ) -> List[SRCard]:
        query = self.db.query(SRCard).filter(SRCard.user_id == user_id)

        if suspended is not None:
            query = query.filter(SRCard.suspended == suspended)

        if due_only:
            now = datetime.utcnow()
            query = query.filter(
                SRCard.due_at <= now,
                SRCard.suspended.is_(False),
            )

        return query.order_by(SRCard.due_at.asc()).all()

    def get_due_cards(self, user_id: UUID) -> List[SRCard]:
        return self.get_user_cards(user_id=user_id, due_only=True)

    def create_card(self, card_data: SRCardCreate) -> SRCard:
        existing = self.get_card_by_flashcard(card_data.user_id, card_data.flashcard_id)
        if existing:
            return existing

        payload = card_data.model_dump()
        card = SRCard(**payload)
        card.due_at = datetime.utcnow()

        self.db.add(card)
        self.db.commit()
        self.db.refresh(card)
        return card

    def get_card(self, card_id: UUID) -> Optional[SRCard]:
        return self.db.query(SRCard).filter(SRCard.id == card_id).one_or_none()

    def suspend_card(self, card_id: UUID) -> Optional[SRCard]:
        card = self.get_card(card_id)
        if not card:
            return None

        card.suspended = True
        self.db.commit()
        self.db.refresh(card)
        return card

    def unsuspend_card(self, card_id: UUID) -> Optional[SRCard]:
        card = self.get_card(card_id)
        if not card:
            return None

        card.suspended = False
        self.db.commit()
        self.db.refresh(card)
        return card

    def update_card_after_review(self, card_id: UUID, quality: int) -> Optional[SRCard]:
        card = self.get_card(card_id)
        if not card:
            return None

        quality = max(0, min(5, quality))
        now = datetime.utcnow()

        if quality >= 3:
            if card.repetition == 0:
                interval = 1
            elif card.repetition == 1:
                interval = 6
            else:
                interval = max(1, round(card.interval_d * card.ease_factor))
            card.repetition += 1
            card.interval_d = int(interval)
        else:
            card.repetition = 0
            card.interval_d = 0

        new_ef = card.ease_factor + (
            0.1 - (5 - quality) * (0.08 + (5 - quality) * 0.02)
        )
        card.ease_factor = max(1.3, round(new_ef, 2))

        card.due_at = now + timedelta(days=card.interval_d)
        card.suspended = False

        self.db.commit()
        self.db.refresh(card)
        return card

    def delete_card(self, card_id: UUID) -> bool:
        card = self.get_card(card_id)
        if not card:
            return False

        self.db.delete(card)
        self.db.commit()
        return True

    def get_user_stats(self, user_id: UUID) -> Dict[str, float]:
        now = datetime.utcnow()
        base_query = self.db.query(SRCard).filter(SRCard.user_id == user_id)

        total_cards = base_query.count()
        suspended_cards = base_query.filter(SRCard.suspended.is_(True)).count()
        due_cards = (
            base_query.filter(SRCard.suspended.is_(False))
            .filter(SRCard.due_at <= now)
            .count()
        )
        new_cards = (
            base_query.filter(SRCard.repetition == 0, SRCard.suspended.is_(False)).count()
        )
        learning_cards = (
            base_query.filter(
                SRCard.repetition > 0,
                SRCard.repetition < 3,
                SRCard.suspended.is_(False),
            ).count()
        )
        mature_cards = (
            base_query.filter(
                SRCard.repetition >= 3,
                SRCard.suspended.is_(False),
            ).count()
        )

        avg_ease = (
            self.db.query(func.avg(SRCard.ease_factor))
            .filter(SRCard.user_id == user_id, SRCard.suspended.is_(False))
            .scalar()
        )
        avg_interval = (
            self.db.query(func.avg(SRCard.interval_d))
            .filter(SRCard.user_id == user_id, SRCard.suspended.is_(False))
            .scalar()
        )

        return {
            "total_cards": total_cards,
            "due_cards": due_cards,
            "suspended_cards": suspended_cards,
            "new_cards": new_cards,
            "learning_cards": learning_cards,
            "mature_cards": mature_cards,
            "average_ease_factor": float(avg_ease or 0.0),
            "average_interval": float(avg_interval or 0.0),
        }

    def get_card_by_flashcard(
        self, user_id: UUID, flashcard_id: UUID
    ) -> Optional[SRCard]:
        return (
            self.db.query(SRCard)
            .filter(
                SRCard.user_id == user_id,
                SRCard.flashcard_id == flashcard_id,
            )
            .one_or_none()
        )
