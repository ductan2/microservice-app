package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupNotificationRoutes configures notification-related routes
func SetupNotificationRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Notification == nil || sessionCache == nil {
		return
	}

	notifications := api.Group("/notifications")
	{
		// Notification template routes
		templates := notifications.Group("/templates")
		{
			templates.GET("", controllers.Notification.GetAllTemplates)
			templates.GET("/:id", controllers.Notification.GetTemplateById)
			templates.POST("", controllers.Notification.CreateTemplate)
			templates.PUT("/:id", controllers.Notification.UpdateTemplate)
			templates.DELETE("/:id", controllers.Notification.DeleteTemplate)
			templates.POST("/:templateId/send", controllers.Notification.SendNotificationToUsers)
		}
	}

	// Protected user notification routes
	userNotifications := api.Group("/notifications/users/:userId")
	userNotifications.Use(middleware.AuthRequired(sessionCache))
	{
		userNotifications.GET("/notifications", controllers.Notification.GetUserNotifications)
		userNotifications.POST("/notifications", controllers.Notification.CreateUserNotification)
		userNotifications.PUT("/notifications/read", controllers.Notification.MarkNotificationsAsRead)
		userNotifications.GET("/notifications/unread-count", controllers.Notification.GetUnreadCount)
		userNotifications.DELETE("/notifications/:notificationId", controllers.Notification.DeleteUserNotification)
	}
}
