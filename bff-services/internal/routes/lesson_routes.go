package routes

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupLessonRoutes configures lesson and progress-related routes
func SetupLessonRoutes(api *gin.RouterGroup, controllers *controllers.Controllers, sessionCache *cache.SessionCache) {
	if controllers == nil || controllers.Lesson == nil || sessionCache == nil {
		return
	}

	// Progress tracking routes (protected)
	progress := api.Group("/progress")
	progress.Use(middleware.AuthRequired(sessionCache))
	{
		// Daily activity tracking
		daily := progress.Group("/daily-activity")
		{
			daily.GET("/user/me/today", controllers.Lesson.GetDailyActivityToday)
			daily.GET("/user/me/date/:activity_date", controllers.Lesson.GetDailyActivityByDate)
			daily.GET("/user/me/range", controllers.Lesson.GetDailyActivityRange)
			daily.GET("/user/me/week", controllers.Lesson.GetDailyActivityWeek)
			daily.GET("/user/me/month", controllers.Lesson.GetDailyActivityMonth)
			daily.GET("/user/me/stats/summary", controllers.Lesson.GetDailyActivitySummary)
			daily.POST("/increment", controllers.Lesson.IncrementDailyActivity)
		}

		// Streak tracking
		streaks := progress.Group("/streaks")
		{
			streaks.GET("/user/me", controllers.Lesson.GetMyStreak)
			streaks.POST("/user/me/check", controllers.Lesson.CheckMyStreak)
			streaks.GET("/user/me/status", controllers.Lesson.GetMyStreakStatus)
			streaks.GET("/leaderboard", controllers.Lesson.GetStreakLeaderboard)
			streaks.GET("/user/:user_id", controllers.Lesson.GetStreakByUserID)
		}
	}

	// Leaderboard routes (protected)
	leaderboards := api.Group("/leaderboards")
	leaderboards.Use(middleware.AuthRequired(sessionCache))
	{
		leaderboards.GET("/weekly/current", controllers.Lesson.GetCurrentWeeklyLeaderboard)
		leaderboards.GET("/monthly/current", controllers.Lesson.GetCurrentMonthlyLeaderboard)
		leaderboards.GET("/weekly/history", controllers.Lesson.GetWeeklyLeaderboardHistory)
		leaderboards.GET("/monthly/history", controllers.Lesson.GetMonthlyLeaderboardHistory)
		leaderboards.GET("/user/me/history", controllers.Lesson.GetUserLeaderboardHistory)
		leaderboards.GET("/week/:week_key", controllers.Lesson.GetWeekLeaderboard)
		leaderboards.GET("/month/:month_key", controllers.Lesson.GetMonthLeaderboard)
	}

	// Course enrollment routes (protected)
	courseEnrollments := api.Group("/course-enrollments")
	courseEnrollments.Use(middleware.AuthRequired(sessionCache))
	{
		courseEnrollments.GET("/me", controllers.Lesson.ListMyEnrollments)
		courseEnrollments.POST("", controllers.Lesson.EnrollCourse)
		courseEnrollments.GET("/:enrollment_id", controllers.Lesson.GetEnrollment)
		courseEnrollments.PUT("/:enrollment_id", controllers.Lesson.UpdateEnrollment)
		courseEnrollments.POST("/:enrollment_id/cancel", controllers.Lesson.CancelEnrollment)
	}

	// Course lesson routes (public)
	courseLessons := api.Group("/course-lessons")
	{
		courseLessons.GET("/by-course/:course_id", controllers.Lesson.ListCourseLessonsByCourseID)
		courseLessons.POST("", controllers.Lesson.CreateCourseLesson)
		courseLessons.PUT("/:row_id", controllers.Lesson.UpdateCourseLesson)
		courseLessons.DELETE("/:row_id", controllers.Lesson.DeleteCourseLesson)
	}
}
