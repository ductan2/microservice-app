from __future__ import annotations

from datetime import datetime
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy import desc, func
from sqlalchemy.orm import Session

from app.models.progress_models import DimUser, UserPoints


class UserPointsService:
    """Encapsulates point accounting logic for users."""

    def __init__(self, db: Session):
        self.db = db

    def get_user_points(self, user_id: UUID) -> Optional[UserPoints]:
        return (
            self.db.query(UserPoints)
            .filter(UserPoints.user_id == user_id)
            .one_or_none()
        )

    def initialize_user_points(self, user_id: UUID) -> UserPoints:
        points = UserPoints(
            user_id=user_id,
            lifetime=0,
            weekly=0,
            monthly=0,
            updated_at=datetime.utcnow(),
        )
        self.db.add(points)
        self.db.commit()
        self.db.refresh(points)
        return points

    def get_or_create_points(self, user_id: UUID) -> UserPoints:
        points = self.get_user_points(user_id)
        if points:
            return points
        return self.initialize_user_points(user_id)

    def _apply_delta(self, user_id: UUID, delta: int) -> UserPoints:
        points = self.get_or_create_points(user_id)
        points.lifetime = max(points.lifetime + delta, 0)
        points.weekly = max(points.weekly + delta, 0)
        points.monthly = max(points.monthly + delta, 0)
        points.updated_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(points)
        return points

    def add_points(self, user_id: UUID, points: int) -> UserPoints:
        if points < 0:
            raise ValueError("points must be non-negative")
        return self._apply_delta(user_id, points)

    def subtract_points(self, user_id: UUID, points: int) -> UserPoints:
        if points < 0:
            raise ValueError("points must be non-negative")
        return self._apply_delta(user_id, -points)

    def _leaderboard_query(self, column):
        return self.db.query(UserPoints).order_by(desc(column), UserPoints.updated_at.asc())

    def get_lifetime_leaderboard(
        self, limit: int = 100, offset: int = 0
    ) -> List[UserPoints]:
        return (
            self._leaderboard_query(UserPoints.lifetime)
            .offset(offset)
            .limit(limit)
            .all()
        )

    def get_weekly_leaderboard(
        self, limit: int = 100, offset: int = 0
    ) -> List[UserPoints]:
        return (
            self._leaderboard_query(UserPoints.weekly)
            .offset(offset)
            .limit(limit)
            .all()
        )

    def get_monthly_leaderboard(
        self, limit: int = 100, offset: int = 0
    ) -> List[UserPoints]:
        return (
            self._leaderboard_query(UserPoints.monthly)
            .offset(offset)
            .limit(limit)
            .all()
        )

    def reset_weekly_points(self) -> int:
        updated = self.db.query(UserPoints).update(
            {
                UserPoints.weekly: 0,
                UserPoints.updated_at: datetime.utcnow(),
            },
            synchronize_session=False,
        )
        self.db.commit()
        return updated

    def reset_monthly_points(self) -> int:
        updated = self.db.query(UserPoints).update(
            {
                UserPoints.monthly: 0,
                UserPoints.updated_at: datetime.utcnow(),
            },
            synchronize_session=False,
        )
        self.db.commit()
        return updated

    def _calculate_rank(self, column, value: int) -> int:
        better_count = (
            self.db.query(func.count(UserPoints.user_id))
            .filter(column > value)
            .scalar()
            or 0
        )
        return int(better_count) + 1

    def get_user_ranks(self, user_id: UUID) -> Optional[Dict[str, int]]:
        points = self.get_user_points(user_id)
        if not points:
            return None

        lifetime_rank = self._calculate_rank(UserPoints.lifetime, points.lifetime)
        weekly_rank = self._calculate_rank(UserPoints.weekly, points.weekly)
        monthly_rank = self._calculate_rank(UserPoints.monthly, points.monthly)

        return {
            "lifetime_rank": lifetime_rank,
            "weekly_rank": weekly_rank,
            "monthly_rank": monthly_rank,
        }

    def get_top_users_with_details(
        self, period: str, limit: int = 10
    ) -> List[Dict[str, object]]:
        period_map = {
            "lifetime": UserPoints.lifetime,
            "weekly": UserPoints.weekly,
            "monthly": UserPoints.monthly,
        }
        column = period_map.get(period.lower())
        if column is None:
            raise ValueError("invalid leaderboard period")

        rows = (
            self.db.query(UserPoints, DimUser)
            .outerjoin(DimUser, DimUser.user_id == UserPoints.user_id)
            .order_by(desc(column), UserPoints.updated_at.asc())
            .limit(limit)
            .all()
        )

        results: List[Dict[str, object]] = []
        for rank, (points, dim_user) in enumerate(rows, start=1):
            results.append(
                {
                    "rank": rank,
                    "user_id": str(points.user_id),
                    "points": getattr(points, column.key),
                    "user": {
                        "locale": dim_user.locale if dim_user else None,
                        "level_hint": dim_user.level_hint if dim_user else None,
                    },
                }
            )
        return results
