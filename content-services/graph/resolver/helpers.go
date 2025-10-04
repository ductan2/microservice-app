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
	}
}

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

	if input.CreatedBy != nil && *input.CreatedBy != "" {
		createdBy, err := uuid.Parse(*input.CreatedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid createdBy: %v", err)
		}
		filter.CreatedBy = &createdBy
	}

	return filter, nil
}

func buildLessonOrder(input *model.LessonOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.LessonOrderFieldPublishedAt:
		option.Field = "published_at"
	case model.LessonOrderFieldVersion:
		option.Field = "version"
	default:
		option.Field = "created_at"
	}
	return option
}

func buildLessonSectionFilter(input *model.LessonSectionFilterInput) *repository.LessonSectionFilter {
	if input == nil {
		return nil
	}
	filter := &repository.LessonSectionFilter{}
	if input.Type != nil {
		sectionType := normalizeLessonSectionType(*input.Type)
		filter.Type = &sectionType
	}
	return filter
}

func buildLessonSectionOrder(input *model.LessonSectionOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.LessonSectionOrderFieldCreatedAt:
		option.Field = "created_at"
	default:
		option.Field = "ord"
	}
	return option
}

func buildFlashcardSetFilter(input *model.FlashcardSetFilterInput) (*repository.FlashcardSetFilter, error) {
	if input == nil {
		return nil, nil
	}
	filter := &repository.FlashcardSetFilter{}
	if input.TopicID != nil && *input.TopicID != "" {
		id, err := uuid.Parse(*input.TopicID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid topicId: %v", err)
		}
		filter.TopicID = &id
	}
	if input.LevelID != nil && *input.LevelID != "" {
		id, err := uuid.Parse(*input.LevelID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid levelId: %v", err)
		}
		filter.LevelID = &id
	}
	if input.CreatedBy != nil && *input.CreatedBy != "" {
		id, err := uuid.Parse(*input.CreatedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid createdBy: %v", err)
		}
		filter.CreatedBy = &id
	}
	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}
	return filter, nil
}

func buildFlashcardSetOrder(input *model.FlashcardSetOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.FlashcardSetOrderFieldCardCount:
		option.Field = "card_count"
	default:
		option.Field = "created_at"
	}
	return option
}

func buildFlashcardFilter(input *model.FlashcardFilterInput) *repository.FlashcardFilter {
	if input == nil {
		return nil
	}
	filter := &repository.FlashcardFilter{}
	if input.HasMedia != nil {
		filter.HasMedia = input.HasMedia
	}
	return filter
}

func buildFlashcardOrder(input *model.FlashcardOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.FlashcardOrderFieldCreatedAt:
		option.Field = "created_at"
	default:
		option.Field = "ord"
	}
	return option
}

func buildQuizOrder(input *model.QuizOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.QuizOrderFieldTotalPoints:
		option.Field = "total_points"
	default:
		option.Field = "created_at"
	}
	return option
}

func buildQuizQuestionFilter(input *model.QuizQuestionFilterInput) *repository.QuizQuestionFilter {
	if input == nil {
		return nil
	}
	filter := &repository.QuizQuestionFilter{}
	if input.Type != nil && *input.Type != "" {
		value := strings.TrimSpace(*input.Type)
		filter.Type = &value
	}
	return filter
}

func buildQuizQuestionOrder(input *model.QuizQuestionOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.QuizQuestionOrderFieldPoints:
		option.Field = "points"
	default:
		option.Field = "ord"
	}
	return option
}

func buildMediaFilter(input *model.MediaAssetFilterInput) (*repository.MediaFilter, error) {
	if input == nil {
		return nil, nil
	}
	filter := &repository.MediaFilter{}
	if input.FolderID != nil && *input.FolderID != "" {
		id, err := uuid.Parse(*input.FolderID)
		if err != nil {
			return nil, gqlerror.Errorf("invalid folderId: %v", err)
		}
		filter.FolderID = &id
	}
	if input.Kind != nil {
		kind := strings.ToLower(input.Kind.String())
		filter.Kind = kind
	}
	if input.UploadedBy != nil && *input.UploadedBy != "" {
		id, err := uuid.Parse(*input.UploadedBy)
		if err != nil {
			return nil, gqlerror.Errorf("invalid uploadedBy: %v", err)
		}
		filter.UploadedBy = &id
	}
	if input.Sha256 != nil && *input.Sha256 != "" {
		filter.SHA256 = *input.Sha256
	}
	if input.Search != nil {
		filter.Search = strings.TrimSpace(*input.Search)
	}
	return filter, nil
}

func buildMediaOrder(input *model.MediaAssetOrderInput) *repository.SortOption {
	if input == nil {
		return nil
	}
	option := &repository.SortOption{Direction: mapOrderDirection(input.Direction)}
	switch input.Field {
	case model.MediaAssetOrderFieldBytes:
		option.Field = "bytes"
	default:
		option.Field = "created_at"
	}
	return option
}

func mapOrderDirection(direction model.OrderDirection) repository.SortDirection {
	if direction == model.OrderDirectionAsc {
		return repository.SortAscending
	}
	return repository.SortDescending
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

	var folderID *string
	if media.FolderID != nil {
		id := media.FolderID.String()
		folderID = &id
	}

	var thumbnailURL *string
	if media.ThumbnailURL != "" {
		url := media.ThumbnailURL
		thumbnailURL = &url
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
		ID:           media.ID.String(),
		StorageKey:   media.StorageKey,
		Kind:         mapMediaKind(media.Kind),
		MimeType:     media.MimeType,
		FolderID:     folderID,
		OriginalName: media.OriginalName,
		ThumbnailURL: thumbnailURL,
		Bytes:        media.Bytes,
		DurationMs:   duration,
		Sha256:       media.SHA256,
		CreatedAt:    media.CreatedAt,
		UploadedBy:   uploadedBy,
		DownloadURL:  downloadURL,
	}, nil
}

// mapQuiz converts models.Quiz to model.Quiz (questions resolved separately)
func mapQuiz(q *models.Quiz) *model.Quiz {
	if q == nil {
		return nil
	}

	var lessonID *string
	if q.LessonID != nil {
		id := q.LessonID.String()
		lessonID = &id
	}

	description := toStringPtr(q.Description)
	timeLimit := toIntPtr(q.TimeLimitS)

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

// mapQuizQuestion converts models.QuizQuestion to model.QuizQuestion
func mapQuizQuestion(q *models.QuizQuestion) *model.QuizQuestion {
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

// mapQuestionOption converts models.QuestionOption to model.QuestionOption.
func mapQuestionOption(option *models.QuestionOption) *model.QuestionOption {
	if option == nil {
		return nil
	}

	feedback := toStringPtr(option.Feedback)

	return &model.QuestionOption{
		ID:         option.ID.String(),
		QuestionID: option.QuestionID.String(),
		Ord:        option.Ord,
		Label:      option.Label,
		IsCorrect:  option.IsCorrect,
		Feedback:   feedback,
	}
}

// mapQuestionOptions converts a slice of models.QuestionOption to GraphQL models.
func mapQuestionOptions(options []models.QuestionOption) []*model.QuestionOption {
	result := make([]*model.QuestionOption, 0, len(options))
	for i := range options {
		option := options[i]
		result = append(result, mapQuestionOption(&option))
	}
	return result
}

// mapRepositoryTags converts repository tag models to GraphQL models.
func mapRepositoryTags(tags []models.Tag) []*model.Tag {
	result := make([]*model.Tag, 0, len(tags))
	for i := range tags {
		tag := tags[i]
		result = append(result, &model.Tag{
			ID:   tag.ID.String(),
			Slug: tag.Slug,
			Name: tag.Name,
		})
	}
	return result
}

// contentTagKindToModel converts GraphQL enum to repository kind string.
func contentTagKindToModel(kind model.ContentTagKind) string {
	switch kind {
	case model.ContentTagKindLesson:
		return "lesson"
	case model.ContentTagKindQuiz:
		return "quiz"
	case model.ContentTagKindFlashcardSet:
		return "flashcard_set"
	default:
		return strings.ToLower(string(kind))
	}
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
