from datetime import date, timedelta
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy.orm import Session

from app.models.progress_models import DailyActivity


class DailyActivityService:
    VALID_FIELDS = {
        "lessons_completed",
        "quizzes_completed",
        "minutes",
        "points",
    }

    def __init__(self, db: Session):
        self.db = db

    def _get_or_create_activity(self, user_id: UUID, activity_dt: date) -> DailyActivity:
        activity = (
            self.db.query(DailyActivity)
            .filter(
                DailyActivity.user_id == user_id,
                DailyActivity.activity_dt == activity_dt,
            )
            .one_or_none()
        )

        if activity is None:
            activity = DailyActivity(user_id=user_id, activity_dt=activity_dt)
            self.db.add(activity)
            self.db.flush()

        return activity

    def get_today_activity(self, user_id: UUID) -> Optional[DailyActivity]:
        return self.get_activity_by_date(user_id, date.today())

    def get_activity_by_date(self, user_id: UUID, activity_date: date) -> Optional[DailyActivity]:
        return (
            self.db.query(DailyActivity)
            .filter(
                DailyActivity.user_id == user_id,
                DailyActivity.activity_dt == activity_date,
            )
            .one_or_none()
        )

    def get_activity_range(
        self, user_id: UUID, date_from: date, date_to: date
    ) -> List[DailyActivity]:
        return (
            self.db.query(DailyActivity)
            .filter(
                DailyActivity.user_id == user_id,
                DailyActivity.activity_dt >= date_from,
                DailyActivity.activity_dt <= date_to,
            )
            .order_by(DailyActivity.activity_dt.asc())
            .all()
        )

    def get_week_activity(self, user_id: UUID) -> List[DailyActivity]:
        today = date.today()
        start_of_week = today - timedelta(days=today.weekday())
        end_of_week = start_of_week + timedelta(days=6)
        activities = self.get_activity_range(user_id, start_of_week, end_of_week)
        activity_map = {activity.activity_dt: activity for activity in activities}

        ordered: List[DailyActivity] = []
        for i in range(7):
            current_day = start_of_week + timedelta(days=i)
            if current_day in activity_map:
                ordered.append(activity_map[current_day])
            else:
                ordered.append(
                    DailyActivity(
                        user_id=user_id,
                        activity_dt=current_day,
                        lessons_completed=0,
                        quizzes_completed=0,
                        minutes=0,
                        points=0,
                    )
                )

        return ordered

    def get_month_activity(self, user_id: UUID, year: int, month: int) -> Dict[str, object]:
        start_of_month = date(year, month, 1)
        if month == 12:
            end_of_month = date(year + 1, 1, 1) - timedelta(days=1)
        else:
            end_of_month = date(year, month + 1, 1) - timedelta(days=1)

        activities = self.get_activity_range(user_id, start_of_month, end_of_month)
        totals = self._aggregate_totals(activities)

        day_map = {activity.activity_dt: activity for activity in activities}
        days: List[DailyActivity] = []
        current_day = start_of_month
        while current_day <= end_of_month:
            activity = day_map.get(current_day)
            if activity is None:
                activity = DailyActivity(user_id=user_id, activity_dt=current_day)
            days.append(activity)
            current_day += timedelta(days=1)

        return {
            "year": year,
            "month": month,
            "totals": totals,
            "days": days,
        }

    def get_activity_summary(self, user_id: UUID) -> Dict[str, object]:
        activities = (
            self.db.query(DailyActivity)
            .filter(DailyActivity.user_id == user_id)
            .order_by(DailyActivity.activity_dt.asc())
            .all()
        )

        lifetime_totals = self._aggregate_totals(activities)

        today = date.today()
        last_7_start = today - timedelta(days=6)
        last_30_start = today - timedelta(days=29)

        last_7 = [a for a in activities if a.activity_dt >= last_7_start]
        last_30 = [a for a in activities if a.activity_dt >= last_30_start]

        last_7_totals = self._aggregate_totals(last_7)
        last_30_totals = self._aggregate_totals(last_30)

        active_days = len(activities)
        average_totals = {
            key: (round(value / active_days) if active_days else 0)
            for key, value in lifetime_totals.items()
        }

        most_active = None
        if activities:
            most_active = max(activities, key=lambda a: (a.points, a.activity_dt))

        return {
            "lifetime": lifetime_totals,
            "last_7_days": last_7_totals,
            "last_30_days": last_30_totals,
            "average_per_day": average_totals,
            "total_active_days": active_days,
            "most_active_day": most_active,
        }

    def increment_activity(
        self, user_id: UUID, activity_date: date, field: str, amount: int
    ) -> DailyActivity:
        field = field.lower()
        if field not in self.VALID_FIELDS:
            raise ValueError(f"Invalid activity field: {field}")

        activity = self._get_or_create_activity(user_id, activity_date)
        current_value = getattr(activity, field, 0)
        setattr(activity, field, current_value + amount)

        self.db.commit()
        self.db.refresh(activity)
        return activity

    def _aggregate_totals(self, activities: List[DailyActivity]) -> Dict[str, int]:
        totals = {
            "lessons_completed": 0,
            "quizzes_completed": 0,
            "minutes": 0,
            "points": 0,
        }

        for activity in activities:
            totals["lessons_completed"] += activity.lessons_completed
            totals["quizzes_completed"] += activity.quizzes_completed
            totals["minutes"] += activity.minutes
            totals["points"] += activity.points

        return totals
