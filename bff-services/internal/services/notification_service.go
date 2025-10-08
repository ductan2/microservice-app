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

type NotificationService interface {
	// Template management
	CreateTemplate(ctx context.Context, payload dto.CreateNotificationTemplateRequest) (*HTTPResponse, error)
	GetAllTemplates(ctx context.Context) (*HTTPResponse, error)
	GetTemplateById(ctx context.Context, id string) (*HTTPResponse, error)
	UpdateTemplate(ctx context.Context, id string, payload dto.UpdateNotificationTemplateRequest) (*HTTPResponse, error)
	DeleteTemplate(ctx context.Context, id string) (*HTTPResponse, error)

	// User notifications
	CreateUserNotification(ctx context.Context, userID string, payload dto.CreateUserNotificationRequest) (*HTTPResponse, error)
	GetUserNotifications(ctx context.Context, userID string, limit, offset int, isRead *bool) (*HTTPResponse, error)
	MarkNotificationsAsRead(ctx context.Context, userID string, payload dto.MarkAsReadRequest) (*HTTPResponse, error)
	GetUnreadCount(ctx context.Context, userID string) (*HTTPResponse, error)
	DeleteUserNotification(ctx context.Context, userID, notificationID string) (*HTTPResponse, error)

	// Bulk operations
	SendNotificationToUsers(ctx context.Context, templateID string, payload dto.SendNotificationToUsersRequest) (*HTTPResponse, error)
}

type NotificationServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewNotificationServiceClient(baseURL string, httpClient *http.Client) *NotificationServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &NotificationServiceClient{
		baseURL:    trimmed,
		httpClient: httpClient,
	}
}

// Template management methods
func (c *NotificationServiceClient) CreateTemplate(ctx context.Context, payload dto.CreateNotificationTemplateRequest) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/api/notifications/templates", payload, nil)
}

func (c *NotificationServiceClient) GetAllTemplates(ctx context.Context) (*HTTPResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/api/notifications/templates", nil, nil)
}

func (c *NotificationServiceClient) GetTemplateById(ctx context.Context, id string) (*HTTPResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("template id is required")
	}
	path := "/api/notifications/templates/" + url.PathEscape(id)
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *NotificationServiceClient) UpdateTemplate(ctx context.Context, id string, payload dto.UpdateNotificationTemplateRequest) (*HTTPResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("template id is required")
	}
	path := "/api/notifications/templates/" + url.PathEscape(id)
	return c.doRequest(ctx, http.MethodPut, path, payload, nil)
}

func (c *NotificationServiceClient) DeleteTemplate(ctx context.Context, id string) (*HTTPResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("template id is required")
	}
	path := "/api/notifications/templates/" + url.PathEscape(id)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// User notification methods
func (c *NotificationServiceClient) CreateUserNotification(ctx context.Context, userID string, payload dto.CreateUserNotificationRequest) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	path := "/api/notifications/users/" + url.PathEscape(userID) + "/notifications"
	return c.doRequest(ctx, http.MethodPost, path, payload, nil)
}

func (c *NotificationServiceClient) GetUserNotifications(ctx context.Context, userID string, limit, offset int, isRead *bool) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	path := "/api/notifications/users/" + url.PathEscape(userID) + "/notifications"
	query := url.Values{}
	query.Add("limit", fmt.Sprintf("%d", limit))
	query.Add("offset", fmt.Sprintf("%d", offset))
	if isRead != nil {
		query.Add("is_read", fmt.Sprintf("%t", *isRead))
	}
	path += "?" + query.Encode()

	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *NotificationServiceClient) MarkNotificationsAsRead(ctx context.Context, userID string, payload dto.MarkAsReadRequest) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	path := "/api/notifications/users/" + url.PathEscape(userID) + "/notifications/read"
	return c.doRequest(ctx, http.MethodPut, path, payload, nil)
}

func (c *NotificationServiceClient) GetUnreadCount(ctx context.Context, userID string) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	path := "/api/notifications/users/" + url.PathEscape(userID) + "/notifications/unread-count"
	return c.doRequest(ctx, http.MethodGet, path, nil, nil)
}

func (c *NotificationServiceClient) DeleteUserNotification(ctx context.Context, userID, notificationID string) (*HTTPResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if notificationID == "" {
		return nil, fmt.Errorf("notification id is required")
	}
	path := "/api/notifications/users/" + url.PathEscape(userID) + "/notifications/" + url.PathEscape(notificationID)
	return c.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// Bulk operations
func (c *NotificationServiceClient) SendNotificationToUsers(ctx context.Context, templateID string, payload dto.SendNotificationToUsersRequest) (*HTTPResponse, error) {
	if templateID == "" {
		return nil, fmt.Errorf("template id is required")
	}
	path := "/api/notifications/templates/" + url.PathEscape(templateID) + "/send"
	return c.doRequest(ctx, http.MethodPost, path, payload, nil)
}

func (c *NotificationServiceClient) doRequest(ctx context.Context, method, path string, payload interface{}, headers http.Header) (*HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("notification service base URL is not configured")
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
