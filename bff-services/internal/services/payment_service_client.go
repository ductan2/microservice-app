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

// PaymentService defines payment-related operations proxied to order-service.
type PaymentService interface {
	CreatePaymentIntent(ctx context.Context, token, userID, email, sessionID, orderID string, payload dto.CreatePaymentIntentRequest) (*types.HTTPResponse, error)
	ConfirmPayment(ctx context.Context, token string, paymentIntentID string, payload dto.ConfirmPaymentRequest) (*types.HTTPResponse, error)
	GetPaymentByOrderID(ctx context.Context, token, userID, email, sessionID, orderID string) (*types.HTTPResponse, error)
	GetPaymentMethods(ctx context.Context, token, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetPaymentHistory(ctx context.Context, token, userID, email, sessionID string, query dto.PaymentHistoryQuery) (*types.HTTPResponse, error)
	GetStripeConfig(ctx context.Context) (*types.HTTPResponse, error)
}

type PaymentServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPaymentServiceClient returns a client for payment endpoints.
func NewPaymentServiceClient(baseURL string, httpClient *http.Client) *PaymentServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &PaymentServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *PaymentServiceClient) CreatePaymentIntent(ctx context.Context, token, userID, email, sessionID, orderID string, payload dto.CreatePaymentIntentRequest) (*types.HTTPResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order id is required")
	}
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	path := "/api/v1/orders/" + url.PathEscape(orderID) + "/pay"
	return c.doRequest(ctx, http.MethodPost, path, payload, headers)
}

func (c *PaymentServiceClient) ConfirmPayment(ctx context.Context, token, paymentIntentID string, payload dto.ConfirmPaymentRequest) (*types.HTTPResponse, error) {
	if paymentIntentID == "" {
		return nil, fmt.Errorf("payment intent id is required")
	}
	headers := bearerAuthHeader(token)
	path := "/api/v1/payments/" + url.PathEscape(paymentIntentID) + "/confirm"
	return c.doRequest(ctx, http.MethodPost, path, payload, headers)
}

func (c *PaymentServiceClient) GetPaymentByOrderID(ctx context.Context, token, userID, email, sessionID, orderID string) (*types.HTTPResponse, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order id is required")
	}
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	path := "/api/v1/orders/" + url.PathEscape(orderID) + "/payment"
	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

func (c *PaymentServiceClient) GetPaymentMethods(ctx context.Context, token, userID, email, sessionID string) (*types.HTTPResponse, error) {
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	return c.doRequest(ctx, http.MethodGet, "/api/v1/payment-methods", nil, headers)
}

func (c *PaymentServiceClient) GetPaymentHistory(ctx context.Context, token, userID, email, sessionID string, query dto.PaymentHistoryQuery) (*types.HTTPResponse, error) {
	headers := internalAuthHeaders(userID, email, sessionID)
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	path := "/api/v1/payments"
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
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

func (c *PaymentServiceClient) GetStripeConfig(ctx context.Context) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/stripe/config", nil, nil)
}

func (c *PaymentServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
