package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"content-services/internal/taxonomy"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Topic is the resolver for the topic field.
func (r *lessonResolver) Topic(ctx context.Context, obj *model.Lesson) (*model.Topic, error) {
	lessonID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, nil
	}

	lessonDoc, err := r.LessonService.GetLessonByID(ctx, lessonID)
	if err != nil {
		return nil, nil
	}

	if lessonDoc.TopicID == nil {
		return nil, nil
	}

	if r.Taxonomy == nil {
		return nil, nil
	}

	topic, err := r.Taxonomy.GetTopicByID(ctx, lessonDoc.TopicID.String())
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapTopic(topic), nil
}

// Level is the resolver for the level field.
func (r *lessonResolver) Level(ctx context.Context, obj *model.Lesson) (*model.Level, error) {
	lessonID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, nil
	}

	lessonDoc, err := r.LessonService.GetLessonByID(ctx, lessonID)
	if err != nil {
		return nil, nil
	}

	if lessonDoc.LevelID == nil {
		return nil, nil
	}

	if r.Taxonomy == nil {
		return nil, nil
	}

	level, err := r.Taxonomy.GetLevelByID(ctx, lessonDoc.LevelID.String())
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return mapLevel(level), nil
}

// Sections is the resolver for the sections field.
func (r *lessonResolver) Sections(ctx context.Context, obj *model.Lesson) ([]*model.LessonSection, error) {
	if r.LessonService == nil {
		return nil, gqlerror.Errorf("lesson service not configured")
	}

	lessonID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	sections, err := r.LessonService.GetLessonSections(ctx, lessonID)
	if err != nil {
		return nil, mapLessonSectionError(err)
	}

	return mapLessonSections(sections), nil
}

// Tags is the resolver for the tags field.
func (r *lessonResolver) Tags(ctx context.Context, obj *model.Lesson) ([]*model.Tag, error) {
	if r.TagRepo == nil {
		return []*model.Tag{}, nil
	}

	lessonID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid lesson ID: %v", err)
	}

	tags, err := r.TagRepo.GetContentTags(ctx, "lesson", lessonID)
	if err != nil {
		return nil, err
	}

	return mapRepositoryTags(tags), nil
}

// Lesson returns generated.LessonResolver implementation.
func (r *Resolver) Lesson() generated.LessonResolver { return &lessonResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type lessonResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
