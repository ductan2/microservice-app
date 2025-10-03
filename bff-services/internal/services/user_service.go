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

type UserService interface {
	Register(ctx context.Context, payload dto.RegisterRequest) (*HTTPResponse, error)
	Login(ctx context.Context, payload dto.LoginRequest, userAgent, clientIP string) (*HTTPResponse, error)
}

type UserServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type HTTPResponse struct {
	StatusCode int
	Body       Envelope
}

type Envelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

func NewUserServiceClient(baseURL string, httpClient *http.Client) *UserServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &UserServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

func (c *UserServiceClient) Register(ctx context.Context, payload dto.RegisterRequest) (*HTTPResponse, error) {
	return c.doPost(ctx, "/api/v1/register", payload, nil)
}

func (c *UserServiceClient) Login(ctx context.Context, payload dto.LoginRequest, userAgent, clientIP string) (*HTTPResponse, error) {
	headers := http.Header{}
	if userAgent != "" {
		headers.Set("User-Agent", userAgent)
	}
	if clientIP != "" {
		headers.Set("X-Forwarded-For", clientIP)
	}
	return c.doPost(ctx, "/api/v1/login", payload, headers)
}

func (c *UserServiceClient) doPost(ctx context.Context, path string, payload interface{}, headers http.Header) (*HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("user service base URL is not configured")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	endpoint := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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

	var envelope Envelope
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       envelope,
	}, nil
}

func (r HTTPResponse) IsBodyEmpty() bool {
	return r.Body.Status == "" && r.Body.Message == "" && len(r.Body.Data) == 0 && len(r.Body.Error) == 0
}
