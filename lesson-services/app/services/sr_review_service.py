from datetime import date, datetime, timedelta
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy import func
from sqlalchemy.orm import Session

from app.models.progress_models import SRReview
from app.schemas.progress_schema import SRCardCreate, SRReviewCreate
from app.services.daily_activity_service import DailyActivityService
from app.services.sr_card_service import SRCardService

REVIEW_MINUTES_PER_CARD = 2
REVIEW_POINTS = 5


class SRReviewService:
    def __init__(self, db: Session):
        self.db = db
        self.card_service = SRCardService(db)
        self.activity_service = DailyActivityService(db)

    def create_review(self, review_data: SRReviewCreate) -> SRReview:
        quality = review_data.quality
        if quality < 0 or quality > 5:
            raise ValueError("Quality must be between 0 and 5")

        card = self.card_service.get_card_by_flashcard(
            review_data.user_id, review_data.flashcard_id
        )
        if card is None:
            card = self.card_service.create_card(
                SRCardCreate(
                    user_id=review_data.user_id,
                    flashcard_id=review_data.flashcard_id,
                )
            )

        prev_interval = card.interval_d
        updated_card = self.card_service.update_card_after_review(card.id, quality)
        if updated_card is None:
            raise ValueError("Unable to update spaced repetition card")

        review = SRReview(
            user_id=review_data.user_id,
            flashcard_id=review_data.flashcard_id,
            quality=quality,
            prev_interval=prev_interval,
            new_interval=updated_card.interval_d,
            new_ef=updated_card.ease_factor,
            reviewed_at=datetime.utcnow(),
        )

        self.db.add(review)
        self.db.flush()

        today = review.reviewed_at.date()
        self.activity_service.increment_activity(
            review.user_id, today, "minutes", REVIEW_MINUTES_PER_CARD
        )
        self.activity_service.increment_activity(
            review.user_id, today, "points", REVIEW_POINTS
        )

        self.db.refresh(review)
        return review

    def get_user_reviews(
        self,
        user_id: UUID,
        limit: int = 100,
        offset: int = 0,
        date_from: Optional[date] = None,
        date_to: Optional[date] = None,
    ) -> List[SRReview]:
        query = self.db.query(SRReview).filter(SRReview.user_id == user_id)

        if date_from:
            start = datetime.combine(date_from, datetime.min.time())
            query = query.filter(SRReview.reviewed_at >= start)
        if date_to:
            end = datetime.combine(date_to + timedelta(days=1), datetime.min.time())
            query = query.filter(SRReview.reviewed_at < end)

        return (
            query.order_by(SRReview.reviewed_at.desc())
            .offset(offset)
            .limit(limit)
            .all()
        )

    def get_flashcard_reviews(
        self, user_id: UUID, flashcard_id: UUID
    ) -> List[SRReview]:
        return (
            self.db.query(SRReview)
            .filter(
                SRReview.user_id == user_id,
                SRReview.flashcard_id == flashcard_id,
            )
            .order_by(SRReview.reviewed_at.desc())
            .all()
        )

    def get_today_reviews(self, user_id: UUID) -> List[SRReview]:
        start_of_day = datetime.combine(date.today(), datetime.min.time())
        end_of_day = start_of_day + timedelta(days=1)
        return (
            self.db.query(SRReview)
            .filter(
                SRReview.user_id == user_id,
                SRReview.reviewed_at >= start_of_day,
                SRReview.reviewed_at < end_of_day,
            )
            .order_by(SRReview.reviewed_at.desc())
            .all()
        )

    def get_today_stats(self, user_id: UUID) -> Dict[str, object]:
        reviews = self.get_today_reviews(user_id)
        return self._calculate_review_stats(reviews)

    def get_user_review_stats(self, user_id: UUID) -> Dict[str, object]:
        reviews = (
            self.db.query(SRReview)
            .filter(SRReview.user_id == user_id)
            .order_by(SRReview.reviewed_at.desc())
            .all()
        )
        stats = self._calculate_review_stats(reviews)
        stats["review_streak"] = self.get_review_streak(user_id)
        stats["unique_flashcards"] = (
            self.db.query(func.count(func.distinct(SRReview.flashcard_id)))
            .filter(SRReview.user_id == user_id)
            .scalar()
            or 0
        )

        busiest = (
            self.db.query(
                func.date(SRReview.reviewed_at).label("review_date"),
                func.count(SRReview.id).label("review_count"),
            )
            .filter(SRReview.user_id == user_id)
            .group_by("review_date")
            .order_by(func.count(SRReview.id).desc(), func.date(SRReview.reviewed_at).desc())
            .first()
        )
        if busiest:
            stats["busiest_day"] = busiest.review_date
            stats["busiest_day_count"] = busiest.review_count
        else:
            stats["busiest_day"] = None
            stats["busiest_day_count"] = 0

        stats["total_time_minutes"] = stats["total_reviews"] * REVIEW_MINUTES_PER_CARD
        return stats

    def delete_review(self, review_id: UUID) -> bool:
        review = self.db.query(SRReview).filter(SRReview.id == review_id).one_or_none()
        if review is None:
            return False

        self.db.delete(review)
        self.db.commit()
        return True

    def get_review_streak(self, user_id: UUID) -> int:
        distinct_dates = (
            self.db.query(func.date(SRReview.reviewed_at))
            .filter(SRReview.user_id == user_id)
            .distinct()
            .order_by(func.date(SRReview.reviewed_at).desc())
            .all()
        )

        today = date.today()
        streak = 0
        expected = today

        for (review_date,) in distinct_dates:
            if review_date == expected:
                streak += 1
                expected = expected - timedelta(days=1)
            elif review_date < expected:
                break

        return streak

    def get_review_calendar(
        self, user_id: UUID, year: int, month: int
    ) -> Dict[date, int]:
        start = date(year, month, 1)
        if month == 12:
            end = date(year + 1, 1, 1)
        else:
            end = date(year, month + 1, 1)

        results = (
            self.db.query(
                func.date(SRReview.reviewed_at).label("review_date"),
                func.count(SRReview.id).label("review_count"),
            )
            .filter(
                SRReview.user_id == user_id,
                SRReview.reviewed_at >= datetime.combine(start, datetime.min.time()),
                SRReview.reviewed_at < datetime.combine(end, datetime.min.time()),
            )
            .group_by("review_date")
            .all()
        )

        return {row.review_date: row.review_count for row in results}

    def _calculate_review_stats(self, reviews: List[SRReview]) -> Dict[str, object]:
        total_reviews = len(reviews)
        if total_reviews == 0:
            return {
                "total_reviews": 0,
                "average_quality": 0.0,
                "quality_distribution": {i: 0 for i in range(6)},
                "retention_rate": 0.0,
            }

        quality_distribution = {i: 0 for i in range(6)}
        total_quality = 0
        retained = 0

        for review in reviews:
            score = int(review.quality)
            quality_distribution[score] = quality_distribution.get(score, 0) + 1
            total_quality += score
            if score >= 3:
                retained += 1

        average_quality = total_quality / total_reviews
        retention_rate = retained / total_reviews if total_reviews else 0.0

        return {
            "total_reviews": total_reviews,
            "average_quality": round(average_quality, 2),
            "quality_distribution": quality_distribution,
            "retention_rate": round(retention_rate, 2),
        }
