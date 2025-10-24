package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/types"
)

// QuizAttemptService defines the contract for interacting with quiz attempt endpoints.
type QuizAttemptService interface {
	StartQuizAttempt(ctx context.Context, userID, email, sessionID string, payload dto.QuizAttemptStartRequest) (*types.HTTPResponse, error)
	GetQuizAttempt(ctx context.Context, attemptID, userID, email, sessionID string) (*types.HTTPResponse, error)
	SubmitQuizAttempt(ctx context.Context, attemptID, userID, email, sessionID string, payload dto.QuizAttemptSubmitRequest) (*types.HTTPResponse, error)
	GetUserQuizAttempts(ctx context.Context, userID, email, sessionID, quizID string) (*types.HTTPResponse, error)
	GetUserQuizHistory(ctx context.Context, userID, email, sessionID string, passed *bool, limit, offset int) (*types.HTTPResponse, error)
	GetLessonQuizAttempts(ctx context.Context, lessonID, userID, email, sessionID string) (*types.HTTPResponse, error)
	DeleteQuizAttempt(ctx context.Context, attemptID, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetQuizAttemptsByUserID(ctx context.Context, targetUserID, userID, email, sessionID string) (*types.HTTPResponse, error)
}

// QuizAttemptServiceClient implements QuizAttemptService against the lesson service.
type QuizAttemptServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewQuizAttemptServiceClient builds a new QuizAttemptServiceClient.
func NewQuizAttemptServiceClient(baseURL string, httpClient *http.Client) *QuizAttemptServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &QuizAttemptServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *QuizAttemptServiceClient) StartQuizAttempt(ctx context.Context, userID, email, sessionID string, payload dto.QuizAttemptStartRequest) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if payload.QuizID == "" {
		return nil, fmt.Errorf("quiz ID is required")
	}

	body := struct {
		UserID      string  `json:"user_id"`
		QuizID      string  `json:"quiz_id"`
		LessonID    *string `json:"lesson_id,omitempty"`
		DurationMs  *int    `json:"duration_ms,omitempty"`
		TotalPoints *int    `json:"total_points,omitempty"`
		MaxPoints   *int    `json:"max_points,omitempty"`
		Passed      *bool   `json:"passed,omitempty"`
	}{
		UserID:      userID,
		QuizID:      payload.QuizID,
		LessonID:    payload.LessonID,
		DurationMs:  payload.DurationMs,
		TotalPoints: payload.TotalPoints,
		MaxPoints:   payload.MaxPoints,
		Passed:      payload.Passed,
	}

	return c.doRequest(ctx, http.MethodPost, "/api/v1/quiz-attempts/start", body, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) GetQuizAttempt(ctx context.Context, attemptID, userID, email, sessionID string) (*types.HTTPResponse, error) {
	path := "/api/v1/quiz-attempts/" + url.PathEscape(attemptID)
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) SubmitQuizAttempt(ctx context.Context, attemptID, userID, email, sessionID string, payload dto.QuizAttemptSubmitRequest) (*types.HTTPResponse, error) {
	if payload.TotalPoints < 0 {
		return nil, fmt.Errorf("total points must be non-negative")
	}
	path := "/api/v1/quiz-attempts/" + url.PathEscape(attemptID) + "/submit"
	return c.doRequest(ctx, http.MethodPost, path, payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) GetUserQuizAttempts(ctx context.Context, userID, email, sessionID, quizID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if quizID == "" {
		return nil, fmt.Errorf("quiz ID is required")
	}

	path := "/api/v1/quiz-attempts/user/me/quiz/" + url.PathEscape(quizID)
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) GetUserQuizHistory(ctx context.Context, userID, email, sessionID string, passed *bool, limit, offset int) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if limit < 1 {
		return nil, fmt.Errorf("limit must be at least 1")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}

	path := "/api/v1/quiz-attempts/user/me/history"
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	if passed != nil {
		query.Set("passed", strconv.FormatBool(*passed))
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) GetLessonQuizAttempts(ctx context.Context, lessonID, userID, email, sessionID string) (*types.HTTPResponse, error) {
	if lessonID == "" {
		return nil, fmt.Errorf("lesson ID is required")
	}

	path := "/api/v1/quiz-attempts/lesson/" + url.PathEscape(lessonID) + "/user/me"
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) DeleteQuizAttempt(ctx context.Context, attemptID, userID, email, sessionID string) (*types.HTTPResponse, error) {
	path := "/api/v1/quiz-attempts/" + url.PathEscape(attemptID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, internalAuthHeaders(userID, email, sessionID))
}
func (c *QuizAttemptServiceClient) GetQuizAttemptsByUserID(ctx context.Context, targetUserID, userID, email, sessionID string) (*types.HTTPResponse, error) {

	path := "/api/v1/quiz-attempts/user/" + url.PathEscape(targetUserID)
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *QuizAttemptServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("quiz attempt service base URL is not configured")
	}

	endpoint := c.baseURL + path

	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &types.HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}, nil
}
