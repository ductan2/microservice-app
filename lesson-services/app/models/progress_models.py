from sqlalchemy import Column, String, Integer, Boolean, DateTime, Date, Text, CheckConstraint, ForeignKey, ARRAY, Float
from sqlalchemy.dialects.postgresql import UUID, JSONB
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
import uuid

Base = declarative_base()

class DimUser(Base):
    __tablename__ = "dim_users"
    
    user_id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    locale = Column(Text)
    level_hint = Column(Text)
    updated_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())

class UserLesson(Base):
    __tablename__ = "user_lessons"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    lesson_id = Column(UUID(as_uuid=True), nullable=False)
    status = Column(Text, nullable=False, default='in_progress')
    started_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    completed_at = Column(DateTime(timezone=True))
    last_section_ord = Column(Integer)
    score_total = Column(Integer, nullable=False, default=0)
    
    __table_args__ = (
        CheckConstraint("status IN ('in_progress','completed','abandoned')", name='status_check'),
    )

class CourseEnrollment(Base):
    __tablename__ = "course_enrollments"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    course_id = Column(UUID(as_uuid=True), nullable=False)
    status = Column(Text, nullable=False, default="enrolled")  # enrolled, in_progress, completed, cancelled
    progress_percent = Column(Integer, nullable=False, default=0)
    enrolled_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    started_at = Column(DateTime(timezone=True))
    completed_at = Column(DateTime(timezone=True))
    last_accessed_at = Column(DateTime(timezone=True))

    __table_args__ = (
        CheckConstraint("status IN ('enrolled','in_progress','completed','cancelled')", name="course_status_check"),
    )


class CourseLesson(Base):
    __tablename__ = "course_lessons"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    course_id = Column(UUID(as_uuid=True), nullable=False)
    lesson_id = Column(UUID(as_uuid=True), nullable=False)
    ord = Column(Integer, nullable=False)
    is_required = Column(Boolean, nullable=False, default=True)
    created_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())

    __table_args__ = (
        CheckConstraint("ord >= 0", name="lesson_order_nonnegative"),
    )


class QuizAttempt(Base):
    __tablename__ = "quiz_attempts"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    quiz_id = Column(UUID(as_uuid=True), nullable=False)
    lesson_id = Column(UUID(as_uuid=True))
    started_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    submitted_at = Column(DateTime(timezone=True))
    duration_ms = Column(Integer)
    total_points = Column(Integer, nullable=False, default=0)
    max_points = Column(Integer, nullable=False, default=0)
    passed = Column(Boolean)
    attempt_no = Column(Integer, nullable=False, default=1)
    
    # Relationship
    answers = relationship("QuizAnswer", back_populates="attempt", cascade="all, delete-orphan")

class QuizAnswer(Base):
    __tablename__ = "quiz_answers"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    attempt_id = Column(UUID(as_uuid=True), ForeignKey('quiz_attempts.id', ondelete='CASCADE'), nullable=False)
    question_id = Column(UUID(as_uuid=True), nullable=False)
    selected_ids = Column(ARRAY(UUID(as_uuid=True)), default=[])
    text_answer = Column(Text)
    is_correct = Column(Boolean)
    points_earned = Column(Integer, nullable=False, default=0)
    answered_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    
    # Relationship
    attempt = relationship("QuizAttempt", back_populates="answers")

class SRCard(Base):
    __tablename__ = "sr_cards"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    flashcard_id = Column(UUID(as_uuid=True), nullable=False)
    ease_factor = Column(Float, nullable=False, default=2.5)
    interval_d = Column(Integer, nullable=False, default=0)
    repetition = Column(Integer, nullable=False, default=0)
    due_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    suspended = Column(Boolean, nullable=False, default=False)

class SRReview(Base):
    __tablename__ = "sr_reviews"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    flashcard_id = Column(UUID(as_uuid=True), nullable=False)
    quality = Column(Integer, nullable=False)
    prev_interval = Column(Integer)
    new_interval = Column(Integer)
    new_ef = Column(Float)
    reviewed_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    
    __table_args__ = (
        CheckConstraint("quality BETWEEN 0 AND 5", name='quality_check'),
    )

class DailyActivity(Base):
    __tablename__ = "daily_activity"
    
    user_id = Column(UUID(as_uuid=True), primary_key=True)
    activity_dt = Column(Date, primary_key=True)
    lessons_completed = Column(Integer, nullable=False, default=0)
    quizzes_completed = Column(Integer, nullable=False, default=0)
    minutes = Column(Integer, nullable=False, default=0)
    points = Column(Integer, nullable=False, default=0)

class UserStreak(Base):
    __tablename__ = "user_streaks"
    
    user_id = Column(UUID(as_uuid=True), primary_key=True)
    current_len = Column(Integer, nullable=False, default=0)
    longest_len = Column(Integer, nullable=False, default=0)
    last_day = Column(Date)

class UserPoints(Base):
    __tablename__ = "user_points"
    
    user_id = Column(UUID(as_uuid=True), primary_key=True)
    lifetime = Column(Integer, nullable=False, default=0)
    weekly = Column(Integer, nullable=False, default=0)
    monthly = Column(Integer, nullable=False, default=0)
    updated_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())

class LeaderboardSnapshot(Base):
    __tablename__ = "leaderboard_snapshots"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    period = Column(Text, nullable=False)
    period_key = Column(Text, nullable=False)
    rank = Column(Integer, nullable=False)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    points = Column(Integer, nullable=False)
    taken_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    
    __table_args__ = (
        CheckConstraint("period IN ('weekly','monthly')", name='period_check'),
    )

class ProgressEvent(Base):
    __tablename__ = "progress_events"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    user_id = Column(UUID(as_uuid=True), nullable=False)
    type = Column(Text, nullable=False)
    payload = Column(JSONB, nullable=False)
    created_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())

class Outbox(Base):
    __tablename__ = "outbox"
    
    id = Column(Integer, primary_key=True, autoincrement=True)
    aggregate_id = Column(UUID(as_uuid=True), nullable=False)
    topic = Column(Text, nullable=False)
    type = Column(Text, nullable=False)
    payload = Column(JSONB, nullable=False)
    created_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())
    published_at = Column(DateTime(timezone=True))