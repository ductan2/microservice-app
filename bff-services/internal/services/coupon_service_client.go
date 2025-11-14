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

// CouponService defines coupon-related operations proxied to order-service.
type CouponService interface {
	ListAvailableCoupons(ctx context.Context, token, userID, email, sessionID string, query dto.CouponListQuery) (*types.HTTPResponse, error)
	GetCoupon(ctx context.Context, token, userID, email, sessionID, couponID string) (*types.HTTPResponse, error)
	ValidateCoupon(ctx context.Context, token, userID, email, sessionID string, payload dto.ValidateCouponRequest) (*types.HTTPResponse, error)
	GetUserCouponUsage(ctx context.Context, token, userID, email, sessionID string) (*types.HTTPResponse, error)
}

type CouponServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewCouponServiceClient returns a client for coupon endpoints.
func NewCouponServiceClient(baseURL string, httpClient *http.Client) *CouponServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &CouponServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *CouponServiceClient) ListAvailableCoupons(ctx context.Context, token, userID, email, sessionID string, query dto.CouponListQuery) (*types.HTTPResponse, error) {
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	path := "/api/v1/coupons"
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
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

func (c *CouponServiceClient) GetCoupon(ctx context.Context, token, userID, email, sessionID, couponID string) (*types.HTTPResponse, error) {
	if couponID == "" {
		return nil, fmt.Errorf("coupon id is required")
	}
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	path := "/api/v1/coupons/" + url.PathEscape(couponID)
	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

func (c *CouponServiceClient) ValidateCoupon(ctx context.Context, token, userID, email, sessionID string, payload dto.ValidateCouponRequest) (*types.HTTPResponse, error) {
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	return c.doRequest(ctx, http.MethodPost, "/api/v1/coupons/validate", payload, headers)
}

func (c *CouponServiceClient) GetUserCouponUsage(ctx context.Context, token, userID, email, sessionID string) (*types.HTTPResponse, error) {
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	return c.doRequest(ctx, http.MethodGet, "/api/v1/coupons/usage", nil, headers)
}

func (c *CouponServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
