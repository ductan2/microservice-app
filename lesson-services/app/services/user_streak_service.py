from __future__ import annotations

from datetime import date, timedelta
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy import desc
from sqlalchemy.orm import Session

from app.models.progress_models import DailyActivity, UserStreak


class UserStreakService:
    """Service layer for streak tracking logic."""

    def __init__(self, db: Session):
        self.db = db

    def get_user_streak(self, user_id: UUID) -> Optional[UserStreak]:
        return (
            self.db.query(UserStreak)
            .filter(UserStreak.user_id == user_id)
            .one_or_none()
        )

    def initialize_streak(self, user_id: UUID) -> UserStreak:
        streak = UserStreak(user_id=user_id, current_len=0, longest_len=0, last_day=None)
        self.db.add(streak)
        self.db.commit()
        self.db.refresh(streak)
        return streak

    def get_or_create_streak(self, user_id: UUID) -> UserStreak:
        streak = self.get_user_streak(user_id)
        if streak:
            return streak
        return self.initialize_streak(user_id)

    def _has_activity(self, user_id: UUID, activity_date: date) -> bool:
        activity = (
            self.db.query(DailyActivity)
            .filter(
                DailyActivity.user_id == user_id,
                DailyActivity.activity_dt == activity_date,
            )
            .one_or_none()
        )
        if not activity:
            return False
        return any(
            getattr(activity, field) > 0
            for field in [
                "lessons_completed",
                "quizzes_completed",
                "minutes",
                "points",
            ]
        )

    def check_and_update_streak(
        self, user_id: UUID, activity_date: Optional[date] = None
    ) -> UserStreak:
        activity_date = activity_date or date.today()
        streak = self.get_or_create_streak(user_id)

        if not self._has_activity(user_id, activity_date):
            return streak

        if streak.last_day is None:
            streak.current_len = 1
            streak.longest_len = max(streak.longest_len, streak.current_len)
            streak.last_day = activity_date
        else:
            if activity_date == streak.last_day:
                # already counted for the day
                return streak
            if activity_date == streak.last_day + timedelta(days=1):
                streak.current_len += 1
            else:
                if activity_date < streak.last_day:
                    return streak
                streak.current_len = 1
            streak.last_day = activity_date
            if streak.current_len > streak.longest_len:
                streak.longest_len = streak.current_len

        self.db.commit()
        self.db.refresh(streak)
        return streak

    def break_streak(self, user_id: UUID) -> Optional[UserStreak]:
        streak = self.get_user_streak(user_id)
        if not streak:
            return None
        streak.current_len = 0
        self.db.commit()
        self.db.refresh(streak)
        return streak

    def get_streak_status(self, user_id: UUID) -> Dict[str, object]:
        streak = self.get_or_create_streak(user_id)
        today = date.today()
        has_activity_today = self._has_activity(user_id, today)
        last_day = streak.last_day
        days_since_last: Optional[int] = None
        status = "inactive"

        if last_day is not None:
            days_since_last = (today - last_day).days
            if has_activity_today or days_since_last == 0:
                status = "active"
            elif days_since_last == 1:
                status = "at_risk"
            elif days_since_last > 1:
                status = "broken"
        elif has_activity_today:
            status = "active"

        return {
            "user_id": streak.user_id,
            "current_len": streak.current_len,
            "longest_len": streak.longest_len,
            "last_day": last_day,
            "status": status,
            "has_activity_today": has_activity_today,
            "days_since_last": days_since_last,
        }

    def get_streak_leaderboard(self, limit: int = 50) -> List[UserStreak]:
        return (
            self.db.query(UserStreak)
            .filter(UserStreak.current_len > 0)
            .order_by(desc(UserStreak.current_len), desc(UserStreak.last_day))
            .limit(limit)
            .all()
        )

    def recalculate_streak(self, user_id: UUID) -> UserStreak:
        streak = self.get_or_create_streak(user_id)
        activities = (
            self.db.query(DailyActivity)
            .filter(DailyActivity.user_id == user_id)
            .order_by(DailyActivity.activity_dt.asc())
            .all()
        )

        active_dates = [
            activity.activity_dt
            for activity in activities
            if any(
                getattr(activity, field) > 0
                for field in [
                    "lessons_completed",
                    "quizzes_completed",
                    "minutes",
                    "points",
                ]
            )
        ]

        if not active_dates:
            streak.current_len = 0
            streak.longest_len = 0
            streak.last_day = None
            self.db.commit()
            self.db.refresh(streak)
            return streak

        active_dates = sorted(set(active_dates))
        longest = 0
        current = 0
        previous_day: Optional[date] = None

        for current_day in active_dates:
            if previous_day is None or current_day == previous_day + timedelta(days=1):
                current = current + 1 if previous_day is not None else 1
            else:
                current = 1
            longest = max(longest, current)
            previous_day = current_day

        streak.longest_len = max(longest, streak.longest_len)
        streak.last_day = active_dates[-1]

        # Calculate current streak ending at the most recent activity
        current_streak = 0
        expected_day = streak.last_day
        for current_day in reversed(active_dates):
            if expected_day is None:
                break
            if current_day == expected_day - timedelta(days=current_streak):
                current_streak += 1
            else:
                break

        streak.current_len = current_streak
        streak.longest_len = max(streak.longest_len, streak.current_len)

        self.db.commit()
        self.db.refresh(streak)
        return streak
