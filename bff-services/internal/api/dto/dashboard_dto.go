package dto

// DashboardSummary aggregates the metrics needed for the home dashboard.
type DashboardSummary struct {
	LessonsCompleted   int             `json:"lessons_completed"`
	LessonsInProgress  int             `json:"lessons_in_progress"`
	StudySessions      int64           `json:"study_sessions"`
	StudyTimeMinutes   int64           `json:"study_time_minutes"`
	StudyTimeFormatted string          `json:"study_time_formatted"`
	TotalPoints        PointsBreakdown `json:"total_points"`
	CurrentStreakDays  int             `json:"current_streak_days"`
	LongestStreakDays  int             `json:"longest_streak_days"`
	LastStreakActivity *string         `json:"last_streak_activity,omitempty"`
}

// PointsBreakdown reports the lifetime/weekly/monthly point totals.
type PointsBreakdown struct {
	Lifetime int `json:"lifetime"`
	Weekly   int `json:"weekly"`
	Monthly  int `json:"monthly"`
}

// UserLessonStatsResponse mirrors the lesson service payload.
type UserLessonStatsResponse struct {
	TotalStarted   int     `json:"total_started"`
	InProgress     int     `json:"in_progress"`
	Completed      int     `json:"completed"`
	Abandoned      int     `json:"abandoned"`
	CompletionRate float64 `json:"completion_rate"`
	TotalScore     int     `json:"total_score"`
}

// UserPointsResponse mirrors the points service payload.
type UserPointsResponse struct {
	UserID   string `json:"user_id"`
	Lifetime int    `json:"lifetime"`
	Weekly   int    `json:"weekly"`
	Monthly  int    `json:"monthly"`
}

// UserStreakResponse mirrors the streak service payload.
type UserStreakResponse struct {
	UserID     string  `json:"user_id"`
	CurrentLen int     `json:"current_len"`
	LongestLen int     `json:"longest_len"`
	LastDay    *string `json:"last_day,omitempty"`
}
