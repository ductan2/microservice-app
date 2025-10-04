package dto

// LessonQueryParams represents query parameters for listing lessons.
type LessonQueryParams struct {
	TopicID     string `form:"topicId"`
	LevelID     string `form:"levelId"`
	IsPublished *bool  `form:"isPublished"`
	Page        int    `form:"page,default=1"`
	PageSize    int    `form:"pageSize,default=10"`
}

// CreateLessonRequest represents the payload for creating a lesson.
type CreateLessonRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	TopicID     string `json:"topicId" binding:"required"`
	LevelID     string `json:"levelId" binding:"required"`
	CreatedBy   string `json:"createdBy"`
}

// FlashcardQueryParams represents query parameters for listing flashcard sets.
type FlashcardQueryParams struct {
	TopicID  string `form:"topicId"`
	LevelID  string `form:"levelId"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
}

// QuizQueryParams represents query parameters for listing quizzes.
type QuizQueryParams struct {
	LessonID string `form:"lessonId" binding:"required"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
}
