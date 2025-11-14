package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/types"
)

// StreakService defines the contract for streak operations.
type StreakService interface {
	GetUserPoints(ctx context.Context, userID string) (*types.HTTPResponse, error)
	GetUserStreak(ctx context.Context, userID string) (*types.HTTPResponse, error)
	GetMyStreak(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	CheckMyStreak(ctx context.Context, userID, email, sessionID string, payload *dto.StreakCheckRequest) (*types.HTTPResponse, error)
	GetMyStreakStatus(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetStreakLeaderboard(ctx context.Context, userID, email, sessionID string, limit int) (*types.HTTPResponse, error)
}

// StreakServiceClient implements StreakService against a remote HTTP REST endpoint.
type StreakServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewStreakServiceClient constructs a new StreakServiceClient.
func NewStreakServiceClient(baseURL string, httpClient *http.Client) *StreakServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &StreakServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *StreakServiceClient) GetUserPoints(ctx context.Context, userID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/v1/progress/points/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *StreakServiceClient) GetUserStreak(ctx context.Context, userID string) (*types.HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	path := "/api/v1/progress/streaks/user/" + userID
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *StreakServiceClient) GetMyStreak(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/streaks/user/me", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *StreakServiceClient) CheckMyStreak(ctx context.Context, userID, email, sessionID string, payload *dto.StreakCheckRequest) (*types.HTTPResponse, error) {
	var body interface{}
	if payload != nil {
		body = payload
	}
	return c.doRequest(ctx, http.MethodPost, "/api/v1/progress/streaks/user/me/check", body, internalAuthHeaders(userID, email, sessionID))
}

func (c *StreakServiceClient) GetMyStreakStatus(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/progress/streaks/user/me/status", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *StreakServiceClient) GetStreakLeaderboard(ctx context.Context, userID, email, sessionID string, limit int) (*types.HTTPResponse, error) {
	path := "/api/v1/progress/streaks/leaderboard"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *StreakServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
