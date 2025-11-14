package dto

// CreateCourseRequest represents the input for creating a new course
type CreateCourseRequest struct {
	Title          string  `json:"title"`
	Description    *string `json:"description,omitempty"`
	TopicID        *string `json:"topic_id,omitempty"`
	LevelID        *string `json:"level_id,omitempty"`
	InstructorID   *string `json:"instructor_id,omitempty"`
	ThumbnailURL   *string `json:"thumbnail_url,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	DurationHours  *int     `json:"duration_hours,omitempty"`
	IsPublished    bool     `json:"is_published"`
	IsFeatured     bool     `json:"is_featured"`
}

// UpdateCourseRequest represents the input for updating an existing course
type UpdateCourseRequest struct {
	Title          *string  `json:"title,omitempty"`
	Description    *string  `json:"description,omitempty"`
	TopicID        *string  `json:"topic_id,omitempty"`
	LevelID        *string  `json:"level_id,omitempty"`
	InstructorID   *string  `json:"instructor_id,omitempty"`
	ThumbnailURL   *string  `json:"thumbnail_url,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	DurationHours  *int     `json:"duration_hours,omitempty"`
	IsPublished    *bool    `json:"is_published,omitempty"`
	IsFeatured     *bool    `json:"is_featured,omitempty"`
}

// AddCourseLessonRequest represents the input for adding a lesson to a course
type AddCourseLessonRequest struct {
	CourseID   string `json:"course_id"`
	LessonID   string `json:"lesson_id"`
	Order      int    `json:"order"`
	IsRequired bool   `json:"is_required"`
}

// UpdateCourseLessonRequest represents the input for updating a course lesson
type UpdateCourseLessonRequest struct {
	Order      *int  `json:"order,omitempty"`
	IsRequired *bool `json:"is_required,omitempty"`
}

// CreateCourseReviewRequest represents the input for creating a course review
type CreateCourseReviewRequest struct {
	CourseID string  `json:"course_id"`
	Rating   int     `json:"rating"`
	Comment  *string `json:"comment,omitempty"`
}

// UpdateCourseReviewRequest represents the input for updating a course review
type UpdateCourseReviewRequest struct {
	Rating  *int    `json:"rating,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

// CourseFilterRequest represents filtering options for course queries
type CourseFilterRequest struct {
	TopicID      *string `json:"topic_id,omitempty"`
	LevelID      *string `json:"level_id,omitempty"`
	InstructorID *string `json:"instructor_id,omitempty"`
	IsPublished  *bool   `json:"is_published,omitempty"`
	IsFeatured   *bool   `json:"is_featured,omitempty"`
	Search       *string `json:"search,omitempty"`
}

// CourseOrderRequest represents sorting options for course queries
type CourseOrderRequest struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}