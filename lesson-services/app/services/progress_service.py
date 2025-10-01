from typing import List, Optional
from uuid import UUID
from datetime import datetime, date, timedelta
from sqlalchemy.orm import Session
from sqlalchemy import and_, desc, func
from app.models.progress_models import (
    UserLesson, QuizAttempt, QuizAnswer, SRCard, SRReview,
    DailyActivity, UserStreak, UserPoints, LeaderboardSnapshot,
    ProgressEvent, Outbox
)
from app.schemas.progress_schema import (
    UserLessonCreate, UserLessonUpdate, UserLessonResponse,
    QuizAttemptCreate, QuizAttemptUpdate, QuizAttemptResponse,
    QuizAnswerCreate, QuizAnswerResponse,
    SRCardCreate, SRCardUpdate, SRCardResponse,
    SRReviewCreate, SRReviewResponse,
    DailyActivityCreate, DailyActivityUpdate, DailyActivityResponse,
    UserStreakCreate, UserStreakUpdate, UserStreakResponse,
    UserPointsCreate, UserPointsUpdate, UserPointsResponse,
    LeaderboardResponse, LeaderboardEntry,
    ProgressEventCreate, ProgressEventResponse,
    LessonStatus, LeaderboardPeriod
)
import math

class ProgressService:
    def __init__(self, db: Session):
        self.db = db

    # User Lesson Progress
    async def start_lesson(self, lesson_data: UserLessonCreate) -> UserLessonResponse:
        # Check if lesson already exists
        existing = self.db.query(UserLesson).filter(
            and_(UserLesson.user_id == lesson_data.user_id,
                 UserLesson.lesson_id == lesson_data.lesson_id)
        ).first()
        
        if existing:
            return UserLessonResponse.from_orm(existing)
        
        db_lesson = UserLesson(**lesson_data.dict())
        self.db.add(db_lesson)
        self.db.commit()
        self.db.refresh(db_lesson)
        
        # Create progress event
        await self._create_progress_event(
            lesson_data.user_id, 
            "LessonStarted", 
            {"lesson_id": str(lesson_data.lesson_id)}
        )
        
        return UserLessonResponse.from_orm(db_lesson)

    async def update_lesson_progress(
        self, 
        user_id: UUID, 
        lesson_id: UUID, 
        update_data: UserLessonUpdate
    ) -> Optional[UserLessonResponse]:
        db_lesson = self.db.query(UserLesson).filter(
            and_(UserLesson.user_id == user_id,
                 UserLesson.lesson_id == lesson_id)
        ).first()
        
        if not db_lesson:
            return None
        
        for field, value in update_data.dict(exclude_unset=True).items():
            setattr(db_lesson, field, value)
        
        # If completing lesson, update completion time and daily activity
        if update_data.status == LessonStatus.COMPLETED:
            db_lesson.completed_at = datetime.utcnow()
            await self._update_daily_activity(user_id, lessons_completed=1)
            await self._update_user_points(user_id, update_data.score_total or 0)
            await self._update_streak(user_id)
            
            # Create completion event
            await self._create_progress_event(
                user_id,
                "LessonCompleted",
                {
                    "lesson_id": str(lesson_id),
                    "score": update_data.score_total or 0,
                    "duration_minutes": self._calculate_lesson_duration(db_lesson)
                }
            )
        
        self.db.commit()
        self.db.refresh(db_lesson)
        return UserLessonResponse.from_orm(db_lesson)

    async def get_user_lessons(
        self, 
        user_id: UUID, 
        status: Optional[LessonStatus] = None,
        skip: int = 0,
        limit: int = 100
    ) -> List[UserLessonResponse]:
        query = self.db.query(UserLesson).filter(UserLesson.user_id == user_id)
        
        if status:
            query = query.filter(UserLesson.status == status)
        
        lessons = query.offset(skip).limit(limit).all()
        return [UserLessonResponse.from_orm(lesson) for lesson in lessons]

    # Quiz Attempts
    async def start_quiz_attempt(self, attempt_data: QuizAttemptCreate) -> QuizAttemptResponse:
        # Get next attempt number
        last_attempt = self.db.query(QuizAttempt).filter(
            and_(QuizAttempt.user_id == attempt_data.user_id,
                 QuizAttempt.quiz_id == attempt_data.quiz_id)
        ).order_by(desc(QuizAttempt.attempt_no)).first()
        
        attempt_no = (last_attempt.attempt_no + 1) if last_attempt else 1
        
        db_attempt = QuizAttempt(**attempt_data.dict(), attempt_no=attempt_no)
        self.db.add(db_attempt)
        self.db.commit()
        self.db.refresh(db_attempt)
        
        return QuizAttemptResponse.from_orm(db_attempt)

    async def submit_quiz_attempt(
        self, 
        attempt_id: UUID, 
        answers: List[QuizAnswerCreate]
    ) -> QuizAttemptResponse:
        db_attempt = self.db.query(QuizAttempt).filter(QuizAttempt.id == attempt_id).first()
        if not db_attempt:
            raise ValueError("Quiz attempt not found")
        
        # Save answers
        total_points = 0
        for answer_data in answers:
            answer_data.attempt_id = attempt_id
            db_answer = QuizAnswer(**answer_data.dict())
            self.db.add(db_answer)
            total_points += answer_data.points_earned
        
        # Update attempt
        db_attempt.submitted_at = datetime.utcnow()
        db_attempt.total_points = total_points
        db_attempt.passed = total_points >= (db_attempt.max_points * 0.7)  # 70% pass rate
        
        if db_attempt.started_at:
            duration = datetime.utcnow() - db_attempt.started_at
            db_attempt.duration_ms = int(duration.total_seconds() * 1000)
        
        self.db.commit()
        self.db.refresh(db_attempt)
        
        # Update daily activity and points
        if db_attempt.passed:
            await self._update_daily_activity(db_attempt.user_id, quizzes_completed=1)
            await self._update_user_points(db_attempt.user_id, total_points)
        
        # Create progress event
        await self._create_progress_event(
            db_attempt.user_id,
            "QuizSubmitted",
            {
                "quiz_id": str(db_attempt.quiz_id),
                "lesson_id": str(db_attempt.lesson_id) if db_attempt.lesson_id else None,
                "score": total_points,
                "max_score": db_attempt.max_points,
                "passed": db_attempt.passed,
                "attempt_no": db_attempt.attempt_no
            }
        )
        
        return QuizAttemptResponse.from_orm(db_attempt)

    # Spaced Repetition
    async def create_sr_card(self, card_data: SRCardCreate) -> SRCardResponse:
        # Check if card already exists
        existing = self.db.query(SRCard).filter(
            and_(SRCard.user_id == card_data.user_id,
                 SRCard.flashcard_id == card_data.flashcard_id)
        ).first()
        
        if existing:
            return SRCardResponse.from_orm(existing)
        
        db_card = SRCard(**card_data.dict())
        self.db.add(db_card)
        self.db.commit()
        self.db.refresh(db_card)
        
        return SRCardResponse.from_orm(db_card)

    async def review_flashcard(
        self, 
        user_id: UUID, 
        flashcard_id: UUID, 
        quality: int
    ) -> SRCardResponse:
        # Get the card
        db_card = self.db.query(SRCard).filter(
            and_(SRCard.user_id == user_id,
                 SRCard.flashcard_id == flashcard_id)
        ).first()
        
        if not db_card:
            raise ValueError("SR Card not found")
        
        # SM-2 Algorithm
        prev_interval = db_card.interval_d
        prev_ef = db_card.ease_factor
        
        # Calculate new values
        if quality < 3:
            # Failed review - reset
            db_card.repetition = 0
            db_card.interval_d = 0
            db_card.due_at = datetime.utcnow()
        else:
            # Successful review
            if db_card.repetition == 0:
                db_card.interval_d = 1
            elif db_card.repetition == 1:
                db_card.interval_d = 6
            else:
                db_card.interval_d = int(db_card.interval_d * db_card.ease_factor)
            
            db_card.repetition += 1
            db_card.due_at = datetime.utcnow() + timedelta(days=db_card.interval_d)
        
        # Update ease factor
        db_card.ease_factor = max(1.3, prev_ef + (0.1 - (5 - quality) * (0.08 + (5 - quality) * 0.02)))
        
        # Save review record
        review = SRReview(
            user_id=user_id,
            flashcard_id=flashcard_id,
            quality=quality,
            prev_interval=prev_interval,
            new_interval=db_card.interval_d,
            new_ef=db_card.ease_factor
        )
        self.db.add(review)
        
        self.db.commit()
        self.db.refresh(db_card)
        
        # Create progress event
        await self._create_progress_event(
            user_id,
            "SRReviewed",
            {
                "flashcard_id": str(flashcard_id),
                "quality": quality,
                "new_interval": db_card.interval_d,
                "due_at": db_card.due_at.isoformat()
            }
        )
        
        return SRCardResponse.from_orm(db_card)

    async def get_due_cards(self, user_id: UUID, limit: int = 50) -> List[SRCardResponse]:
        cards = self.db.query(SRCard).filter(
            and_(SRCard.user_id == user_id,
                 SRCard.due_at <= datetime.utcnow(),
                 SRCard.suspended == False)
        ).order_by(SRCard.due_at).limit(limit).all()
        
        return [SRCardResponse.from_orm(card) for card in cards]

    # Leaderboards
    async def get_leaderboard(
        self, 
        period: LeaderboardPeriod, 
        period_key: str,
        limit: int = 100
    ) -> Optional[LeaderboardResponse]:
        entries = self.db.query(LeaderboardSnapshot).filter(
            and_(LeaderboardSnapshot.period == period,
                 LeaderboardSnapshot.period_key == period_key)
        ).order_by(LeaderboardSnapshot.rank).limit(limit).all()
        
        if not entries:
            return None
        
        leaderboard_entries = [
            LeaderboardEntry(
                rank=entry.rank,
                user_id=entry.user_id,
                points=entry.points
            ) for entry in entries
        ]
        
        return LeaderboardResponse(
            period=period,
            period_key=period_key,
            entries=leaderboard_entries,
            taken_at=entries[0].taken_at
        )

    async def get_user_stats(self, user_id: UUID) -> dict:
        # Get user points
        points = self.db.query(UserPoints).filter(UserPoints.user_id == user_id).first()
        
        # Get streak
        streak = self.db.query(UserStreak).filter(UserStreak.user_id == user_id).first()
        
        # Get recent activity
        recent_activity = self.db.query(DailyActivity).filter(
            DailyActivity.user_id == user_id
        ).order_by(desc(DailyActivity.activity_dt)).limit(7).all()
        
        # Get lesson progress
        lesson_stats = self.db.query(
            UserLesson.status,
            func.count(UserLesson.id).label('count')
        ).filter(UserLesson.user_id == user_id).group_by(UserLesson.status).all()
        
        return {
            "points": UserPointsResponse.from_orm(points) if points else None,
            "streak": UserStreakResponse.from_orm(streak) if streak else None,
            "recent_activity": [DailyActivityResponse.from_orm(activity) for activity in recent_activity],
            "lesson_stats": {status: count for status, count in lesson_stats}
        }

    # Helper methods
    async def _create_progress_event(self, user_id: UUID, event_type: str, payload: dict):
        event = ProgressEvent(
            user_id=user_id,
            type=event_type,
            payload=payload
        )
        self.db.add(event)
        
        # Also add to outbox for event publishing
        outbox_event = Outbox(
            aggregate_id=user_id,
            topic="progress.events",
            type=event_type,
            payload=payload
        )
        self.db.add(outbox_event)

    async def _update_daily_activity(
        self, 
        user_id: UUID, 
        lessons_completed: int = 0,
        quizzes_completed: int = 0,
        minutes: int = 0,
        points: int = 0
    ):
        today = date.today()
        activity = self.db.query(DailyActivity).filter(
            and_(DailyActivity.user_id == user_id,
                 DailyActivity.activity_dt == today)
        ).first()
        
        if not activity:
            activity = DailyActivity(
                user_id=user_id,
                activity_dt=today,
                lessons_completed=lessons_completed,
                quizzes_completed=quizzes_completed,
                minutes=minutes,
                points=points
            )
            self.db.add(activity)
        else:
            activity.lessons_completed += lessons_completed
            activity.quizzes_completed += quizzes_completed
            activity.minutes += minutes
            activity.points += points

    async def _update_user_points(self, user_id: UUID, points: int):
        user_points = self.db.query(UserPoints).filter(UserPoints.user_id == user_id).first()
        
        if not user_points:
            user_points = UserPoints(
                user_id=user_id,
                lifetime=points,
                weekly=points,
                monthly=points
            )
            self.db.add(user_points)
        else:
            user_points.lifetime += points
            user_points.weekly += points
            user_points.monthly += points
            user_points.updated_at = datetime.utcnow()

    async def _update_streak(self, user_id: UUID):
        today = date.today()
        yesterday = today - timedelta(days=1)
        
        streak = self.db.query(UserStreak).filter(UserStreak.user_id == user_id).first()
        
        if not streak:
            streak = UserStreak(
                user_id=user_id,
                current_len=1,
                longest_len=1,
                last_day=today
            )
            self.db.add(streak)
        else:
            if streak.last_day == yesterday:
                # Continuing streak
                streak.current_len += 1
                streak.longest_len = max(streak.longest_len, streak.current_len)
            elif streak.last_day != today:
                # New streak or broken streak
                streak.current_len = 1
            
            streak.last_day = today

    def _calculate_lesson_duration(self, lesson: UserLesson) -> int:
        if lesson.completed_at and lesson.started_at:
            delta = lesson.completed_at - lesson.started_at
            return int(delta.total_seconds() / 60)  # minutes
        return 0