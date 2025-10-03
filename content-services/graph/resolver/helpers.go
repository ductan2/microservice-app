package resolver

import (
	"content-services/graph/model"
	"content-services/internal/models"
	"content-services/internal/repository"
	"content-services/internal/taxonomy"
	"content-services/internal/types"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// toStringPtr converts a string to *string, returns nil if empty
func toStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefString safely dereferences a *string, returns empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// derefInt safely dereferences an *int, returns 0 if nil
func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// toIntPtr converts an int to *int, returns nil if zero
func toIntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// mapTaxonomyError maps taxonomy store errors to GraphQL errors
func mapTaxonomyError(resource string, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, taxonomy.ErrDuplicate):
		return gqlerror.Errorf("%s already exists", resource)
	case errors.Is(err, taxonomy.ErrNotFound):
		return gqlerror.Errorf("%s not found", resource)
	default:
		return err
	}
}

// mapTopic converts taxonomy.Topic to model.Topic
func mapTopic(topic *taxonomy.Topic) *model.Topic {
	if topic == nil {
		return nil
	}
	return &model.Topic{
		ID:        topic.ID,
		Slug:      topic.Slug,
		Name:      topic.Name,
		CreatedAt: topic.CreatedAt,
	}
}

// mapLevel converts taxonomy.Level to model.Level
func mapLevel(level *taxonomy.Level) *model.Level {
	if level == nil {
		return nil
	}
	return &model.Level{
		ID:   level.ID,
		Code: level.Code,
		Name: level.Name,
	}
}

// mapTag converts taxonomy.Tag to model.Tag
func mapTag(tag *taxonomy.Tag) *model.Tag {
	if tag == nil {
		return nil
	}
	return &model.Tag{
		ID:   tag.ID,
		Slug: tag.Slug,
		Name: tag.Name,
	}
}

// ============= LESSON MAPPERS =============

// mapLessonError maps lesson store errors to GraphQL errors
func mapLessonError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrLessonNotFound):
		return gqlerror.Errorf("lesson not found")
	case errors.Is(err, types.ErrDuplicateCode):
		return gqlerror.Errorf("lesson code already exists")
	case errors.Is(err, types.ErrAlreadyPublished):
		return gqlerror.Errorf("lesson is already published")
	default:
		return err
	}
}

// mapFlashcardError converts flashcard repository errors to GraphQL-friendly errors.
func mapFlashcardError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, repository.ErrFlashcardSetNotFound):
		return gqlerror.Errorf("flashcard set not found")
	case errors.Is(err, repository.ErrFlashcardNotFound):
		return gqlerror.Errorf("flashcard not found")
	default:
		return err
    
// mapLessonSectionError maps lesson section errors to GraphQL errors.
func mapLessonSectionError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, types.ErrLessonSectionNotFound):
		return gqlerror.Errorf("lesson section not found")
	default:
		return mapLessonError(err)
	}
}

// mapLesson converts models.Lesson to model.Lesson
func mapLesson(l *models.Lesson) *model.Lesson {
	if l == nil {
		return nil
	}

	var code *string
	if l.Code != "" {
		code = &l.Code
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
		Description: toStringPtr(l.Description),
		IsPublished: l.IsPublished,
		Version:     l.Version,
		CreatedBy:   createdBy,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}

	if l.PublishedAt.Valid {
		mapped.PublishedAt = &l.PublishedAt.Time
	}

	// Note: Topic and Level are resolved separately via field resolvers
	// They are not set here

	return mapped
}

// mapLessonSection converts models.LessonSection to model.LessonSection.
func mapLessonSection(section *models.LessonSection) *model.LessonSection {
	if section == nil {
		return nil
	}

	return &model.LessonSection{
		ID:        section.ID.String(),
		LessonID:  section.LessonID.String(),
		Ord:       section.Ord,
		Type:      mapLessonSectionType(section.Type),
		Body:      cloneBody(section.Body),
		CreatedAt: section.CreatedAt,
	}
}

// mapLessonSections converts a slice of models.LessonSection to GraphQL model.
func mapLessonSections(sections []models.LessonSection) []*model.LessonSection {
	if len(sections) == 0 {
		return []*model.LessonSection{}
	}

	result := make([]*model.LessonSection, 0, len(sections))
	for i := range sections {
		result = append(result, mapLessonSection(&sections[i]))
	}

	return result
}

// mapLessonSectionType converts persisted type string to GraphQL enum.
func mapLessonSectionType(sectionType string) model.LessonSectionType {
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

// normalizeLessonSectionType converts GraphQL enum to storage string.
func normalizeLessonSectionType(sectionType model.LessonSectionType) string {
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

// buildLessonFilter converts GraphQL filter input to repository filter.
func buildLessonFilter(input *model.LessonFilterInput) (*repository.LessonFilter, error) {
	if input == nil {
		return nil, nil
	}

	filter := &repository.LessonFilter{}

	if input.TopicID != nil && *input.TopicID != "" {
		topicID, err := uuid.Parse(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topic ID: %v", err)
		}
		filter.TopicID = &topicID
	}

	if input.LevelID != nil && *input.LevelID != "" {
		levelID, err := uuid.Parse(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid level ID: %v", err)
		}
		filter.LevelID = &levelID
	}

	if input.IsPublished != nil {
		filter.IsPublished = input.IsPublished
	}

	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}

	return filter, nil
}

func cloneBody(body map[string]any) map[string]any {
	if body == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(body))
	for k, v := range body {
		cloned[k] = v
	}

	return cloned
}

// mapMediaAsset converts models.MediaAsset to model.MediaAsset with presigned URL
func (r *Resolver) mapMediaAsset(ctx context.Context, media *models.MediaAsset) (*model.MediaAsset, error) {
	if media == nil {
		return nil, nil
	}

	var uploadedBy *string
	if media.UploadedBy != nil {
		id := media.UploadedBy.String()
		uploadedBy = &id
	}

	duration := toIntPtr(media.DurationMs)

	if media.ID == uuid.Nil {
		media.ID = uuid.New()
	}

	downloadURL := ""
	if r.Media != nil {
		url, err := r.Media.GetPresignedURL(ctx, media.ID)
		if err != nil {
			return nil, err
		}
		downloadURL = url
	}

	return &model.MediaAsset{
		ID:          media.ID.String(),
		StorageKey:  media.StorageKey,
		Kind:        mapMediaKind(media.Kind),
		MimeType:    media.MimeType,
		Bytes:       media.Bytes,
		DurationMs:  duration,
		Sha256:      media.SHA256,
		CreatedAt:   media.CreatedAt,
		UploadedBy:  uploadedBy,
		DownloadURL: downloadURL,
	}, nil
}

// mapFlashcardSet converts models.FlashcardSet to model.FlashcardSet.
func mapFlashcardSet(set *models.FlashcardSet) *model.FlashcardSet {
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
		Description: toStringPtr(set.Description),
		TopicID:     topicID,
		LevelID:     levelID,
		CreatedAt:   set.CreatedAt,
		CreatedBy:   createdBy,
	}
}

// mapFlashcard converts models.Flashcard to model.Flashcard.
func mapFlashcard(card *models.Flashcard) *model.Flashcard {
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

// mapFlashcards converts slice of models.Flashcard to GraphQL models.
func mapFlashcards(cards []models.Flashcard) []*model.Flashcard {
	result := make([]*model.Flashcard, 0, len(cards))
	for i := range cards {
		result = append(result, mapFlashcard(&cards[i]))
	}
	return result
}

// mapMediaKind converts string to model.MediaKind enum
func mapMediaKind(kind string) model.MediaKind {
	switch strings.ToLower(kind) {
	case "image":
		return model.MediaKindImage
	case "audio":
		return model.MediaKindAudio
	default:
		return model.MediaKind(strings.ToUpper(kind))
	}
}
