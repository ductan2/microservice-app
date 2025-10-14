package resolver

import (
	"content-services/graph/model"
	"content-services/internal/taxonomy"
	"context"
	"errors"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Topic is the resolver for the topic field.
func (r *queryResolver) Topic(ctx context.Context, id *string, slug *string) (*model.Topic, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	if id == nil && slug == nil {
		return nil, nil
	}
	var (
		topic *taxonomy.Topic
		err   error
	)
	switch {
	case id != nil:
		topic, err = r.Taxonomy.GetTopicByID(ctx, *id)
	case slug != nil:
		topic, err = r.Taxonomy.GetTopicBySlug(ctx, *slug)
	}
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapTopic(topic), nil
}

// Topics is the resolver for the topics field.
func (r *queryResolver) Topics(ctx context.Context, search *string, page *int, pageSize *int) (*model.TopicCollection, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	term := ""
	if search != nil {
		term = strings.TrimSpace(*search)
	}
	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}
	pageSizeVal := 10
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	topics, total, err := r.Taxonomy.ListTopics(ctx, term, pageVal, pageSizeVal)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Topic, 0, len(topics))
	for _, topic := range topics {
		result = append(result, mapTopic(&topic))
	}
	return &model.TopicCollection{
		Items:      result,
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}

// Level is the resolver for the level field.
func (r *queryResolver) Level(ctx context.Context, id *string, code *string) (*model.Level, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	if id == nil && code == nil {
		return nil, nil
	}
	var (
		level *taxonomy.Level
		err   error
	)
	switch {
	case id != nil:
		level, err = r.Taxonomy.GetLevelByID(ctx, *id)
	case code != nil:
		level, err = r.Taxonomy.GetLevelByCode(ctx, *code)
	}
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapLevel(level), nil
}

// Levels is the resolver for the levels field.
func (r *queryResolver) Levels(ctx context.Context, search *string, page *int, pageSize *int) (*model.LevelCollection, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	term := ""
	if search != nil {
		term = strings.TrimSpace(*search)
	}
	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}
	pageSizeVal := 10
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	levels, total, err := r.Taxonomy.ListLevels(ctx, term, pageVal, pageSizeVal)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Level, 0, len(levels))
	for _, level := range levels {
		result = append(result, mapLevel(&level))
	}
	return &model.LevelCollection{
		Items:      result,
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}

// Tag is the resolver for the tag field.
func (r *queryResolver) Tag(ctx context.Context, id *string, slug *string) (*model.Tag, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	if id == nil && slug == nil {
		return nil, nil
	}
	var (
		tag *taxonomy.Tag
		err error
	)
	switch {
	case id != nil:
		tag, err = r.Taxonomy.GetTagByID(ctx, *id)
	case slug != nil:
		tag, err = r.Taxonomy.GetTagBySlug(ctx, *slug)
	}
	if err != nil {
		if errors.Is(err, taxonomy.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapTag(tag), nil
}

// Tags is the resolver for the tags field.
func (r *queryResolver) Tags(ctx context.Context, search *string, page *int, pageSize *int) (*model.TagCollection, error) {
	if r.Taxonomy == nil {
		return nil, gqlerror.Errorf("taxonomy store not configured")
	}
	term := ""
	if search != nil {
		term = strings.TrimSpace(*search)
	}
	pageVal := 1
	if page != nil && *page > 0 {
		pageVal = *page
	}
	pageSizeVal := 10
	if pageSize != nil && *pageSize > 0 {
		pageSizeVal = *pageSize
	}

	tags, total, err := r.Taxonomy.ListTags(ctx, term, pageVal, pageSizeVal)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Tag, 0, len(tags))
	for _, tag := range tags {
		result = append(result, mapTag(&tag))
	}
	return &model.TagCollection{
		Items:      result,
		TotalCount: int(total),
		Page:       pageVal,
		PageSize:   pageSizeVal,
	}, nil
}
