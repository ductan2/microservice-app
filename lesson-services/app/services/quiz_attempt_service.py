from datetime import datetime
from typing import Dict, List, Optional
from uuid import UUID

from sqlalchemy import desc, func
from sqlalchemy.orm import Session, selectinload

from app.models.progress_models import QuizAnswer, QuizAttempt
from app.schemas import (
    QuizAnswerCreate,
    QuizAnswerSubmission,
    QuizAttemptCreate,
    QuizAttemptSubmit,
)


class QuizAttemptService:
    def __init__(self, db: Session):
        self.db = db

    def start_quiz(self, attempt_data: QuizAttemptCreate) -> QuizAttempt:
        last_attempt = (
            self.db.query(QuizAttempt)
            .filter(
                QuizAttempt.user_id == attempt_data.user_id,
                QuizAttempt.quiz_id == attempt_data.quiz_id,
            )
            .order_by(QuizAttempt.attempt_no.desc())
            .first()
        )

        attempt_no = (last_attempt.attempt_no + 1) if last_attempt else 1

        payload = attempt_data.model_dump()
        payload["attempt_no"] = attempt_no

        attempt = QuizAttempt(**payload)
        self.db.add(attempt)
        self.db.commit()
        self.db.refresh(attempt)
        return attempt

    def get_attempt(self, attempt_id: UUID) -> Optional[QuizAttempt]:
        return (
            self.db.query(QuizAttempt)
            .options(selectinload(QuizAttempt.answers))
            .filter(QuizAttempt.id == attempt_id)
            .one_or_none()
        )

    def submit_quiz(
        self, attempt_id: UUID, submission: QuizAttemptSubmit
    ) -> Optional[QuizAttempt]:
        attempt = self.get_attempt(attempt_id)
        if not attempt or attempt.submitted_at is not None:
            return None

        if submission.answers:
            for answer_data in submission.answers:
                answer_payload = QuizAnswerCreate(
                    attempt_id=attempt.id, **answer_data.model_dump()
                )
                self._upsert_answer(attempt, answer_payload)

        attempt.total_points = submission.total_points
        if submission.max_points is not None:
            attempt.max_points = submission.max_points

        submitted_at = submission.submitted_at or datetime.utcnow()
        attempt.submitted_at = submitted_at

        if submission.duration_ms is not None:
            attempt.duration_ms = submission.duration_ms
        elif attempt.started_at:
            duration = submitted_at - attempt.started_at
            attempt.duration_ms = int(duration.total_seconds() * 1000)

        if submission.passed is not None:
            attempt.passed = submission.passed
        else:
            max_points = attempt.max_points or 0
            attempt.passed = max_points == 0 or submission.total_points >= max_points

        self.db.commit()
        self.db.refresh(attempt)
        return attempt

    def _upsert_answer(
        self, attempt: QuizAttempt, answer_data: QuizAnswerCreate
    ) -> QuizAnswer:
        existing = (
            self.db.query(QuizAnswer)
            .filter(
                QuizAnswer.attempt_id == attempt.id,
                QuizAnswer.question_id == answer_data.question_id,
            )
            .one_or_none()
        )

        payload = answer_data.model_dump()
        payload["attempt_id"] = attempt.id

        if existing:
            for field, value in payload.items():
                setattr(existing, field, value)
            existing.answered_at = datetime.utcnow()
            obj = existing
        else:
            obj = QuizAnswer(**payload)
            obj.answered_at = datetime.utcnow()
            self.db.add(obj)

        return obj

    def get_user_quiz_attempts(
        self, user_id: UUID, quiz_id: UUID
    ) -> List[QuizAttempt]:
        return (
            self.db.query(QuizAttempt)
            .filter(
                QuizAttempt.user_id == user_id,
                QuizAttempt.quiz_id == quiz_id,
            )
            .order_by(desc(QuizAttempt.started_at))
            .all()
        )

    def get_user_quiz_history(
        self,
        user_id: UUID,
        passed: Optional[bool] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> List[QuizAttempt]:
        query = self.db.query(QuizAttempt).filter(QuizAttempt.user_id == user_id)

        if passed is not None:
            query = query.filter(QuizAttempt.passed == passed)

        return (
            query.order_by(desc(QuizAttempt.submitted_at), desc(QuizAttempt.started_at))
            .offset(offset)
            .limit(limit)
            .all()
        )

    def get_lesson_quiz_attempts(
        self, lesson_id: UUID, user_id: UUID
    ) -> List[QuizAttempt]:
        return (
            self.db.query(QuizAttempt)
            .filter(
                QuizAttempt.lesson_id == lesson_id,
                QuizAttempt.user_id == user_id,
            )
            .order_by(desc(QuizAttempt.started_at))
            .all()
        )

    def delete_attempt(self, attempt_id: UUID) -> bool:
        attempt = self.db.query(QuizAttempt).filter(QuizAttempt.id == attempt_id).one_or_none()
        if not attempt:
            return False

        self.db.delete(attempt)
        self.db.commit()
        return True

    def get_quiz_statistics(self, user_id: UUID, quiz_id: UUID) -> Dict[str, Optional[float]]:
        base_query = self.db.query(QuizAttempt).filter(
            QuizAttempt.user_id == user_id,
            QuizAttempt.quiz_id == quiz_id,
        )

        total_attempts = base_query.count()
        passed_attempts = base_query.filter(QuizAttempt.passed.is_(True)).count()

        average_score = (
            self.db.query(func.avg(QuizAttempt.total_points))
            .filter(
                QuizAttempt.user_id == user_id,
                QuizAttempt.quiz_id == quiz_id,
            )
            .scalar()
        )

        best_score = (
            self.db.query(func.max(QuizAttempt.total_points))
            .filter(
                QuizAttempt.user_id == user_id,
                QuizAttempt.quiz_id == quiz_id,
            )
            .scalar()
        )

        latest_attempt = (
            self.db.query(func.max(QuizAttempt.submitted_at))
            .filter(
                QuizAttempt.user_id == user_id,
                QuizAttempt.quiz_id == quiz_id,
            )
            .scalar()
        )

        pass_rate = (
            (passed_attempts / total_attempts) * 100 if total_attempts else 0.0
        )

        return {
            "total_attempts": total_attempts,
            "passed_attempts": passed_attempts,
            "pass_rate": round(pass_rate, 2),
            "average_score": float(average_score) if average_score is not None else None,
            "best_score": int(best_score) if best_score is not None else None,
            "latest_attempt_at": latest_attempt,
        }

    def get_best_attempt(self, user_id: UUID, quiz_id: UUID) -> Optional[QuizAttempt]:
        return (
            self.db.query(QuizAttempt)
            .filter(
                QuizAttempt.user_id == user_id,
                QuizAttempt.quiz_id == quiz_id,
            )
            .order_by(desc(QuizAttempt.total_points), QuizAttempt.submitted_at.asc())
            .first()
        )

