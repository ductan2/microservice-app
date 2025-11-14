package controllers

import (
	"net/http"
	"strconv"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// NotificationController handles notification and template operations.
type NotificationController struct {
	notificationService services.NotificationService
}

// NewNotificationController constructs a new NotificationController.
func NewNotificationController(notificationService services.NotificationService) *NotificationController {
	return &NotificationController{notificationService: notificationService}
}

func (n *NotificationController) CreateTemplate(c *gin.Context) {
	var req dto.CreateNotificationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := n.notificationService.CreateTemplate(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, "Unable to create notification template", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) GetAllTemplates(c *gin.Context) {
	resp, err := n.notificationService.GetAllTemplates(c.Request.Context())
	if err != nil {
		utils.Fail(c, "Unable to get notification templates", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) GetTemplateById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Fail(c, "Template ID is required", http.StatusBadRequest, "missing template id")
		return
	}

	resp, err := n.notificationService.GetTemplateById(c.Request.Context(), id)
	if err != nil {
		utils.Fail(c, "Unable to get notification template", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Fail(c, "Template ID is required", http.StatusBadRequest, "missing template id")
		return
	}

	var req dto.UpdateNotificationTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := n.notificationService.UpdateTemplate(c.Request.Context(), id, req)
	if err != nil {
		utils.Fail(c, "Unable to update notification template", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.Fail(c, "Template ID is required", http.StatusBadRequest, "missing template id")
		return
	}

	resp, err := n.notificationService.DeleteTemplate(c.Request.Context(), id)
	if err != nil {
		utils.Fail(c, "Unable to delete notification template", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// User notification endpoints
func (n *NotificationController) CreateUserNotification(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user id")
		return
	}

	var req dto.CreateUserNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := n.notificationService.CreateUserNotification(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, "Unable to create user notification", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) GetUserNotifications(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user id")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	isReadStr := c.Query("is_read")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a number")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be a number")
		return
	}

	var isRead *bool
	if isReadStr != "" {
		parsed, err := strconv.ParseBool(isReadStr)
		if err != nil {
			utils.Fail(c, "Invalid is_read parameter", http.StatusBadRequest, "is_read must be true or false")
			return
		}
		isRead = &parsed
	}

	resp, err := n.notificationService.GetUserNotifications(c.Request.Context(), userID, limit, offset, isRead)
	if err != nil {
		utils.Fail(c, "Unable to get user notifications", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) MarkNotificationsAsRead(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user id")
		return
	}

	var req dto.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := n.notificationService.MarkNotificationsAsRead(c.Request.Context(), userID, req)
	if err != nil {
		utils.Fail(c, "Unable to mark notifications as read", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) GetUnreadCount(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user id")
		return
	}

	resp, err := n.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		utils.Fail(c, "Unable to get unread count", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (n *NotificationController) DeleteUserNotification(c *gin.Context) {
	userID := c.Param("userId")
	notificationID := c.Param("notificationId")

	if userID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user id")
		return
	}
	if notificationID == "" {
		utils.Fail(c, "Notification ID is required", http.StatusBadRequest, "missing notification id")
		return
	}

	resp, err := n.notificationService.DeleteUserNotification(c.Request.Context(), userID, notificationID)
	if err != nil {
		utils.Fail(c, "Unable to delete user notification", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// Bulk operations
func (n *NotificationController) SendNotificationToUsers(c *gin.Context) {
	templateID := c.Param("templateId")
	if templateID == "" {
		utils.Fail(c, "Template ID is required", http.StatusBadRequest, "missing template id")
		return
	}

	var req dto.SendNotificationToUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := n.notificationService.SendNotificationToUsers(c.Request.Context(), templateID, req)
	if err != nil {
		utils.Fail(c, "Unable to send notifications to users", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
