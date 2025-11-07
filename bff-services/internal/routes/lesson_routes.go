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

	progress := api.Group("/progress")
	progress.Use(middleware.AuthRequired(sessionCache))

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

	streaks := progress.Group("/streaks")
	{
		streaks.GET("/user/me", controllers.Lesson.GetMyStreak)
		streaks.POST("/user/me/check", controllers.Lesson.CheckMyStreak)
		streaks.GET("/user/me/status", controllers.Lesson.GetMyStreakStatus)
		streaks.GET("/leaderboard", controllers.Lesson.GetStreakLeaderboard)
		streaks.GET("/user/:user_id", controllers.Lesson.GetStreakByUserID)
	}

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

	// Course enrollments (auth required)
	cEnroll := api.Group("/course-enrollments")
	cEnroll.Use(middleware.AuthRequired(sessionCache))
	{
		cEnroll.GET("/me", controllers.Lesson.ListMyEnrollments)
		cEnroll.POST("", controllers.Lesson.EnrollCourse)
		cEnroll.GET("/:enrollment_id", controllers.Lesson.GetEnrollment)
		cEnroll.PUT("/:enrollment_id", controllers.Lesson.UpdateEnrollment)
		cEnroll.POST("/:enrollment_id/cancel", controllers.Lesson.CancelEnrollment)
	}

	// Course lessons (no auth required)
	cLessons := api.Group("/course-lessons")
	{
		cLessons.GET("/by-course/:course_id", controllers.Lesson.ListCourseLessonsByCourseID)
		cLessons.POST("", controllers.Lesson.CreateCourseLesson)
		cLessons.PUT("/:row_id", controllers.Lesson.UpdateCourseLesson)
		cLessons.DELETE("/:row_id", controllers.Lesson.DeleteCourseLesson)
	}
}
