package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupLessonRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers.Lesson == nil || sessionCache == nil {
		return
	}

	daily := api.Group("/daily-activity")
	daily.Use(middleware.AuthRequired(sessionCache))
	{
		daily.GET("/user/me/today", controllers.Lesson.GetDailyActivityToday)
		daily.GET("/user/me/date/:activity_date", controllers.Lesson.GetDailyActivityByDate)
		daily.GET("/user/me/range", controllers.Lesson.GetDailyActivityRange)
		daily.GET("/user/me/week", controllers.Lesson.GetDailyActivityWeek)
		daily.GET("/user/me/month", controllers.Lesson.GetDailyActivityMonth)
		daily.GET("/user/me/stats/summary", controllers.Lesson.GetDailyActivitySummary)
		daily.POST("/increment", controllers.Lesson.IncrementDailyActivity)
	}

	dimUser := api.Group("/users")
	dimUser.Use(middleware.AuthRequired(sessionCache))
	{
		dimUser.GET("/me", controllers.Lesson.GetUserPreferences)
		dimUser.POST("", controllers.Lesson.CreateUserPreferences)
		dimUser.PUT("/me", controllers.Lesson.UpdateUserPreferences)
		dimUser.PATCH("/me/locale", controllers.Lesson.UpdateUserLocale)
		dimUser.DELETE("/me", controllers.Lesson.DeleteUserPreferences)
	}
}
