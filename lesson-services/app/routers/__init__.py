"""
Routers package for the lesson services API.
Contains all API route definitions organized by feature.
"""

from . import (
    daily_activity_routes,
    dim_user_routes,
    health_routes,
    leaderboard_routes,
    outbox_routes,
    progress_event_routes,
    quiz_answer_routes,
    quiz_attempt_routes,
    sr_card_routes,
    sr_review_routes,
    user_lesson_routes,
    user_points_routes,
    user_streak_routes,
)

__all__ = [
    "daily_activity_routes",
    "dim_user_routes", 
    "health_routes",
    "leaderboard_routes",
    "outbox_routes",
    "progress_event_routes",
    "quiz_answer_routes",
    "quiz_attempt_routes",
    "sr_card_routes",
    "sr_review_routes",
    "user_lesson_routes",
    "user_points_routes",
    "user_streak_routes",
]
