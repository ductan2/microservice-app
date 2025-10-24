package dto

// QuizAttemptStartRequest represents the payload for starting a quiz attempt from the BFF.
type QuizAttemptStartRequest struct {
	QuizID      string  `json:"quiz_id" binding:"required,uuid"`
	LessonID    *string `json:"lesson_id,omitempty" binding:"omitempty,uuid"`
	DurationMs  *int    `json:"duration_ms,omitempty"`
	TotalPoints *int    `json:"total_points,omitempty"`
	MaxPoints   *int    `json:"max_points,omitempty"`
	Passed      *bool   `json:"passed,omitempty"`
}

// QuizAnswerSubmissionRequest represents an answer submission within a quiz attempt submission payload.
type QuizAnswerSubmissionRequest struct {
	QuestionID   string   `json:"question_id" binding:"required,uuid"`
	SelectedIDs  []string `json:"selected_ids,omitempty"`
	TextAnswer   *string  `json:"text_answer,omitempty"`
	IsCorrect    *bool    `json:"is_correct,omitempty"`
	PointsEarned *int     `json:"points_earned,omitempty"`
}

// QuizAttemptSubmitRequest represents the payload for submitting a quiz attempt.
type QuizAttemptSubmitRequest struct {
	TotalPoints int                           `json:"total_points" binding:"required"`
	MaxPoints   *int                          `json:"max_points,omitempty"`
	Passed      *bool                         `json:"passed,omitempty"`
	SubmittedAt *string                       `json:"submitted_at,omitempty"`
	DurationMs  *int                          `json:"duration_ms,omitempty"`
	Answers     []QuizAnswerSubmissionRequest `json:"answers,omitempty"`
}
