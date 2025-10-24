package dto

// DailyActivityIncrementRequest represents payload for incrementing a user's activity counters.
type DailyActivityIncrementRequest struct {
	ActivityDate *string `json:"activity_dt,omitempty"`
	Field        string  `json:"field" binding:"required,oneof=lessons_completed quizzes_completed minutes points"`
	Amount       int     `json:"amount" binding:"required,min=1"`
}
