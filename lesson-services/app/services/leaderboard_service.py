from datetime import date, datetime, timedelta
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy import desc, func
from sqlalchemy.orm import Session

from app.models.progress_models import LeaderboardSnapshot, UserPoints
from app.schemas.leaderboard_schema import (
    LeaderboardEntry,
    LeaderboardPeriod,
    LeaderboardResponse,
    LeaderboardSnapshotCreate,
)


class LeaderboardService:
    def __init__(self, db: Session):
        self.db = db

    # ------------------------------------------------------------------
    # Helper methods
    # ------------------------------------------------------------------
    def _calculate_week_key(self, value: date) -> str:
        iso_year, iso_week, _ = value.isocalendar()
        return f"{iso_year}-W{iso_week:02d}"

    def _calculate_month_key(self, value: date) -> str:
        return value.strftime("%Y-%m")

    def _build_response(
        self,
        period: LeaderboardPeriod,
        period_key: str,
        rows: List[LeaderboardSnapshot],
    ) -> Optional[LeaderboardResponse]:
        if not rows:
            return None

        entries = [
            LeaderboardEntry(rank=row.rank, user_id=row.user_id, points=row.points)
            for row in rows
        ]
        taken_at = max(row.taken_at for row in rows)

        return LeaderboardResponse(
            period=period,
            period_key=period_key,
            entries=entries,
            taken_at=taken_at,
        )

    def _get_leaderboard(
        self,
        period: LeaderboardPeriod,
        period_key: str,
        limit: Optional[int] = None,
        offset: int = 0,
    ) -> Optional[LeaderboardResponse]:
        query = (
            self.db.query(LeaderboardSnapshot)
            .filter(
                LeaderboardSnapshot.period == period.value,
                LeaderboardSnapshot.period_key == period_key,
            )
            .order_by(LeaderboardSnapshot.rank.asc())
        )

        if offset:
            query = query.offset(offset)
        if limit:
            query = query.limit(limit)

        rows = query.all()
        return self._build_response(period, period_key, rows)

    def _get_period_history(
        self,
        period: LeaderboardPeriod,
        limit: int,
        offset: int,
    ) -> List[LeaderboardResponse]:
        period_max = func.max(LeaderboardSnapshot.taken_at)
        rows = (
            self.db.query(
                LeaderboardSnapshot.period_key,
                period_max.label("taken_at"),
            )
            .filter(LeaderboardSnapshot.period == period.value)
            .group_by(LeaderboardSnapshot.period_key)
            .order_by(period_max.desc())
            .offset(offset)
            .limit(limit)
            .all()
        )

        responses: List[LeaderboardResponse] = []
        for row in rows:
            response = self._get_leaderboard(period, row.period_key)
            if response:
                responses.append(response)
        return responses

    # ------------------------------------------------------------------
    # Public API methods
    # ------------------------------------------------------------------
    def get_current_weekly_leaderboard(
        self,
        limit: int = 100,
        offset: int = 0,
    ) -> Optional[LeaderboardResponse]:
        current_key = self._calculate_week_key(datetime.utcnow().date())
        return self._get_leaderboard(LeaderboardPeriod.WEEKLY, current_key, limit, offset)

    def get_current_monthly_leaderboard(
        self,
        limit: int = 100,
        offset: int = 0,
    ) -> Optional[LeaderboardResponse]:
        current_key = self._calculate_month_key(datetime.utcnow().date())
        return self._get_leaderboard(LeaderboardPeriod.MONTHLY, current_key, limit, offset)

    def get_weekly_history(
        self,
        limit: int = 4,
        offset: int = 0,
    ) -> List[LeaderboardResponse]:
        return self._get_period_history(LeaderboardPeriod.WEEKLY, limit, offset)

    def get_monthly_history(
        self,
        limit: int = 6,
        offset: int = 0,
    ) -> List[LeaderboardResponse]:
        return self._get_period_history(LeaderboardPeriod.MONTHLY, limit, offset)

    def create_snapshot(
        self,
        period: LeaderboardPeriod,
        payload: LeaderboardSnapshotCreate,
    ) -> int:
        taken_at = payload.taken_at or datetime.utcnow()
        entries = [
            LeaderboardSnapshot(
                period=period.value,
                period_key=payload.period_key,
                rank=entry.rank,
                user_id=entry.user_id,
                points=entry.points,
                taken_at=taken_at,
            )
            for entry in payload.entries
        ]

        if not entries:
            return 0

        self.db.add_all(entries)
        self.db.commit()
        return len(entries)

    def get_user_leaderboard_history(
        self, user_id: UUID
    ) -> Dict[str, List[LeaderboardResponse]]:
        period_max = func.max(LeaderboardSnapshot.taken_at)
        rows = (
            self.db.query(
                LeaderboardSnapshot.period,
                LeaderboardSnapshot.period_key,
                period_max.label("taken_at"),
            )
            .filter(LeaderboardSnapshot.user_id == user_id)
            .group_by(LeaderboardSnapshot.period, LeaderboardSnapshot.period_key)
            .order_by(period_max.desc())
            .all()
        )

        history_map = {
            LeaderboardPeriod.WEEKLY.value: [],
            LeaderboardPeriod.MONTHLY.value: [],
        }

        for row in rows:
            period_enum = LeaderboardPeriod(row.period)
            response = self._get_leaderboard(period_enum, row.period_key)
            if response:
                history_map[row.period].append(response)

        return {
            "weekly": history_map[LeaderboardPeriod.WEEKLY.value],
            "monthly": history_map[LeaderboardPeriod.MONTHLY.value],
        }

    def get_leaderboard_by_week(
        self, week_key: str, limit: Optional[int] = None, offset: int = 0
    ) -> Optional[LeaderboardResponse]:
        return self._get_leaderboard(
            LeaderboardPeriod.WEEKLY, week_key, limit=limit, offset=offset
        )

    def get_leaderboard_by_month(
        self, month_key: str, limit: Optional[int] = None, offset: int = 0
    ) -> Optional[LeaderboardResponse]:
        return self._get_leaderboard(
            LeaderboardPeriod.MONTHLY, month_key, limit=limit, offset=offset
        )

    def get_user_current_ranks(self, user_id: UUID) -> Dict[str, Optional[int]]:
        today = datetime.utcnow().date()
        week_key = self._calculate_week_key(today)
        month_key = self._calculate_month_key(today)

        weekly = (
            self.db.query(LeaderboardSnapshot)
            .filter(
                LeaderboardSnapshot.period == LeaderboardPeriod.WEEKLY.value,
                LeaderboardSnapshot.period_key == week_key,
                LeaderboardSnapshot.user_id == user_id,
            )
            .order_by(desc(LeaderboardSnapshot.taken_at))
            .first()
        )

        monthly = (
            self.db.query(LeaderboardSnapshot)
            .filter(
                LeaderboardSnapshot.period == LeaderboardPeriod.MONTHLY.value,
                LeaderboardSnapshot.period_key == month_key,
                LeaderboardSnapshot.user_id == user_id,
            )
            .order_by(desc(LeaderboardSnapshot.taken_at))
            .first()
        )

        return {
            "weekly_rank": weekly.rank if weekly else None,
            "monthly_rank": monthly.rank if monthly else None,
        }

    def cleanup_old_snapshots(
        self, keep_weeks: int = 52, keep_months: int = 24
    ) -> int:
        if keep_weeks < 0 or keep_months < 0:
            raise ValueError("Retention periods must be non-negative")

        cutoff_week = datetime.utcnow() - timedelta(weeks=keep_weeks)
        cutoff_month = datetime.utcnow() - timedelta(days=keep_months * 30)

        weekly_deleted = (
            self.db.query(LeaderboardSnapshot)
            .filter(
                LeaderboardSnapshot.period == LeaderboardPeriod.WEEKLY.value,
                LeaderboardSnapshot.taken_at < cutoff_week,
            )
            .delete(synchronize_session=False)
        )

        monthly_deleted = (
            self.db.query(LeaderboardSnapshot)
            .filter(
                LeaderboardSnapshot.period == LeaderboardPeriod.MONTHLY.value,
                LeaderboardSnapshot.taken_at < cutoff_month,
            )
            .delete(synchronize_session=False)
        )

        self.db.commit()
        return (weekly_deleted or 0) + (monthly_deleted or 0)

    def create_snapshot_from_points(
        self,
        period: LeaderboardPeriod,
        limit: int = 100,
    ) -> int:
        """Generate leaderboard snapshot using the user_points table."""

        taken_at = datetime.utcnow()
        period_key = (
            self._calculate_week_key(taken_at.date())
            if period == LeaderboardPeriod.WEEKLY
            else self._calculate_month_key(taken_at.date())
        )

        order_column = (
            UserPoints.weekly
            if period == LeaderboardPeriod.WEEKLY
            else UserPoints.monthly
        )

        points_rows = (
            self.db.query(UserPoints.user_id, order_column.label("points"))
            .order_by(order_column.desc())
            .limit(limit)
            .all()
        )

        if not points_rows:
            return 0

        snapshots = [
            LeaderboardSnapshot(
                period=period.value,
                period_key=period_key,
                rank=index + 1,
                user_id=row.user_id,
                points=row.points,
                taken_at=taken_at,
            )
            for index, row in enumerate(points_rows)
        ]

        self.db.add_all(snapshots)
        self.db.commit()
        return len(snapshots)

