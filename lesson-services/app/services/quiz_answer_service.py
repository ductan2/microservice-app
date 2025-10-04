from datetime import datetime
from typing import List, Optional, Tuple
from uuid import UUID

from sqlalchemy.orm import Session

from app.models.progress_models import QuizAnswer, QuizAttempt
from app.schemas.progress_schema import (
    QuizAnswerCreate,
    QuizAnswerSummary,
    QuizAnswerUpdate,
)


class QuizAnswerService:
    def __init__(self, db: Session):
        self.db = db

    def _get_attempt(self, attempt_id: UUID) -> Optional[QuizAttempt]:
        return (
            self.db.query(QuizAttempt)
            .filter(QuizAttempt.id == attempt_id)
            .one_or_none()
        )

    def _ensure_attempt_open(self, attempt_id: UUID) -> QuizAttempt:
        attempt = self._get_attempt(attempt_id)
        if not attempt:
            raise ValueError("Quiz attempt not found")
        if attempt.submitted_at is not None:
            raise ValueError("Quiz attempt has already been submitted")
        return attempt

    def get_attempt_answers(self, attempt_id: UUID) -> List[QuizAnswer]:
        return (
            self.db.query(QuizAnswer)
            .filter(QuizAnswer.attempt_id == attempt_id)
            .order_by(QuizAnswer.answered_at.asc())
            .all()
        )

    def create_answer(self, payload: QuizAnswerCreate) -> QuizAnswer:
        self._ensure_attempt_open(payload.attempt_id)

        answer = QuizAnswer(**payload.model_dump())
        answer.answered_at = datetime.utcnow()

        self.db.add(answer)
        self.db.commit()
        self.db.refresh(answer)
        return answer

    def get_answer(self, answer_id: UUID) -> Optional[QuizAnswer]:
        return (
            self.db.query(QuizAnswer)
            .filter(QuizAnswer.id == answer_id)
            .one_or_none()
        )

    def update_answer(
        self, answer_id: UUID, update: QuizAnswerUpdate
    ) -> Optional[QuizAnswer]:
        answer = self.get_answer(answer_id)
        if not answer:
            return None

        self._ensure_attempt_open(answer.attempt_id)

        for field, value in update.model_dump(exclude_unset=True).items():
            setattr(answer, field, value)

        answer.answered_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(answer)
        return answer

    def delete_answer(self, answer_id: UUID) -> bool:
        answer = self.get_answer(answer_id)
        if not answer:
            return False

        self._ensure_attempt_open(answer.attempt_id)

        self.db.delete(answer)
        self.db.commit()
        return True

    def get_answer_summary(self, attempt_id: UUID) -> QuizAnswerSummary:
        answers = self.get_attempt_answers(attempt_id)
        total = len(answers)
        correct = len([a for a in answers if a.is_correct])
        points = sum(a.points_earned for a in answers)
        accuracy = (correct / total * 100) if total else 0.0

        return QuizAnswerSummary(
            total_answers=total,
            correct_answers=correct,
            accuracy=round(accuracy, 2),
            points_earned=points,
        )

    def validate_answer(
        self,
        question_id: UUID,
        selected_ids: List[UUID],
        text_answer: Optional[str],
    ) -> Tuple[bool, int]:
        """Placeholder validation logic.

        In a production environment this method would query the content service to
        verify the supplied answer. For now we simply treat the presence of any
        selection or text as a submitted answer and defer correctness scoring to
        external systems.
        """

        is_correct = False
        points = 0
        if selected_ids or (text_answer and text_answer.strip()):
            # Mark the answer as attempted; awarding points requires external data.
            is_correct = False
            points = 0
        return is_correct, points

    def bulk_create_answers(
        self, attempt_id: UUID, answers: List[QuizAnswerCreate]
    ) -> List[QuizAnswer]:
        self._ensure_attempt_open(attempt_id)

        created_answers: List[QuizAnswer] = []
        for answer_data in answers:
            payload = answer_data.model_dump()
            payload["attempt_id"] = attempt_id
            answer = QuizAnswer(**payload)
            answer.answered_at = datetime.utcnow()
            self.db.add(answer)
            created_answers.append(answer)

        if created_answers:
            self.db.commit()
            for answer in created_answers:
                self.db.refresh(answer)

        return created_answers

