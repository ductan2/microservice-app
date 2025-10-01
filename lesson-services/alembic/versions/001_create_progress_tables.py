"""Create progress tables

Revision ID: 001_create_progress_tables
Revises: 
Create Date: 2025-01-01 00:00:00.000000

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision = '001_create_progress_tables'
down_revision = None
branch_labels = None
depends_on = None


def upgrade() -> None:
    # Enable uuid-ossp extension
    op.execute('CREATE EXTENSION IF NOT EXISTS "uuid-ossp"')
    
    # user profile mirror (for analytics join; no hard FK)
    op.create_table('dim_users',
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('locale', sa.Text(), nullable=True),
    sa.Column('level_hint', sa.Text(), nullable=True),
    sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.PrimaryKeyConstraint('user_id')
    )

    # lesson enrollment/state
    op.create_table('user_lessons',
    sa.Column('id', postgresql.UUID(as_uuid=True), server_default=sa.text('uuid_generate_v4()'), nullable=False),
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('lesson_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('status', sa.Text(), nullable=False, default='in_progress'),
    sa.Column('started_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.Column('completed_at', sa.DateTime(timezone=True), nullable=True),
    sa.Column('last_section_ord', sa.Integer(), nullable=True),
    sa.Column('score_total', sa.Integer(), nullable=False, default=0),
    sa.CheckConstraint("status IN ('in_progress','completed','abandoned')", name='status_check'),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('user_id', 'lesson_id')
    )
    op.create_index('user_lessons_user_status_idx', 'user_lessons', ['user_id', 'status'])

    # attempts
    op.create_table('quiz_attempts',
    sa.Column('id', postgresql.UUID(as_uuid=True), server_default=sa.text('uuid_generate_v4()'), nullable=False),
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('quiz_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('lesson_id', postgresql.UUID(as_uuid=True), nullable=True),
    sa.Column('started_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.Column('submitted_at', sa.DateTime(timezone=True), nullable=True),
    sa.Column('duration_ms', sa.Integer(), nullable=True),
    sa.Column('total_points', sa.Integer(), nullable=False, default=0),
    sa.Column('max_points', sa.Integer(), nullable=False, default=0),
    sa.Column('passed', sa.Boolean(), nullable=True),
    sa.Column('attempt_no', sa.Integer(), nullable=False, default=1),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_index('quiz_attempts_user_quiz_time_idx', 'quiz_attempts', ['user_id', 'quiz_id', sa.text('started_at DESC')])

    op.create_table('quiz_answers',
    sa.Column('id', postgresql.UUID(as_uuid=True), server_default=sa.text('uuid_generate_v4()'), nullable=False),
    sa.Column('attempt_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('question_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('selected_ids', postgresql.ARRAY(postgresql.UUID(as_uuid=True)), default='{}'),
    sa.Column('text_answer', sa.Text(), nullable=True),
    sa.Column('is_correct', sa.Boolean(), nullable=True),
    sa.Column('points_earned', sa.Integer(), nullable=False, default=0),
    sa.Column('answered_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.ForeignKeyConstraint(['attempt_id'], ['quiz_attempts.id'], ondelete='CASCADE'),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_index('quiz_answers_attempt_idx', 'quiz_answers', ['attempt_id'])

    # spaced repetition (SM-2 fields)
    op.create_table('sr_cards',
    sa.Column('id', postgresql.UUID(as_uuid=True), server_default=sa.text('uuid_generate_v4()'), nullable=False),
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('flashcard_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('ease_factor', sa.Float(), nullable=False, default=2.5),
    sa.Column('interval_d', sa.Integer(), nullable=False, default=0),
    sa.Column('repetition', sa.Integer(), nullable=False, default=0),
    sa.Column('due_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.Column('suspended', sa.Boolean(), nullable=False, default=False),
    sa.PrimaryKeyConstraint('id'),
    sa.UniqueConstraint('user_id', 'flashcard_id')
    )
    op.create_index('sr_due_idx', 'sr_cards', ['user_id', 'due_at'])

    op.create_table('sr_reviews',
    sa.Column('id', postgresql.UUID(as_uuid=True), server_default=sa.text('uuid_generate_v4()'), nullable=False),
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('flashcard_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('quality', sa.Integer(), nullable=False),
    sa.Column('prev_interval', sa.Integer(), nullable=True),
    sa.Column('new_interval', sa.Integer(), nullable=True),
    sa.Column('new_ef', sa.Float(), nullable=True),
    sa.Column('reviewed_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.CheckConstraint('quality BETWEEN 0 AND 5', name='quality_check'),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_index('sr_reviews_user_time_idx', 'sr_reviews', ['user_id', sa.text('reviewed_at DESC')])

    # streaks + daily counters
    op.create_table('daily_activity',
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('activity_dt', sa.Date(), nullable=False),
    sa.Column('lessons_completed', sa.Integer(), nullable=False, default=0),
    sa.Column('quizzes_completed', sa.Integer(), nullable=False, default=0),
    sa.Column('minutes', sa.Integer(), nullable=False, default=0),
    sa.Column('points', sa.Integer(), nullable=False, default=0),
    sa.PrimaryKeyConstraint('user_id', 'activity_dt')
    )

    op.create_table('user_streaks',
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('current_len', sa.Integer(), nullable=False, default=0),
    sa.Column('longest_len', sa.Integer(), nullable=False, default=0),
    sa.Column('last_day', sa.Date(), nullable=True),
    sa.PrimaryKeyConstraint('user_id')
    )

    # points/leaderboards
    op.create_table('user_points',
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('lifetime', sa.Integer(), nullable=False, default=0),
    sa.Column('weekly', sa.Integer(), nullable=False, default=0),
    sa.Column('monthly', sa.Integer(), nullable=False, default=0),
    sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.PrimaryKeyConstraint('user_id')
    )

    # rolling leaderboard snapshots (for fast reads)
    op.create_table('leaderboard_snapshots',
    sa.Column('id', sa.BigInteger(), autoincrement=True, nullable=False),
    sa.Column('period', sa.Text(), nullable=False),
    sa.Column('period_key', sa.Text(), nullable=False),
    sa.Column('rank', sa.Integer(), nullable=False),
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('points', sa.Integer(), nullable=False),
    sa.Column('taken_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.CheckConstraint("period IN ('weekly','monthly')", name='period_check'),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_index('leaderboard_period_idx', 'leaderboard_snapshots', ['period', 'period_key', 'rank'])

    # completion events feed (write-ahead of outbox to throttle)
    op.create_table('progress_events',
    sa.Column('id', sa.BigInteger(), autoincrement=True, nullable=False),
    sa.Column('user_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('type', sa.Text(), nullable=False),
    sa.Column('payload', postgresql.JSONB(astext_type=sa.Text()), nullable=False),
    sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_index('progress_events_type_time_idx', 'progress_events', ['type', sa.text('created_at DESC')])

    # outbox
    op.create_table('outbox',
    sa.Column('id', sa.BigInteger(), autoincrement=True, nullable=False),
    sa.Column('aggregate_id', postgresql.UUID(as_uuid=True), nullable=False),
    sa.Column('topic', sa.Text(), nullable=False),
    sa.Column('type', sa.Text(), nullable=False),
    sa.Column('payload', postgresql.JSONB(astext_type=sa.Text()), nullable=False),
    sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('now()'), nullable=False),
    sa.Column('published_at', sa.DateTime(timezone=True), nullable=True),
    sa.PrimaryKeyConstraint('id')
    )
    op.create_index('outbox_unpub_idx', 'outbox', ['published_at'], postgresql_where=sa.text('published_at IS NULL'))


def downgrade() -> None:
    op.drop_index('outbox_unpub_idx', table_name='outbox')
    op.drop_table('outbox')
    op.drop_index('progress_events_type_time_idx', table_name='progress_events')
    op.drop_table('progress_events')
    op.drop_index('leaderboard_period_idx', table_name='leaderboard_snapshots')
    op.drop_table('leaderboard_snapshots')
    op.drop_table('user_points')
    op.drop_table('user_streaks')
    op.drop_table('daily_activity')
    op.drop_index('sr_reviews_user_time_idx', table_name='sr_reviews')
    op.drop_table('sr_reviews')
    op.drop_index('sr_due_idx', table_name='sr_cards')
    op.drop_table('sr_cards')
    op.drop_index('quiz_answers_attempt_idx', table_name='quiz_answers')
    op.drop_table('quiz_answers')
    op.drop_index('quiz_attempts_user_quiz_time_idx', table_name='quiz_attempts')
    op.drop_table('quiz_attempts')
    op.drop_index('user_lessons_user_status_idx', table_name='user_lessons')
    op.drop_table('user_lessons')
    op.drop_table('dim_users')