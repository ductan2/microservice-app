package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/types"
)

// DailyService defines the contract for daily activity operations.
type DailyService interface {
	GetDailyActivityToday(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetDailyActivityByDate(ctx context.Context, userID, email, sessionID, activityDate string) (*types.HTTPResponse, error)
	GetDailyActivityRange(ctx context.Context, userID, email, sessionID, dateFrom, dateTo string) (*types.HTTPResponse, error)
	GetDailyActivityWeek(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetDailyActivityMonth(ctx context.Context, userID, email, sessionID, year, month string) (*types.HTTPResponse, error)
	GetDailyActivitySummary(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	IncrementDailyActivity(ctx context.Context, userID, email, sessionID string, payload dto.DailyActivityIncrementRequest) (*types.HTTPResponse, error)
}

// DailyServiceClient implements DailyService against a remote HTTP REST endpoint.
type DailyServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewDailyServiceClient constructs a new DailyServiceClient.
func NewDailyServiceClient(baseURL string, httpClient *http.Client) *DailyServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &DailyServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *DailyServiceClient) GetDailyActivityToday(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/daily-activity/user/me/today", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) GetDailyActivityByDate(ctx context.Context, userID, email, sessionID, activityDate string) (*types.HTTPResponse, error) {
	if activityDate == "" {
		return nil, fmt.Errorf("activity date is required")
	}
	path := "/api/v1/progress/daily-activity/user/me/date/" + activityDate
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) GetDailyActivityRange(ctx context.Context, userID, email, sessionID, dateFrom, dateTo string) (*types.HTTPResponse, error) {
	path := "/api/v1/progress/daily-activity/user/me/range"
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
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) GetDailyActivityWeek(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/daily-activity/user/me/week", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) GetDailyActivityMonth(ctx context.Context, userID, email, sessionID, year, month string) (*types.HTTPResponse, error) {
	path := "/api/v1/progress/daily-activity/user/me/month"
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
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) GetDailyActivitySummary(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/daily-activity/user/me/stats/summary", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) IncrementDailyActivity(ctx context.Context, userID, email, sessionID string, payload dto.DailyActivityIncrementRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/progress/daily-activity/increment", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *DailyServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
