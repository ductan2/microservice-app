package dto

// DailyActivityIncrementRequest represents payload for incrementing a user's activity counters.
type DailyActivityIncrementRequest struct {
	ActivityDate *string `json:"activity_dt,omitempty"`
	Field        string  `json:"field" binding:"required,oneof=lessons_completed quizzes_completed minutes points"`
	Amount       int     `json:"amount" binding:"required,min=1"`
}

// StreakCheckRequest represents optional payload for manual streak validation.
type StreakCheckRequest struct {
	ActivityDate *string `json:"activity_date,omitempty"`
}

// Course enrollment DTOs
type CourseEnrollmentCreate struct {
	CourseID string `json:"course_id" binding:"required,uuid4"`
}

type CourseEnrollmentUpdate struct {
	Status          *string `json:"status,omitempty"`
	ProgressPercent *int    `json:"progress_percent,omitempty"`
	StartedAt       *string `json:"started_at,omitempty"`
	CompletedAt     *string `json:"completed_at,omitempty"`
	LastAccessedAt  *string `json:"last_accessed_at,omitempty"`
}

// Course lesson DTOs
type CourseLessonCreate struct {
	CourseID   string `json:"course_id" binding:"required,uuid4"`
	LessonID   string `json:"lesson_id" binding:"required,uuid4"`
	Ord        int    `json:"ord" binding:"required,min=0"`
	IsRequired bool   `json:"is_required"`
}

type CourseLessonUpdate struct {
	Ord        *int  `json:"ord,omitempty"`
	IsRequired *bool `json:"is_required,omitempty"`
}
