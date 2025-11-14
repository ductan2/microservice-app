package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupQuizAttemptRoutes configures quiz attempt related routes
func SetupQuizAttemptRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.QuizAttempt == nil || sessionCache == nil {
		return
	}

	// Protected quiz attempt routes
	attempts := api.Group("/quiz-attempts")
	attempts.Use(middleware.AuthRequired(sessionCache))
	{
		attempts.POST("/start", controllers.QuizAttempt.StartQuizAttempt)
		attempts.GET("/:attempt_id", controllers.QuizAttempt.GetQuizAttempt)
		attempts.POST("/:attempt_id/submit", controllers.QuizAttempt.SubmitQuizAttempt)
		attempts.GET("/user/:user_id", controllers.QuizAttempt.GetQuizAttemptsByUserID)
		attempts.GET("/user/me/quiz/:quiz_id", controllers.QuizAttempt.GetUserQuizAttempts)
		attempts.GET("/user/me/history", controllers.QuizAttempt.GetUserQuizHistory)
		attempts.GET("/lesson/:lesson_id/user/me", controllers.QuizAttempt.GetLessonQuizAttempts)
		attempts.DELETE("/:attempt_id", controllers.QuizAttempt.DeleteQuizAttempt)
	}
}
