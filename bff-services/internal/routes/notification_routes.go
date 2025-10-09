package routes

import (
	"bff-services/internal/api/controllers"

	"github.com/gin-gonic/gin"
)

// SetupNotificationRoutes configures notification-related routes
func SetupNotificationRoutes(api *gin.RouterGroup, controllers *controllers.Controllers) {
	if controllers.Notification == nil {
		return
	}

	// Notification template routes
	templates := api.Group("/notifications/templates")
	{
		templates.POST("", controllers.Notification.CreateTemplate)
		templates.GET("", controllers.Notification.GetAllTemplates)
		templates.GET("/:id", controllers.Notification.GetTemplateById)
		templates.PUT("/:id", controllers.Notification.UpdateTemplate)
		templates.DELETE("/:id", controllers.Notification.DeleteTemplate)
	}

	// User notification routes
	userNotifications := api.Group("/notifications/users/:userId/notifications")
	{
		userNotifications.POST("", controllers.Notification.CreateUserNotification)
		userNotifications.GET("", controllers.Notification.GetUserNotifications)
		userNotifications.PUT("/read", controllers.Notification.MarkNotificationsAsRead)
		userNotifications.GET("/unread-count", controllers.Notification.GetUnreadCount)
		userNotifications.DELETE("/:notificationId", controllers.Notification.DeleteUserNotification)
	}

	// Bulk operations
	api.POST("/notifications/templates/:templateId/send", controllers.Notification.SendNotificationToUsers)
}
