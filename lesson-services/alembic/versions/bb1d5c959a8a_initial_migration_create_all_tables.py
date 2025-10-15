"""Initial migration - create all tables

Revision ID: bb1d5c959a8a
Revises: 
Create Date: 2025-10-11 15:55:36.074202

"""

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision = "bb1d5c959a8a"
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    # Create dim_users table
    op.create_table(
        "dim_users",
        sa.Column("user_id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("locale", sa.Text(), nullable=True),
        sa.Column("level_hint", sa.Text(), nullable=True),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
    )

    # Create user_lessons table
    op.create_table(
        "user_lessons",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("lesson_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("status", sa.Text(), nullable=False, server_default="in_progress"),
        sa.Column(
            "started_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.Column("completed_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("last_section_ord", sa.Integer(), nullable=True),
        sa.Column("score_total", sa.Integer(), nullable=False, server_default="0"),
        sa.CheckConstraint(
            "status IN ('in_progress','completed','abandoned')", name="status_check"
        ),
    )

    # Create quiz_attempts table
    op.create_table(
        "quiz_attempts",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("quiz_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("lesson_id", postgresql.UUID(as_uuid=True), nullable=True),
        sa.Column(
            "started_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.Column("submitted_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("duration_ms", sa.Integer(), nullable=True),
        sa.Column("total_points", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("max_points", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("passed", sa.Boolean(), nullable=True),
        sa.Column("attempt_no", sa.Integer(), nullable=False, server_default="1"),
    )

    # Create quiz_answers table
    op.create_table(
        "quiz_answers",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column(
            "attempt_id",
            postgresql.UUID(as_uuid=True),
            sa.ForeignKey("quiz_attempts.id", ondelete="CASCADE"),
            nullable=False,
        ),
        sa.Column("question_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column(
            "selected_ids",
            postgresql.ARRAY(postgresql.UUID(as_uuid=True)),
            server_default="{}",
        ),
        sa.Column("text_answer", sa.Text(), nullable=True),
        sa.Column("is_correct", sa.Boolean(), nullable=True),
        sa.Column("points_earned", sa.Integer(), nullable=False, server_default="0"),
        sa.Column(
            "answered_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
    )

    # Create sr_cards table (spaced repetition cards)
    op.create_table(
        "sr_cards",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("flashcard_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("ease_factor", sa.Float(), nullable=False, server_default="2.5"),
        sa.Column("interval_d", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("repetition", sa.Integer(), nullable=False, server_default="0"),
        sa.Column(
            "due_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.Column("suspended", sa.Boolean(), nullable=False, server_default="false"),
    )

    # Create sr_reviews table (spaced repetition reviews)
    op.create_table(
        "sr_reviews",
        sa.Column("id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("flashcard_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("quality", sa.Integer(), nullable=False),
        sa.Column("prev_interval", sa.Integer(), nullable=True),
        sa.Column("new_interval", sa.Integer(), nullable=True),
        sa.Column("new_ef", sa.Float(), nullable=True),
        sa.Column(
            "reviewed_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.CheckConstraint("quality BETWEEN 0 AND 5", name="quality_check"),
    )

    # Create daily_activity table
    op.create_table(
        "daily_activity",
        sa.Column("user_id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("activity_dt", sa.Date(), primary_key=True),
        sa.Column(
            "lessons_completed", sa.Integer(), nullable=False, server_default="0"
        ),
        sa.Column(
            "quizzes_completed", sa.Integer(), nullable=False, server_default="0"
        ),
        sa.Column("minutes", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("points", sa.Integer(), nullable=False, server_default="0"),
    )

    # Create user_streaks table
    op.create_table(
        "user_streaks",
        sa.Column("user_id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("current_len", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("longest_len", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("last_day", sa.Date(), nullable=True),
    )

    # Create user_points table
    op.create_table(
        "user_points",
        sa.Column("user_id", postgresql.UUID(as_uuid=True), primary_key=True),
        sa.Column("lifetime", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("weekly", sa.Integer(), nullable=False, server_default="0"),
        sa.Column("monthly", sa.Integer(), nullable=False, server_default="0"),
        sa.Column(
            "updated_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
    )

    # Create leaderboard_snapshots table
    op.create_table(
        "leaderboard_snapshots",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True),
        sa.Column("period", sa.Text(), nullable=False),
        sa.Column("period_key", sa.Text(), nullable=False),
        sa.Column("rank", sa.Integer(), nullable=False),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("points", sa.Integer(), nullable=False),
        sa.Column(
            "taken_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.CheckConstraint("period IN ('weekly','monthly')", name="period_check"),
    )

    # Create progress_events table
    op.create_table(
        "progress_events",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True),
        sa.Column("user_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("type", sa.Text(), nullable=False),
        sa.Column("payload", postgresql.JSONB(astext_type=sa.Text()), nullable=False),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
    )

    # Create outbox table
    op.create_table(
        "outbox",
        sa.Column("id", sa.Integer(), primary_key=True, autoincrement=True),
        sa.Column("aggregate_id", postgresql.UUID(as_uuid=True), nullable=False),
        sa.Column("topic", sa.Text(), nullable=False),
        sa.Column("type", sa.Text(), nullable=False),
        sa.Column("payload", postgresql.JSONB(astext_type=sa.Text()), nullable=False),
        sa.Column(
            "created_at",
            sa.DateTime(timezone=True),
            nullable=False,
            server_default=sa.func.now(),
        ),
        sa.Column("published_at", sa.DateTime(timezone=True), nullable=True),
    )


def downgrade() -> None:
    # Drop tables in reverse order
    op.drop_table("outbox")
    op.drop_table("progress_events")
    op.drop_table("leaderboard_snapshots")
    op.drop_table("user_points")
    op.drop_table("user_streaks")
    op.drop_table("daily_activity")
    op.drop_table("sr_reviews")
    op.drop_table("sr_cards")
    op.drop_table("quiz_answers")
    op.drop_table("quiz_attempts")
    op.drop_table("user_lessons")
    op.drop_table("dim_users")