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

type UserService interface {
	Register(ctx context.Context, payload dto.RegisterRequest) (*types.HTTPResponse, error)
	Login(ctx context.Context, payload dto.LoginRequest, userAgent, clientIP string) (*types.HTTPResponse, error)
	Logout(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	VerifyEmail(ctx context.Context, token string) (*types.HTTPResponse, error)
	RequestPasswordReset(ctx context.Context, payload dto.PasswordResetRequest) (*types.HTTPResponse, error)
	ConfirmPasswordReset(ctx context.Context, payload dto.PasswordResetConfirmRequest) (*types.HTTPResponse, error)
	ChangePassword(ctx context.Context, userID, email, sessionID string, payload dto.ChangePasswordRequest) (*types.HTTPResponse, error)
	SetupMFA(ctx context.Context, userID, email, sessionID string, payload dto.MFASetupRequest) (*types.HTTPResponse, error)
	VerifyMFA(ctx context.Context, userID, email, sessionID string, payload dto.MFAVerifyRequest) (*types.HTTPResponse, error)
	DisableMFA(ctx context.Context, userID, email, sessionID string, payload dto.MFADisableRequest) (*types.HTTPResponse, error)
	GetMFAMethods(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetSessions(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	DeleteSession(ctx context.Context, userID, email, sessionID, deleteSessionID string) (*types.HTTPResponse, error)
	RevokeAllSessions(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	ListSessionsByUserID(ctx context.Context, targetUserID, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetUsers(ctx context.Context, page, pageSize, status, search, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetUserById(ctx context.Context, userID, email, sessionID, UserFindID string) (*types.HTTPResponse, error)
	// New methods for internal communication with user context
	GetProfileWithContext(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	UpdateProfileWithContext(ctx context.Context, userID, email, sessionID string, payload dto.UpdateProfileRequest) (*types.HTTPResponse, error)
	UpdateUserRoleWithContext(ctx context.Context, userID, email, sessionID string, targetID string, payload dto.UpdateUserRoleRequest) (*types.HTTPResponse, error)
	LockAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error)
	UnlockAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error)
	SoftDeleteAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error)
	RestoreAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error)
	// Activity session methods
	StartActivitySession(ctx context.Context, payload dto.StartSessionRequest, userID, email, sessionID string) (*types.HTTPResponse, error)
	EndActivitySession(ctx context.Context, payload dto.EndSessionRequest, userID, email, sessionID string) (*types.HTTPResponse, error)
	GetActivitySessions(ctx context.Context, userID, email, sessionID string, page, limit int, startDate, endDate *time.Time) (*types.HTTPResponse, error)
	GetSessionStats(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error)
	UpdateActivitySession(ctx context.Context, payload dto.UpdateSessionRequest, userID, email, sessionID string) (*types.HTTPResponse, error)
}

type UserServiceClient struct {
	baseURL    string
	httpClient *http.Client
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

func (c *UserServiceClient) Register(ctx context.Context, payload dto.RegisterRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/users/register", payload, nil)
}

func (c *UserServiceClient) Login(ctx context.Context, payload dto.LoginRequest, userAgent, clientIP string) (*types.HTTPResponse, error) {
	headers := http.Header{}
	if userAgent != "" {
		headers.Set("User-Agent", userAgent)
	}
	if clientIP != "" {
		headers.Set("X-Forwarded-For", clientIP)
	}
	return c.doRequest(ctx, http.MethodPost, "/api/v1/users/login", payload, headers)
}

func (c *UserServiceClient) Logout(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/users/logout", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) VerifyEmail(ctx context.Context, token string) (*types.HTTPResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("verification token is required")
	}

	path := "/api/v1/users/verify-email?token=" + url.QueryEscape(token)
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *UserServiceClient) RequestPasswordReset(ctx context.Context, payload dto.PasswordResetRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/users/password/reset/request", payload, nil)
}

func (c *UserServiceClient) ConfirmPasswordReset(ctx context.Context, payload dto.PasswordResetConfirmRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/users/password/reset/confirm", payload, nil)
}

func (c *UserServiceClient) ChangePassword(ctx context.Context, userID, email, sessionID string, payload dto.ChangePasswordRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/users/password/change", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) SetupMFA(ctx context.Context, userID, email, sessionID string, payload dto.MFASetupRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/mfa/setup", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) VerifyMFA(ctx context.Context, userID, email, sessionID string, payload dto.MFAVerifyRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/mfa/verify", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) DisableMFA(ctx context.Context, userID, email, sessionID string, payload dto.MFADisableRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/mfa/disable", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetMFAMethods(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/mfa/methods", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetSessions(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/sessions", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) DeleteSession(ctx context.Context, userID, email, sessionID, deleteSessionID string) (*types.HTTPResponse, error) {
	if deleteSessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}
	path := "/api/v1/sessions/" + deleteSessionID
	return c.doRequest(ctx, http.MethodDelete, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) RevokeAllSessions(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/sessions/revoke-all", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) ListSessionsByUserID(ctx context.Context, targetUserID, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/sessions/user/"+targetUserID, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetUsers(ctx context.Context, page, pageSize, status, search, userID, email, sessionID string) (*types.HTTPResponse, error) {
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
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetUserById(ctx context.Context, userID, email, sessionID, UserFindID string) (*types.HTTPResponse, error) {
	path := fmt.Sprintf("/api/v1/users/%s", UserFindID)
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetProfileWithContext(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/users/profile", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) UpdateProfileWithContext(ctx context.Context, userID, email, sessionID string, payload dto.UpdateProfileRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPut, "/api/v1/users/profile", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) UpdateUserRoleWithContext(ctx context.Context, userID, email, sessionID string, targetID string, payload dto.UpdateUserRoleRequest) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v1/users/%s/role", targetID), payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) LockAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error) {
	path := appendReason(fmt.Sprintf("/api/v1/users/%s/lock", targetID), reason)
	return c.doRequest(ctx, http.MethodPost, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) UnlockAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error) {
	path := appendReason(fmt.Sprintf("/api/v1/users/%s/unlock", targetID), reason)
	return c.doRequest(ctx, http.MethodPost, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) SoftDeleteAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error) {
	path := appendReason(fmt.Sprintf("/api/v1/users/%s/delete", targetID), reason)
	return c.doRequest(ctx, http.MethodDelete, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) RestoreAccountWithContext(ctx context.Context, userID, email, sessionID, targetID, reason string) (*types.HTTPResponse, error) {
	path := appendReason(fmt.Sprintf("/api/v1/users/%s/restore", targetID), reason)
	return c.doRequest(ctx, http.MethodPost, path, nil, internalAuthHeaders(userID, email, sessionID))
}

// Activity session methods
func (c *UserServiceClient) StartActivitySession(ctx context.Context, payload dto.StartSessionRequest, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/sessions/start", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) EndActivitySession(ctx context.Context, payload dto.EndSessionRequest, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/sessions/end", payload, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetActivitySessions(ctx context.Context, userID, email, sessionID string, page, limit int, startDate, endDate *time.Time) (*types.HTTPResponse, error) {
	path := "/api/v1/sessions"
	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", page))
	query.Add("limit", fmt.Sprintf("%d", limit))
	if startDate != nil {
		query.Add("start_date", startDate.Format(time.RFC3339))
	}
	if endDate != nil {
		query.Add("end_date", endDate.Format(time.RFC3339))
	}
	path += "?" + query.Encode()
	return c.doRequest(ctx, http.MethodGet, path, nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) GetSessionStats(ctx context.Context, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/v1/sessions/stats", nil, internalAuthHeaders(userID, email, sessionID))
}

func (c *UserServiceClient) UpdateActivitySession(ctx context.Context, payload dto.UpdateSessionRequest, userID, email, sessionID string) (*types.HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/v1/sessions/update", payload, internalAuthHeaders(userID, email, sessionID))
}

func appendReason(path, reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return path
	}

	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}

	return path + sep + "reason=" + url.QueryEscape(reason)
}

func (c *UserServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	return doRequest(ctx, c.baseURL, method, path, c.httpClient, payload, headers)
}
