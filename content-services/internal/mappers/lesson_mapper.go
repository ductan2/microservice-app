package mappers

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/utils"
	"strings"
	"time"
)

// LessonToGraphQL converts models.Lesson to model.Lesson
func LessonToGraphQL(l *models.Lesson) *model.Lesson {
	if l == nil {
		return nil
	}

	var code *string
	if l.Code != "" {
		code = &l.Code
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = l.CreatedAt
	}

	var createdBy *string
	if l.CreatedBy != nil {
		id := l.CreatedBy.String()
		createdBy = &id
	}

	mapped := &model.Lesson{
		ID:          l.ID.String(),
		Code:        code,
		Title:       l.Title,
		Description: utils.ToStringPtr(l.Description),
		IsPublished: l.IsPublished,
		Version:     l.Version,
		CreatedBy:   createdBy,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
		Tags:        []*model.Tag{},
		Sections:    []*model.LessonSection{},
	}

	if l.PublishedAt.Valid {
		mapped.PublishedAt = &l.PublishedAt.Time
	}

	// Note: Topic and Level are resolved separately via field resolvers
	// They are not set here

	return mapped
}

// LessonSectionToGraphQL converts models.LessonSection to model.LessonSection
func LessonSectionToGraphQL(section *models.LessonSection) *model.LessonSection {
	if section == nil {
		return nil
	}

	return &model.LessonSection{
		ID:        section.ID.String(),
		LessonID:  section.LessonID.String(),
		Ord:       section.Ord,
		Type:      MapLessonSectionType(section.Type),
		Body:      utils.CloneBody(section.Body),
		CreatedAt: section.CreatedAt,
	}
}

// LessonsToGraphQL converts a slice of models.Lesson to GraphQL models
func LessonsToGraphQL(lessons []models.Lesson) []*model.Lesson {
	if len(lessons) == 0 {
		return []*model.Lesson{}
	}

	result := make([]*model.Lesson, 0, len(lessons))
	for i := range lessons {
		result = append(result, LessonToGraphQL(&lessons[i]))
	}

	return result
}

// LessonSectionsToGraphQL converts a slice of models.LessonSection to GraphQL model
func LessonSectionsToGraphQL(sections []models.LessonSection) []*model.LessonSection {
	if len(sections) == 0 {
		return []*model.LessonSection{}
	}

	result := make([]*model.LessonSection, 0, len(sections))
	for i := range sections {
		result = append(result, LessonSectionToGraphQL(&sections[i]))
	}

	return result
}

// MapLessonSectionType converts persisted type string to GraphQL enum
func MapLessonSectionType(sectionType string) model.LessonSectionType {
	switch strings.ToLower(sectionType) {
	case "dialog":
		return model.LessonSectionTypeDialog
	case "audio":
		return model.LessonSectionTypeAudio
	case "image":
		return model.LessonSectionTypeImage
	case "exercise":
		return model.LessonSectionTypeExercise
	default:
		return model.LessonSectionTypeText
	}
}

// NormalizeLessonSectionType converts GraphQL enum to storage string
func NormalizeLessonSectionType(sectionType model.LessonSectionType) string {
	switch sectionType {
	case model.LessonSectionTypeDialog:
		return "dialog"
	case model.LessonSectionTypeAudio:
		return "audio"
	case model.LessonSectionTypeImage:
		return "image"
	case model.LessonSectionTypeExercise:
		return "exercise"
	default:
		return "text"
	}
}

// QuizToGraphQL converts models.Quiz to model.Quiz (questions resolved separately)
func QuizToGraphQL(q *models.Quiz) *model.Quiz {
	if q == nil {
		return nil
	}

	var lessonID *string
	if q.LessonID != nil {
		id := q.LessonID.String()
		lessonID = &id
	}

	description := utils.ToStringPtr(q.Description)
	timeLimit := utils.ToIntPtr(q.TimeLimitS)

	return &model.Quiz{
		ID:          q.ID.String(),
		LessonID:    lessonID,
		Title:       q.Title,
		Description: description,
		TotalPoints: q.TotalPoints,
		TimeLimitS:  timeLimit,
		CreatedAt:   q.CreatedAt,
	}
}

// QuizQuestionToGraphQL converts models.QuizQuestion to model.QuizQuestion
func QuizQuestionToGraphQL(q *models.QuizQuestion) *model.QuizQuestion {
	if q == nil {
		return nil
	}

	var promptMedia *string
	if q.PromptMedia != nil {
		id := q.PromptMedia.String()
		promptMedia = &id
	}

	metadata := map[string]any{}
	if q.Metadata != nil {
		metadata = q.Metadata
	}

	return &model.QuizQuestion{
		ID:          q.ID.String(),
		QuizID:      q.QuizID.String(),
		Ord:         q.Ord,
		Type:        q.Type,
		Prompt:      q.Prompt,
		PromptMedia: promptMedia,
		Points:      q.Points,
		Metadata:    metadata,
	}
}

// QuestionOptionToGraphQL converts models.QuestionOption to model.QuestionOption
func QuestionOptionToGraphQL(option *models.QuestionOption) *model.QuestionOption {
	if option == nil {
		return nil
	}

	feedback := utils.ToStringPtr(option.Feedback)

	return &model.QuestionOption{
		ID:         option.ID.String(),
		QuestionID: option.QuestionID.String(),
		Ord:        option.Ord,
		Label:      option.Label,
		IsCorrect:  option.IsCorrect,
		Feedback:   feedback,
	}
}

// QuestionOptionsToGraphQL converts a slice of models.QuestionOption to GraphQL models
func QuestionOptionsToGraphQL(options []models.QuestionOption) []*model.QuestionOption {
	result := make([]*model.QuestionOption, 0, len(options))
	for i := range options {
		option := options[i]
		result = append(result, QuestionOptionToGraphQL(&option))
	}
	return result
}

// FlashcardSetToGraphQL converts models.FlashcardSet to model.FlashcardSet
func FlashcardSetToGraphQL(set *models.FlashcardSet) *model.FlashcardSet {
	if set == nil {
		return nil
	}
	var (
		topicID   *string
		levelID   *string
		createdBy *string
	)
	if set.TopicID != nil {
		id := set.TopicID.String()
		topicID = &id
	}
	if set.LevelID != nil {
		id := set.LevelID.String()
		levelID = &id
	}
	if set.CreatedBy != nil {
		id := set.CreatedBy.String()
		createdBy = &id
	}
	return &model.FlashcardSet{
		ID:          set.ID.String(),
		Title:       set.Title,
		Description: utils.ToStringPtr(set.Description),
		TopicID:     topicID,
		LevelID:     levelID,
		CreatedAt:   set.CreatedAt,
		CreatedBy:   createdBy,
	}
}

// FlashcardToGraphQL converts models.Flashcard to model.Flashcard
func FlashcardToGraphQL(card *models.Flashcard) *model.Flashcard {
	if card == nil {
		return nil
	}
	var (
		frontMediaID *string
		backMediaID  *string
	)
	if card.FrontMediaID != nil {
		id := card.FrontMediaID.String()
		frontMediaID = &id
	}
	if card.BackMediaID != nil {
		id := card.BackMediaID.String()
		backMediaID = &id
	}
	modelHints := make([]string, len(card.Hints))
	copy(modelHints, card.Hints)

	return &model.Flashcard{
		ID:           card.ID.String(),
		SetID:        card.SetID.String(),
		FrontText:    card.FrontText,
		BackText:     card.BackText,
		FrontMediaID: frontMediaID,
		BackMediaID:  backMediaID,
		Ord:          card.Ord,
		Hints:        modelHints,
		CreatedAt:    card.CreatedAt,
	}
}

// FlashcardsToGraphQL converts slice of models.Flashcard to GraphQL models
func FlashcardsToGraphQL(cards []models.Flashcard) []*model.Flashcard {
	result := make([]*model.Flashcard, 0, len(cards))
	for i := range cards {
		result = append(result, FlashcardToGraphQL(&cards[i]))
	}
	return result
}