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

// OrderService defines operations backed by the order-service.
type OrderService interface {
	CreateOrder(ctx context.Context, token, userID, email, sessionID string, payload dto.CreateOrderRequest) (*types.HTTPResponse, error)
	ListOrders(ctx context.Context, token, userID, email, sessionID string, query dto.OrderListQuery) (*types.HTTPResponse, error)
	GetOrder(ctx context.Context, token, userID, email, sessionID, orderID string) (*types.HTTPResponse, error)
	CancelOrder(ctx context.Context, token, userID, email, sessionID, orderID string, payload dto.CancelOrderRequest) (*types.HTTPResponse, error)
}

type OrderServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOrderServiceClient builds a client that proxies calls to order-service.
func NewOrderServiceClient(baseURL string, httpClient *http.Client) *OrderServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &OrderServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *OrderServiceClient) CreateOrder(ctx context.Context, token, userID, email, sessionID string, payload dto.CreateOrderRequest) (*types.HTTPResponse, error) {
	headers := c.combineHeaders(token, userID, email, sessionID)
	return c.doRequest(ctx, http.MethodPost, "/api/v1/orders", payload, headers)
}

func (c *OrderServiceClient) ListOrders(ctx context.Context, token, userID, email, sessionID string, query dto.OrderListQuery) (*types.HTTPResponse, error) {
	headers := c.combineHeaders(token, userID, email, sessionID)
	path := "/api/v1/orders"

	params := url.Values{}
	if query.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", query.Limit))
	}
	if query.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", query.Offset))
	}
	if query.Page > 0 {
		params.Set("page", fmt.Sprintf("%d", query.Page))
	}
	if strings.TrimSpace(query.Status) != "" {
		params.Set("status", strings.TrimSpace(query.Status))
	}
	if strings.TrimSpace(query.SortBy) != "" {
		params.Set("sort_by", strings.TrimSpace(query.SortBy))
	}
	if strings.TrimSpace(query.SortOrder) != "" {
		params.Set("sort_order", strings.TrimSpace(query.SortOrder))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

func (c *OrderServiceClient) GetOrder(ctx context.Context, token, userID, email, sessionID, orderID string) (*types.HTTPResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order id is required")
	}
	headers := c.combineHeaders(token, userID, email, sessionID)
	path := "/api/v1/orders/" + url.PathEscape(orderID)
	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

func (c *OrderServiceClient) CancelOrder(ctx context.Context, token, userID, email, sessionID, orderID string, payload dto.CancelOrderRequest) (*types.HTTPResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order id is required")
	}
	headers := c.combineHeaders(token, userID, email, sessionID)
	path := "/api/v1/orders/" + url.PathEscape(orderID) + "/cancel"
	return c.doRequest(ctx, http.MethodPost, path, payload, headers)
}

func (c *OrderServiceClient) combineHeaders(token, userID, email, sessionID string) http.Header {
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	return headers
}

func (c *OrderServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
