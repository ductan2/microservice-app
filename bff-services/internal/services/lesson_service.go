package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/types"
)

// LessonService defines the contract for interacting with the lesson service API.
type LessonService interface {
	GetUserPoints(ctx context.Context, userID string) (*types.HTTPResponse, error)
	GetUserStreak(ctx context.Context, userID string) (*types.HTTPResponse, error)
	GetDailyActivityToday(ctx context.Context, token string) (*types.HTTPResponse, error)
	GetDailyActivityByDate(ctx context.Context, token, activityDate string) (*types.HTTPResponse, error)
	GetDailyActivityRange(ctx context.Context, token, dateFrom, dateTo string) (*types.HTTPResponse, error)
	GetDailyActivityWeek(ctx context.Context, token string) (*types.HTTPResponse, error)
	GetDailyActivityMonth(ctx context.Context, token, year, month string) (*types.HTTPResponse, error)
	GetDailyActivitySummary(ctx context.Context, token string) (*types.HTTPResponse, error)
	IncrementDailyActivity(ctx context.Context, token string, payload dto.DailyActivityIncrementRequest) (*types.HTTPResponse, error)
	GetUserPreferences(ctx context.Context, token string) (*types.HTTPResponse, error)
	CreateUserPreferences(ctx context.Context, token string, payload dto.DimUserCreateRequest) (*types.HTTPResponse, error)
	UpdateUserPreferences(ctx context.Context, token string, payload dto.DimUserUpdateRequest) (*types.HTTPResponse, error)
	UpdateUserLocale(ctx context.Context, token string, payload dto.DimUserLocaleUpdateRequest) (*types.HTTPResponse, error)
	DeleteUserPreferences(ctx context.Context, token string) (*types.HTTPResponse, error)
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

func (c *LessonServiceClient) GetUserPoints(ctx context.Context, userID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/points/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetUserStreak(ctx context.Context, userID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/streaks/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LessonServiceClient) GetDailyActivityToday(ctx context.Context, token string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/daily-activity/user/me/today", nil, authHeader(token))
}

func (c *LessonServiceClient) GetDailyActivityByDate(ctx context.Context, token, activityDate string) (*types.HTTPResponse, error) {
	if activityDate == "" {
		return nil, fmt.Errorf("activity date is required")
	}
	path := "/api/daily-activity/user/me/date/" + activityDate
	return c.doRequest(ctx, http.MethodGet, path, nil, authHeader(token))
}

func (c *LessonServiceClient) GetDailyActivityRange(ctx context.Context, token, dateFrom, dateTo string) (*types.HTTPResponse, error) {
	path := "/api/daily-activity/user/me/range"
	query := url.Values{}
	if dateFrom != "" {
		query.Add("date_from", dateFrom)
	}
	if dateTo != "" {
		query.Add("date_to", dateTo)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, authHeader(token))
}

func (c *LessonServiceClient) GetDailyActivityWeek(ctx context.Context, token string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/daily-activity/user/me/week", nil, authHeader(token))
}

func (c *LessonServiceClient) GetDailyActivityMonth(ctx context.Context, token, year, month string) (*types.HTTPResponse, error) {
	path := "/api/daily-activity/user/me/month"
	query := url.Values{}
	if year != "" {
		query.Add("year", year)
	}
	if month != "" {
		query.Add("month", month)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, authHeader(token))
}

func (c *LessonServiceClient) GetDailyActivitySummary(ctx context.Context, token string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/daily-activity/user/me/stats/summary", nil, authHeader(token))
}

func (c *LessonServiceClient) IncrementDailyActivity(ctx context.Context, token string, payload dto.DailyActivityIncrementRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/daily-activity/increment", payload, authHeader(token))
}

func (c *LessonServiceClient) GetUserPreferences(ctx context.Context, token string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/users/me", nil, authHeader(token))
}

func (c *LessonServiceClient) CreateUserPreferences(ctx context.Context, token string, payload dto.DimUserCreateRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/users", payload, authHeader(token))
}

func (c *LessonServiceClient) UpdateUserPreferences(ctx context.Context, token string, payload dto.DimUserUpdateRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPut, "/api/users/me", payload, authHeader(token))
}

func (c *LessonServiceClient) UpdateUserLocale(ctx context.Context, token string, payload dto.DimUserLocaleUpdateRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPatch, "/api/users/me/locale", payload, authHeader(token))
}

func (c *LessonServiceClient) DeleteUserPreferences(ctx context.Context, token string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, "/api/users/me", nil, authHeader(token))
}

func (c *LessonServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
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

	return &types.HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}, nil
}

func authHeader(token string) http.Header {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	return header
}
