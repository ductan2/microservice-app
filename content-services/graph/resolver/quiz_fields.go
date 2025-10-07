package resolver

import (
	"content-services/graph/generated"
	"content-services/graph/model"
	"context"

	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Quiz returns generated.QuizResolver implementation.
func (r *Resolver) Quiz() generated.QuizResolver { return &quizResolver{r} }

type quizResolver struct{ *Resolver }

// Questions is the resolver for the questions field.
func (r *quizResolver) Questions(ctx context.Context, obj *model.Quiz) ([]*model.QuizQuestion, error) {
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, gqlerror.Errorf("quiz service not configured")
	}

	quizID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	questions, err := quizService.GetQuizQuestions(ctx, quizID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.QuizQuestion, 0, len(questions))
	for i := range questions {
		question := questions[i]
		result = append(result, mapQuizQuestion(&question))
	}

	return result, nil
}

// Tags is the resolver for the tags field.
func (r *quizResolver) Tags(ctx context.Context, obj *model.Quiz) ([]*model.Tag, error) {
	if r.TagRepo == nil {
		return []*model.Tag{}, nil
	}

	quizID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	tags, err := r.TagRepo.GetContentTags(ctx, "quiz", quizID)
	if err != nil {
		return nil, err
	}

	return mapRepositoryTags(tags), nil
}

// Topic is the resolver for the topic field.
func (r *quizResolver) Topic(ctx context.Context, obj *model.Quiz) (*model.Topic, error) {
	if r.Taxonomy == nil {
		return nil, nil
	}

	// Get the quiz to access TopicID
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, nil
	}

	quizID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	quiz, err := quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if quiz.TopicID == nil {
		return nil, nil
	}

	topic, err := r.Taxonomy.GetTopicByID(ctx, quiz.TopicID.String())
	if err != nil {
		return nil, err
	}

	return mapTopic(topic), nil
}

// Level is the resolver for the level field.
func (r *quizResolver) Level(ctx context.Context, obj *model.Quiz) (*model.Level, error) {
	if r.Taxonomy == nil {
		return nil, nil
	}

	// Get the quiz to access LevelID
	quizService := r.Resolver.QuizService
	if quizService == nil {
		return nil, nil
	}

	quizID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, gqlerror.Errorf("invalid quiz ID: %v", err)
	}

	quiz, err := quizService.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if quiz.LevelID == nil {
		return nil, nil
	}

	level, err := r.Taxonomy.GetLevelByID(ctx, quiz.LevelID.String())
	if err != nil {
		return nil, err
	}

	return mapLevel(level), nil
}
