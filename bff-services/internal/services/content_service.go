package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bff-services/internal/api/dto"
)

// ContentService defines the contract for interacting with the content service GraphQL API.
type ContentService interface {
	GetTopicsLevelsTags(ctx context.Context, token string) (*HTTPResponse, error)
	GetLessons(ctx context.Context, token string, params dto.LessonQueryParams) (*HTTPResponse, error)
	CreateLesson(ctx context.Context, token string, payload dto.CreateLessonRequest) (*HTTPResponse, error)
	GetFlashcardSets(ctx context.Context, token string, params dto.FlashcardQueryParams) (*HTTPResponse, error)
	GetQuizzes(ctx context.Context, token string, params dto.QuizQueryParams) (*HTTPResponse, error)
}

// ContentServiceClient implements ContentService against a remote HTTP GraphQL endpoint.
type ContentServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// NewContentServiceClient constructs a new ContentServiceClient.
func NewContentServiceClient(baseURL string, httpClient *http.Client) *ContentServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &ContentServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

const (
	topicsLevelsTagsQuery = `query {
  topics {
    id
    slug
    name
    createdAt
  }
  levels {
    id
    code
    name
  }
  tags {
    id
    slug
    name
  }
}`

	lessonsQuery = `query ($filter: LessonFilterInput, $page: Int, $pageSize: Int) {
  lessons(filter: $filter, page: $page, pageSize: $pageSize) {
    items {
      id
      code
      title
      description
      topic {
        id
        name
      }
      level {
        id
        name
      }
      isPublished
      createdAt
      sections {
        id
        type
        body
      }
    }
    totalCount
  }
}`

	createLessonMutation = `mutation ($input: CreateLessonInput!) {
  createLesson(input: $input) {
    id
    title
    code
  }
}`

	flashcardSetsQuery = `query ($topicId: ID, $levelId: ID, $page: Int, $pageSize: Int) {
  flashcardSets(topicId: $topicId, levelId: $levelId, page: $page, pageSize: $pageSize) {
    items {
      id
      title
      description
      cards {
        id
        frontText
        backText
        hints
      }
    }
    totalCount
  }
}`

	quizzesQuery = `query ($lessonId: ID!, $page: Int, $pageSize: Int) {
  quizzes(lessonId: $lessonId, page: $page, pageSize: $pageSize) {
    items {
      id
      title
    }
  }
}`
)

// GetTopicsLevelsTags fetches topics, levels, and tags metadata.
func (c *ContentServiceClient) GetTopicsLevelsTags(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doGraphQLRequest(ctx, topicsLevelsTagsQuery, nil, token)
}

// GetLessons fetches paginated lessons using optional filters.
func (c *ContentServiceClient) GetLessons(ctx context.Context, token string, params dto.LessonQueryParams) (*HTTPResponse, error) {
	variables := make(map[string]interface{})
	filter := make(map[string]interface{})

	if params.TopicID != "" {
		filter["topicId"] = params.TopicID
	}
	if params.LevelID != "" {
		filter["levelId"] = params.LevelID
	}
	if params.IsPublished != nil {
		filter["isPublished"] = *params.IsPublished
	}
	if len(filter) > 0 {
		variables["filter"] = filter
	}
	if params.Page > 0 {
		variables["page"] = params.Page
	}
	if params.PageSize > 0 {
		variables["pageSize"] = params.PageSize
	}

	return c.doGraphQLRequest(ctx, lessonsQuery, variables, token)
}

// CreateLesson creates a new lesson.
func (c *ContentServiceClient) CreateLesson(ctx context.Context, token string, payload dto.CreateLessonRequest) (*HTTPResponse, error) {
	input := map[string]interface{}{
		"title":       payload.Title,
		"description": payload.Description,
		"topicId":     payload.TopicID,
		"levelId":     payload.LevelID,
	}
	if payload.CreatedBy != "" {
		input["createdBy"] = payload.CreatedBy
	}

	variables := map[string]interface{}{
		"input": input,
	}

	return c.doGraphQLRequest(ctx, createLessonMutation, variables, token)
}

// GetFlashcardSets fetches flashcard sets with optional filters.
func (c *ContentServiceClient) GetFlashcardSets(ctx context.Context, token string, params dto.FlashcardQueryParams) (*HTTPResponse, error) {
	variables := make(map[string]interface{})
	if params.TopicID != "" {
		variables["topicId"] = params.TopicID
	}
	if params.LevelID != "" {
		variables["levelId"] = params.LevelID
	}
	if params.Page > 0 {
		variables["page"] = params.Page
	}
	if params.PageSize > 0 {
		variables["pageSize"] = params.PageSize
	}

	return c.doGraphQLRequest(ctx, flashcardSetsQuery, variables, token)
}

// GetQuizzes fetches quizzes for a lesson.
func (c *ContentServiceClient) GetQuizzes(ctx context.Context, token string, params dto.QuizQueryParams) (*HTTPResponse, error) {
	if params.LessonID == "" {
		return nil, fmt.Errorf("lessonId is required")
	}

	variables := map[string]interface{}{
		"lessonId": params.LessonID,
	}
	if params.Page > 0 {
		variables["page"] = params.Page
	}
	if params.PageSize > 0 {
		variables["pageSize"] = params.PageSize
	}

	return c.doGraphQLRequest(ctx, quizzesQuery, variables, token)
}

func (c *ContentServiceClient) doGraphQLRequest(ctx context.Context, query string, variables map[string]interface{}, token string) (*HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("content service base URL is not configured")
	}

	payload := graphQLRequest{Query: query}
	if len(variables) > 0 {
		payload.Variables = variables
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql payload: %w", err)
	}

	endpoint := c.baseURL + "/graphql"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create graphql request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform graphql request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read graphql response: %w", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}, nil
}
