package dto

// CreateLessonRequest represents the input for creating a new lesson
type CreateLessonRequest struct {
	Code        *string `json:"code,omitempty"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	TopicID     *string `json:"topic_id,omitempty"`
	LevelID     *string `json:"level_id,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
}

// UpdateLessonRequest represents the input for updating an existing lesson
type UpdateLessonRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	TopicID     *string `json:"topic_id,omitempty"`
	LevelID     *string `json:"level_id,omitempty"`
}

// CreateLessonSectionRequest represents the input for creating a new lesson section
type CreateLessonSectionRequest struct {
	LessonID string                 `json:"lesson_id"`
	Order    int                    `json:"order"`
	Type     string                 `json:"type"`
	Body     map[string]interface{} `json:"body"`
}

// UpdateLessonSectionRequest represents the input for updating a lesson section
type UpdateLessonSectionRequest struct {
	Order *int                    `json:"order,omitempty"`
	Type  *string                 `json:"type,omitempty"`
	Body  *map[string]interface{} `json:"body,omitempty"`
}

// LessonFilterRequest represents filtering options for lesson queries
type LessonFilterRequest struct {
	TopicID   *string `json:"topic_id,omitempty"`
	LevelID   *string `json:"level_id,omitempty"`
	IsPublished *bool  `json:"is_published,omitempty"`
	Search    *string `json:"search,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
}

// LessonOrderRequest represents sorting options for lesson queries
type LessonOrderRequest struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}