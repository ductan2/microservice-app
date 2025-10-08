package dto

// CreateNotificationTemplateRequest represents payload for creating a notification template
type CreateNotificationTemplateRequest struct {
	Type  string                 `json:"type" binding:"required"`
	Title string                 `json:"title" binding:"required"`
	Body  string                 `json:"body" binding:"required"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

// UpdateNotificationTemplateRequest represents payload for updating a notification template
type UpdateNotificationTemplateRequest struct {
	Type  *string                 `json:"type,omitempty"`
	Title *string                 `json:"title,omitempty"`
	Body  *string                 `json:"body,omitempty"`
	Data  *map[string]interface{} `json:"data,omitempty"`
}

// CreateUserNotificationRequest represents payload for creating a user notification
type CreateUserNotificationRequest struct {
	NotificationID string `json:"notification_id" binding:"required"`
}

// MarkAsReadRequest represents payload for marking notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []string `json:"notification_ids" binding:"required,min=1"`
}

// SendNotificationToUsersRequest represents payload for sending notifications to multiple users
type SendNotificationToUsersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
}

// NotificationTemplateResponse represents a notification template
type NotificationTemplateResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
}

// NotificationTemplateWithCountResponse represents a notification template with user count
type NotificationTemplateWithCountResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
	UserCount int                    `json:"user_count"`
}

// UserNotificationResponse represents a user notification
type UserNotificationResponse struct {
	ID             string                       `json:"id"`
	UserID         string                       `json:"user_id"`
	NotificationID string                       `json:"notification_id"`
	IsRead         bool                         `json:"is_read"`
	CreatedAt      string                       `json:"created_at"`
	ReadAt         *string                      `json:"read_at,omitempty"`
	Template       NotificationTemplateResponse `json:"template"`
}

// UnreadCountResponse represents unread notification count
type UnreadCountResponse struct {
	UnreadCount int `json:"unread_count"`
}

// BulkNotificationResponse represents response for bulk notification operations
type BulkNotificationResponse struct {
	NotificationsCreated int                        `json:"notifications_created"`
	Notifications        []UserNotificationResponse `json:"notifications"`
}

// UpdatedCountResponse represents response for update operations
type UpdatedCountResponse struct {
	UpdatedCount int `json:"updated_count"`
}
