from sqlalchemy import and_, desc, func
from sqlalchemy.orm import Session
from typing import List, Optional, Dict, Any
from uuid import UUID
from datetime import datetime, date, time, timedelta

from app.models.progress_models import ProgressEvent
from app.schemas import ProgressEventCreate


class ProgressEventService:
    def __init__(self, db: Session):
        self.db = db

    def _apply_date_filters(
        self,
        query,
        date_from: Optional[date],
        date_to: Optional[date],
    ):
        if date_from:
            start_dt = datetime.combine(date_from, time.min)
            query = query.filter(ProgressEvent.created_at >= start_dt)
        if date_to:
            end_dt = datetime.combine(date_to, time.max)
            query = query.filter(ProgressEvent.created_at <= end_dt)
        return query

    def get_user_events(
        self,
        user_id: UUID,
        event_type: Optional[str] = None,
        limit: int = 100,
        offset: int = 0,
        date_from: Optional[date] = None,
        date_to: Optional[date] = None,
    ) -> List[ProgressEvent]:
        query = self.db.query(ProgressEvent).filter(ProgressEvent.user_id == user_id)

        if event_type:
            query = query.filter(ProgressEvent.type == event_type)

        query = self._apply_date_filters(query, date_from, date_to)

        return (
            query.order_by(desc(ProgressEvent.created_at))
            .offset(offset)
            .limit(limit)
            .all()
        )

    def get_event(self, event_id: int) -> Optional[ProgressEvent]:
        return (
            self.db.query(ProgressEvent)
            .filter(ProgressEvent.id == event_id)
            .one_or_none()
        )

    def create_event(self, event_data: ProgressEventCreate) -> ProgressEvent:
        new_event = ProgressEvent(**event_data.model_dump())
        self.db.add(new_event)
        self.db.commit()
        self.db.refresh(new_event)
        return new_event

    def get_events_by_type(
        self,
        user_id: UUID,
        event_type: str,
        limit: int = 50,
        offset: int = 0,
    ) -> List[ProgressEvent]:
        return self.get_user_events(
            user_id=user_id,
            event_type=event_type,
            limit=limit,
            offset=offset,
        )

    def get_recent_events(self, user_id: UUID, limit: int = 50) -> List[ProgressEvent]:
        return (
            self.db.query(ProgressEvent)
            .filter(ProgressEvent.user_id == user_id)
            .order_by(desc(ProgressEvent.created_at))
            .limit(limit)
            .all()
        )

    def delete_event(self, event_id: int) -> bool:
        event = self.get_event(event_id)
        if not event:
            return False

        self.db.delete(event)
        self.db.commit()
        return True

    def get_event_type_stats(
        self,
        date_from: Optional[date] = None,
        date_to: Optional[date] = None,
    ) -> Dict[str, int]:
        query = self.db.query(
            ProgressEvent.type,
            func.count(ProgressEvent.id).label("count"),
        )

        if date_from or date_to:
            query = self._apply_date_filters(query, date_from, date_to)

        results = (
            query.group_by(ProgressEvent.type)
            .order_by(ProgressEvent.type.asc())
            .all()
        )

        return {row.type: row.count for row in results}

    def get_user_event_timeline(
        self,
        user_id: UUID,
        date_from: date,
        date_to: date,
    ) -> List[ProgressEvent]:
        query = self.db.query(ProgressEvent).filter(ProgressEvent.user_id == user_id)
        query = self._apply_date_filters(query, date_from, date_to)
        return query.order_by(ProgressEvent.created_at.asc()).all()

    def bulk_create_events(self, events: List[Dict[str, Any]]) -> List[ProgressEvent]:
        new_events = [ProgressEvent(**event) for event in events]
        self.db.add_all(new_events)
        self.db.commit()
        # Refresh objects individually to populate generated fields
        for event in new_events:
            self.db.refresh(event)
        return new_events

    def get_event_count_by_day(
        self,
        user_id: UUID,
        days: int = 30,
    ) -> Dict[date, int]:
        cutoff = datetime.utcnow() - timedelta(days=days)
        day_expr = func.date_trunc("day", ProgressEvent.created_at)
        query = (
            self.db.query(
                day_expr.label("event_day"),
                func.count(ProgressEvent.id).label("count"),
            )
            .filter(
                and_(
                    ProgressEvent.user_id == user_id,
                    ProgressEvent.created_at >= cutoff,
                )
            )
            .group_by(day_expr)
            .order_by(day_expr)
        )

        return {row.event_day.date(): row.count for row in query.all()}

    def publish_event_to_queue(self, event: ProgressEvent) -> bool:
        """Placeholder for publishing events to a message queue."""
        # Queue infrastructure is not available in this codebase yet. Return False
        # to indicate that the event was not published but the operation succeeded.
        return False

