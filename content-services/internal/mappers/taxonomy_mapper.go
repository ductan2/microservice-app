package mappers

import (
	"content-services/graph/model"
	"content-services/internal/taxonomy"
)

// TopicToGraphQL converts taxonomy.Topic to model.Topic
func TopicToGraphQL(topic *taxonomy.Topic) *model.Topic {
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

// LevelToGraphQL converts taxonomy.Level to model.Level
func LevelToGraphQL(level *taxonomy.Level) *model.Level {
	if level == nil {
		return nil
	}
	return &model.Level{
		ID:   level.ID,
		Code: level.Code,
		Name: level.Name,
	}
}

// TagToGraphQL converts taxonomy.Tag to model.Tag
func TagToGraphQL(tag *taxonomy.Tag) *model.Tag {
	if tag == nil {
		return nil
	}
	return &model.Tag{
		ID:   tag.ID,
		Slug: tag.Slug,
		Name: tag.Name,
	}
}