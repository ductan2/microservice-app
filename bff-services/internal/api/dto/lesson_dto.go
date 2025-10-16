package dto

// DailyActivityIncrementRequest represents payload for incrementing a user's activity counters.
type DailyActivityIncrementRequest struct {
	ActivityDate *string `json:"activity_dt,omitempty"`
	Field        string  `json:"field" binding:"required,oneof=lessons_completed quizzes_completed minutes points"`
	Amount       int     `json:"amount" binding:"required,min=1"`
}

// DimUserCreateRequest mirrors the dim user creation payload expected by the lesson service.
type DimUserCreateRequest struct {
	UserID    string  `json:"user_id"`
	Locale    *string `json:"locale,omitempty" binding:"omitempty,max=50"`
	LevelHint *string `json:"level_hint,omitempty" binding:"omitempty,max=50"`
}

// DimUserUpdateRequest mirrors the partial update payload for dim user records.
type DimUserUpdateRequest struct {
	Locale    *string `json:"locale,omitempty" binding:"omitempty,max=50"`
	LevelHint *string `json:"level_hint,omitempty" binding:"omitempty,max=50"`
}

// DimUserLocaleUpdateRequest represents the payload for locale-only updates.
type DimUserLocaleUpdateRequest struct {
	Locale string `json:"locale" binding:"required,max=50"`
}
