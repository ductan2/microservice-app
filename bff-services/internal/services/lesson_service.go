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
)

// LessonService defines the contract for interacting with the lesson service API.
type LessonService interface {
	GetUserPoints(ctx context.Context, userID string) (*HTTPResponse, error)
	GetUserStreak(ctx context.Context, userID string) (*HTTPResponse, error)
}

// LessonServiceClient implements LessonService against a remote HTTP REST endpoint.
type LessonServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLessonServiceClient constructs a new LessonServiceClient.
func NewLessonServiceClient(baseURL string, httpClient *http.Client) *LessonServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &LessonServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *LessonServiceClient) GetUserPoints(ctx context.Context, userID string) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/points/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetUserStreak(ctx context.Context, userID string) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/streaks/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("lesson service base URL is not configured")
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

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}, nil
}
