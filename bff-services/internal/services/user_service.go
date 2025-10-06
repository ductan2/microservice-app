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
)

type UserService interface {
	Register(ctx context.Context, payload dto.RegisterRequest) (*HTTPResponse, error)
	Login(ctx context.Context, payload dto.LoginRequest, userAgent, clientIP string) (*HTTPResponse, error)
	Logout(ctx context.Context, token string) (*HTTPResponse, error)
	VerifyEmail(ctx context.Context, token string) (*HTTPResponse, error)
	GetProfile(ctx context.Context, token string) (*HTTPResponse, error)
	UpdateProfile(ctx context.Context, token string, payload dto.UpdateProfileRequest) (*HTTPResponse, error)
	CheckAuth(ctx context.Context, token string) (*HTTPResponse, error)
	RequestPasswordReset(ctx context.Context, payload dto.PasswordResetRequest) (*HTTPResponse, error)
	ConfirmPasswordReset(ctx context.Context, payload dto.PasswordResetConfirmRequest) (*HTTPResponse, error)
	ChangePassword(ctx context.Context, token string, payload dto.ChangePasswordRequest) (*HTTPResponse, error)
	SetupMFA(ctx context.Context, token string, payload dto.MFASetupRequest) (*HTTPResponse, error)
	VerifyMFA(ctx context.Context, token string, payload dto.MFAVerifyRequest) (*HTTPResponse, error)
	DisableMFA(ctx context.Context, token string, payload dto.MFADisableRequest) (*HTTPResponse, error)
	GetMFAMethods(ctx context.Context, token string) (*HTTPResponse, error)
	GetSessions(ctx context.Context, token string) (*HTTPResponse, error)
	DeleteSession(ctx context.Context, token, sessionID string) (*HTTPResponse, error)
	RevokeAllSessions(ctx context.Context, token string) (*HTTPResponse, error)
	GetUsers(ctx context.Context, page, pageSize, status, search string) (*HTTPResponse, error)
	// New methods for internal communication with user context
	GetProfileWithContext(ctx context.Context, userID, email, sessionID string) (*HTTPResponse, error)
	UpdateProfileWithContext(ctx context.Context, userID, email, sessionID string, payload dto.UpdateProfileRequest) (*HTTPResponse, error)
}

type UserServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
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
	return c.doRequest(ctx, http.MethodPost, "/api/v1/register", payload, nil)
}

func (c *UserServiceClient) Login(ctx context.Context, payload dto.LoginRequest, userAgent, clientIP string) (*HTTPResponse, error) {
	headers := http.Header{}
	if userAgent != "" {
		headers.Set("User-Agent", userAgent)
	}
	if clientIP != "" {
		headers.Set("X-Forwarded-For", clientIP)
	}
	return c.doRequest(ctx, http.MethodPost, "/api/v1/login", payload, headers)
}

func (c *UserServiceClient) Logout(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/logout", nil, authHeader(token))
}

func (c *UserServiceClient) VerifyEmail(ctx context.Context, token string) (*HTTPResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("verification token is required")
	}

	path := "/api/v1/verify-email?token=" + url.QueryEscape(token)
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *UserServiceClient) GetProfile(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/profile", nil, authHeader(token))
}

func (c *UserServiceClient) UpdateProfile(ctx context.Context, token string, payload dto.UpdateProfileRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPut, "/api/v1/profile", payload, authHeader(token))
}

func (c *UserServiceClient) CheckAuth(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/profile/check-auth", nil, authHeader(token))
}

func (c *UserServiceClient) RequestPasswordReset(ctx context.Context, payload dto.PasswordResetRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/password/reset/request", payload, nil)
}

func (c *UserServiceClient) ConfirmPasswordReset(ctx context.Context, payload dto.PasswordResetConfirmRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/password/reset/confirm", payload, nil)
}

func (c *UserServiceClient) ChangePassword(ctx context.Context, token string, payload dto.ChangePasswordRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/password/change", payload, authHeader(token))
}

func (c *UserServiceClient) SetupMFA(ctx context.Context, token string, payload dto.MFASetupRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/mfa/setup", payload, authHeader(token))
}

func (c *UserServiceClient) VerifyMFA(ctx context.Context, token string, payload dto.MFAVerifyRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/mfa/verify", payload, authHeader(token))
}

func (c *UserServiceClient) DisableMFA(ctx context.Context, token string, payload dto.MFADisableRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/mfa/disable", payload, authHeader(token))
}

func (c *UserServiceClient) GetMFAMethods(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/mfa/methods", nil, authHeader(token))
}

func (c *UserServiceClient) GetSessions(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/sessions", nil, authHeader(token))
}

func (c *UserServiceClient) DeleteSession(ctx context.Context, token, sessionID string) (*HTTPResponse, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	path := "/api/v1/sessions/" + sessionID
	return c.doRequest(ctx, http.MethodDelete, path, nil, authHeader(token))
}

func (c *UserServiceClient) RevokeAllSessions(ctx context.Context, token string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/sessions/revoke-all", nil, authHeader(token))
}

func (c *UserServiceClient) GetUsers(ctx context.Context, page, pageSize, status, search string) (*HTTPResponse, error) {
	path := "/api/v1/users"
	query := url.Values{}
	if page != "" {
		query.Add("page", page)
	}
	if pageSize != "" {
		query.Add("page_size", pageSize)
	}
	if status != "" {
		query.Add("status", status)
	}
	if search != "" {
		query.Add("search", search)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *UserServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("user service base URL is not configured")
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

func (r HTTPResponse) IsBodyEmpty() bool {
	return len(bytes.TrimSpace(r.Body)) == 0
}

func authHeader(token string) http.Header {
	if token == "" {
		return nil
	}
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	return header
}

// internalAuthHeaders creates headers for internal microservice communication
func internalAuthHeaders(userID, email, sessionID string) http.Header {
	header := http.Header{}
	header.Set("X-User-ID", userID)
	header.Set("X-User-Email", email)
	header.Set("X-Session-ID", sessionID)
	return header
}

func (c *UserServiceClient) GetProfileWithContext(ctx context.Context, userID, email, sessionID string) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/profile", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) UpdateProfileWithContext(ctx context.Context, userID, email, sessionID string, payload dto.UpdateProfileRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPut, "/api/v1/profile", payload, internalAuthHeaders(userID, email, sessionID))
}
