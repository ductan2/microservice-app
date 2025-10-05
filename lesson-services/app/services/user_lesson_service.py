from __future__ import annotations

from datetime import datetime
from typing import List, Optional
from uuid import UUID

from sqlalchemy import func
from sqlalchemy.orm import Session

from app.models.progress_models import UserLesson
from app.schemas import (
    LessonStatus,
    UserLessonCompletionRequest,
    UserLessonCreate,
    UserLessonStats,
    UserLessonUpdate,
)


class UserLessonService:
    """Business logic for managing user lesson progress."""

    def __init__(self, db: Session):
        self.db = db

    def _status_value(self, status: LessonStatus | str | None) -> Optional[str]:
        if status is None:
            return None
        if isinstance(status, LessonStatus):
            return status.value
        return status

    def _query(self):
        return self.db.query(UserLesson)

    def get_user_lessons(
        self, user_id: UUID, status: Optional[LessonStatus] = None
    ) -> List[UserLesson]:
        """Return all lessons for a user with optional status filter."""

        query = self._query().filter(UserLesson.user_id == user_id)
        status_value = self._status_value(status)
        if status_value:
            query = query.filter(UserLesson.status == status_value)
        return query.order_by(UserLesson.started_at.desc()).all()

    def get_user_lesson(self, user_id: UUID, lesson_id: UUID) -> Optional[UserLesson]:
        """Return the most recent lesson entry for the user/lesson pair."""

        return (
            self._query()
            .filter(
                UserLesson.user_id == user_id,
                UserLesson.lesson_id == lesson_id,
            )
            .order_by(UserLesson.started_at.desc())
            .first()
        )

    def start_lesson(self, payload: UserLessonCreate) -> UserLesson:
        """Start a lesson for a user or return the active attempt."""

        existing = self.get_user_lesson(payload.user_id, payload.lesson_id)
        if existing and existing.status == LessonStatus.IN_PROGRESS.value:
            return existing

        data = payload.model_dump()
        data["status"] = LessonStatus.IN_PROGRESS.value
        data["score_total"] = data.get("score_total") or 0

        lesson = UserLesson(**data)
        self.db.add(lesson)
        self.db.commit()
        self.db.refresh(lesson)
        return lesson

    def update_progress(
        self, user_id: UUID, lesson_id: UUID, update: UserLessonUpdate
    ) -> Optional[UserLesson]:
        """Update lesson progress fields for a user."""

        lesson = self.get_user_lesson(user_id, lesson_id)
        if not lesson:
            return None

        payload = update.model_dump(exclude_unset=True)

        if "status" in payload and payload["status"] is not None:
            lesson.status = self._status_value(payload["status"]) or lesson.status

        if "last_section_ord" in payload:
            lesson.last_section_ord = payload["last_section_ord"]

        if "score_total" in payload and payload["score_total"] is not None:
            lesson.score_total = payload["score_total"]

        if "completed_at" in payload:
            lesson.completed_at = payload["completed_at"]

        if (
            lesson.status == LessonStatus.COMPLETED.value
            and lesson.completed_at is None
        ):
            lesson.completed_at = datetime.utcnow()

        self.db.commit()
        self.db.refresh(lesson)
        return lesson

    def complete_lesson(
        self,
        user_id: UUID,
        lesson_id: UUID,
        completion: UserLessonCompletionRequest,
    ) -> Optional[UserLesson]:
        """Mark a lesson as completed for a user."""

        lesson = self.get_user_lesson(user_id, lesson_id)
        if not lesson:
            return None

        if completion.last_section_ord is not None:
            lesson.last_section_ord = completion.last_section_ord
        if completion.score_total is not None:
            lesson.score_total = completion.score_total

        lesson.status = LessonStatus.COMPLETED.value
        lesson.completed_at = completion.completed_at or datetime.utcnow()

        self.db.commit()
        self.db.refresh(lesson)
        return lesson

    def abandon_lesson(self, user_id: UUID, lesson_id: UUID) -> Optional[UserLesson]:
        """Mark a lesson as abandoned for the user."""

        lesson = self.get_user_lesson(user_id, lesson_id)
        if not lesson:
            return None

        lesson.status = LessonStatus.ABANDONED.value
        lesson.completed_at = None
        self.db.commit()
        self.db.refresh(lesson)
        return lesson

    def get_in_progress_lessons(self, user_id: UUID) -> List[UserLesson]:
        return self.get_user_lessons(user_id, LessonStatus.IN_PROGRESS)

    def get_completed_lessons(
        self, user_id: UUID, limit: int = 50, offset: int = 0
    ) -> List[UserLesson]:
        return (
            self._query()
            .filter(
                UserLesson.user_id == user_id,
                UserLesson.status == LessonStatus.COMPLETED.value,
            )
            .order_by(UserLesson.completed_at.desc(), UserLesson.started_at.desc())
            .offset(offset)
            .limit(limit)
            .all()
        )

    def delete_user_lesson(self, user_id: UUID, lesson_id: UUID) -> bool:
        lesson = self.get_user_lesson(user_id, lesson_id)
        if not lesson:
            return False

        self.db.delete(lesson)
        self.db.commit()
        return True

    def get_lesson_stats(self, user_id: UUID) -> UserLessonStats:
        total_started = (
            self.db.query(func.count(UserLesson.id))
            .filter(UserLesson.user_id == user_id)
            .scalar()
            or 0
        )

        in_progress = (
            self.db.query(func.count(UserLesson.id))
            .filter(
                UserLesson.user_id == user_id,
                UserLesson.status == LessonStatus.IN_PROGRESS.value,
            )
            .scalar()
            or 0
        )

        completed = (
            self.db.query(func.count(UserLesson.id))
            .filter(
                UserLesson.user_id == user_id,
                UserLesson.status == LessonStatus.COMPLETED.value,
            )
            .scalar()
            or 0
        )

        abandoned = (
            self.db.query(func.count(UserLesson.id))
            .filter(
                UserLesson.user_id == user_id,
                UserLesson.status == LessonStatus.ABANDONED.value,
            )
            .scalar()
            or 0
        )

        total_score = (
            self.db.query(func.coalesce(func.sum(UserLesson.score_total), 0))
            .filter(UserLesson.user_id == user_id)
            .scalar()
            or 0
        )

        completion_rate = (
            float(completed) / float(total_started) if total_started > 0 else 0.0
        )

        return UserLessonStats(
            total_started=total_started,
            in_progress=in_progress,
            completed=completed,
            abandoned=abandoned,
            completion_rate=completion_rate,
            total_score=total_score,
        )
