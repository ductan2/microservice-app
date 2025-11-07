package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"bff-services/internal/types"
)

// LeaderboardService defines the contract for leaderboard operations.
type LeaderboardService interface {
	GetCurrentWeeklyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetCurrentMonthlyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetWeeklyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetMonthlyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error)
	GetUserLeaderboardHistory(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetWeekLeaderboard(ctx context.Context, weekKey string, limit *int, offset int) (*types.HTTPResponse, error)
	GetMonthLeaderboard(ctx context.Context, monthKey string, limit *int, offset int) (*types.HTTPResponse, error)
}

// LeaderboardServiceClient implements LeaderboardService against a remote HTTP REST endpoint.
type LeaderboardServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLeaderboardServiceClient constructs a new LeaderboardServiceClient.
func NewLeaderboardServiceClient(baseURL string, httpClient *http.Client) *LeaderboardServiceClient {
	return &LeaderboardServiceClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *LeaderboardServiceClient) GetCurrentWeeklyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/weekly/current"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LeaderboardServiceClient) GetCurrentMonthlyLeaderboard(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/monthly/current"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LeaderboardServiceClient) GetWeeklyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/weekly/history"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LeaderboardServiceClient) GetMonthlyLeaderboardHistory(ctx context.Context, limit int, offset int) (*types.HTTPResponse, error) {
	path := "/api/v1/leaderboards/monthly/history"
	query := url.Values{}
	if limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LeaderboardServiceClient) GetUserLeaderboardHistory(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/leaderboards/user/me/history", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *LeaderboardServiceClient) GetWeekLeaderboard(ctx context.Context, weekKey string, limit *int, offset int) (*types.HTTPResponse, error) {
	if weekKey == "" {
		return nil, fmt.Errorf("week key is required")
	}
	path := "/api/v1/leaderboards/week/" + weekKey
	query := url.Values{}
	if limit != nil && *limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LeaderboardServiceClient) GetMonthLeaderboard(ctx context.Context, monthKey string, limit *int, offset int) (*types.HTTPResponse, error) {
	if monthKey == "" {
		return nil, fmt.Errorf("month key is required")
	}
	path := "/api/v1/leaderboards/month/" + monthKey
	query := url.Values{}
	if limit != nil && *limit > 0 {
		query.Add("limit", fmt.Sprintf("%d", *limit))
	}
	if offset > 0 {
		query.Add("offset", fmt.Sprintf("%d", offset))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *LeaderboardServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
