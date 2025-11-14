package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"bff-services/internal/api/dto"
	"bff-services/internal/cache"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/types"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// LessonController handles lesson, activity, streak, and leaderboard operations.
type LessonController struct {
	lessonService      services.LessonService
	streakCacheService *cache.StreakCacheService
}

// NewLessonController constructs a new LessonController.
func NewLessonController(lessonService services.LessonService) *LessonController {
	return &LessonController{lessonService: lessonService}
}

// NewLessonControllerWithCache constructs a new LessonController with caching support.
func NewLessonControllerWithCache(lessonService services.LessonService, streakCacheService *cache.StreakCacheService) *LessonController {
	return &LessonController{
		lessonService:      lessonService,
		streakCacheService: streakCacheService,
	}
}

func (l *LessonController) GetDailyActivityToday(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetDailyActivityToday(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch today's activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetDailyActivityByDate(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	activityDate := c.Param("activity_date")
	if activityDate == "" {
		utils.Fail(c, "activity_date is required", http.StatusBadRequest, "missing activity date")
		return
	}

	resp, err := l.lessonService.GetDailyActivityByDate(c.Request.Context(), userID, email, sessionID, activityDate)
	if err != nil {
		utils.Fail(c, "Unable to fetch activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetDailyActivityRange(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetDailyActivityRange(
		c.Request.Context(),
		userID, email, sessionID,
		c.Query("date_from"),
		c.Query("date_to"),
	)
	if err != nil {
		utils.Fail(c, "Unable to fetch activity range", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetDailyActivityWeek(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetDailyActivityWeek(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch weekly activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetDailyActivityMonth(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetDailyActivityMonth(
		c.Request.Context(),
		userID, email, sessionID,
		c.Query("year"),
		c.Query("month"),
	)
	if err != nil {
		utils.Fail(c, "Unable to fetch monthly activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetDailyActivitySummary(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetDailyActivitySummary(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch activity summary", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) IncrementDailyActivity(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.DailyActivityIncrementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := l.lessonService.IncrementDailyActivity(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to update activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetMyStreak(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}
	var cachedStreak *cache.StreakData
	var cachedActivity []cache.ActivityData
	var cacheErr error

	if l.streakCacheService != nil {
		cachedStreak, cacheErr = l.streakCacheService.GetCachedStreak(c.Request.Context(), userID)
		if cacheErr == nil && cachedStreak != nil {
			cachedActivity, cacheErr = l.streakCacheService.GetCachedWeekActivity(c.Request.Context(), userID)
			if cacheErr == nil && cachedActivity != nil {
				l.respondWithStreakData(c, cachedStreak, cachedActivity)
				return
			}
		}
	}

	type result struct {
		resp *types.HTTPResponse
		err  error
	}

	streakChan := make(chan result, 1)
	weekActivityChan := make(chan result, 1)

	go func() {
		resp, err := l.lessonService.GetMyStreak(c.Request.Context(), userID, email, sessionID)
		streakChan <- result{resp, err}
	}()

	go func() {
		resp, err := l.lessonService.GetDailyActivityWeek(c.Request.Context(), userID, email, sessionID)
		weekActivityChan <- result{resp, err}
	}()

	streakResult := <-streakChan
	weekActivityResult := <-weekActivityChan

	streakResp := streakResult.resp
	errStreak := streakResult.err
	weekActivityResp := weekActivityResult.resp
	errWeek := weekActivityResult.err

	if errStreak != nil || errWeek != nil || streakResp == nil || weekActivityResp == nil || streakResp.StatusCode >= 400 || weekActivityResp.StatusCode >= 400 {
		c.JSON(http.StatusBadGateway, gin.H{
			"status":  "failed",
			"message": "Unable to fetch user streak or week activity",
		})
		return
	}

	var streakPayload struct {
		Status string `json:"status"`
		Data   struct {
			CurrentLen int     `json:"current_len"`
			LongestLen int     `json:"longest_len"`
			LastDay    *string `json:"last_day"`
		} `json:"data"`
	}
	if err := json.Unmarshal(streakResp.Body, &streakPayload); err != nil || strings.ToLower(streakPayload.Status) != "success" {
		c.JSON(http.StatusBadGateway, gin.H{"status": "failed", "message": "Invalid streak response"})
		return
	}

	var weekPayload struct {
		Status string `json:"status"`
		Data   []struct {
			ActivityDate     string `json:"activity_dt"`
			LessonsCompleted int    `json:"lessons_completed"`
			QuizzesCompleted int    `json:"quizzes_completed"`
			Minutes          int    `json:"minutes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(weekActivityResp.Body, &weekPayload); err != nil || strings.ToLower(weekPayload.Status) != "success" {
		c.JSON(http.StatusBadGateway, gin.H{"status": "failed", "message": "Invalid week activity response"})
		return
	}

	// Convert API response to cache format and store in Redis
	streakData := &cache.StreakData{
		CurrentLen: streakPayload.Data.CurrentLen,
		LongestLen: streakPayload.Data.LongestLen,
		LastDay:    streakPayload.Data.LastDay,
	}

	activityData := make([]cache.ActivityData, 0, len(weekPayload.Data))
	for _, d := range weekPayload.Data {
		activityData = append(activityData, cache.ActivityData{
			ActivityDate:     d.ActivityDate,
			LessonsCompleted: d.LessonsCompleted,
			QuizzesCompleted: d.QuizzesCompleted,
			Minutes:          d.Minutes,
		})
	}

	if l.streakCacheService != nil {
		go l.streakCacheService.CacheStreak(c.Request.Context(), userID, streakData)
		go l.streakCacheService.CacheWeekActivity(c.Request.Context(), userID, activityData)
	}

	l.respondWithStreakData(c, streakData, activityData)
}

// respondWithStreakData formats and sends streak response
func (l *LessonController) respondWithStreakData(c *gin.Context, streak *cache.StreakData, activities []cache.ActivityData) {
	activity := make([]map[string]interface{}, 0, len(activities))
	for _, d := range activities {
		completed := d.Minutes >= 10 || d.LessonsCompleted > 0 || d.QuizzesCompleted > 0
		activity = append(activity, map[string]interface{}{
			"date":      d.ActivityDate,
			"completed": completed,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"current_streak": streak.CurrentLen,
			"longest_streak": streak.LongestLen,
			"activity":       activity,
		},
	})
}

func (l *LessonController) CheckMyStreak(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var (
		req     dto.StreakCheckRequest
		payload *dto.StreakCheckRequest
	)

	if c.Request.Body != nil {
		if err := c.ShouldBindJSON(&req); err != nil {
			if err != io.EOF {
				utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
				return
			}
		} else {
			payload = &req
		}
	}

	resp, err := l.lessonService.CheckMyStreak(c.Request.Context(), userID, email, sessionID, payload)
	if err != nil {
		utils.Fail(c, "Unable to check streak", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetMyStreakStatus(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetMyStreakStatus(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch streak status", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetStreakLeaderboard(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	}

	resp, err := l.lessonService.GetStreakLeaderboard(c.Request.Context(), userID, email, sessionID, limit)
	if err != nil {
		utils.Fail(c, "Unable to fetch streak leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetStreakByUserID(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	targetID := c.Param("user_id")
	if targetID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user_id path parameter")
		return
	}

	resp, err := l.lessonService.GetUserStreak(c.Request.Context(), targetID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user streak", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetCurrentWeeklyLeaderboard(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	limit := 100
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsed
	}

	resp, err := l.lessonService.GetCurrentWeeklyLeaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch weekly leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetCurrentMonthlyLeaderboard(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	limit := 100
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsed
	}

	resp, err := l.lessonService.GetCurrentMonthlyLeaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch monthly leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetWeeklyLeaderboardHistory(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 52 {
			parsed = 52
		}
		limit = parsed
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsed
	}

	resp, err := l.lessonService.GetWeeklyLeaderboardHistory(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch weekly leaderboard history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetMonthlyLeaderboardHistory(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	limit := 12
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 60 {
			parsed = 60
		}
		limit = parsed
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsed
	}

	resp, err := l.lessonService.GetMonthlyLeaderboardHistory(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch monthly leaderboard history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetUserLeaderboardHistory(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetUserLeaderboardHistory(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user leaderboard history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetWeekLeaderboard(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	weekKey := c.Param("week_key")
	if weekKey == "" {
		utils.Fail(c, "Week key is required", http.StatusBadRequest, "missing week_key path parameter")
		return
	}

	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = &parsed
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsed
	}

	resp, err := l.lessonService.GetWeekLeaderboard(c.Request.Context(), weekKey, limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch week leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetMonthLeaderboard(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	monthKey := c.Param("month_key")
	if monthKey == "" {
		utils.Fail(c, "Month key is required", http.StatusBadRequest, "missing month_key path parameter")
		return
	}

	var limit *int
	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			utils.Fail(c, "Invalid limit parameter", http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = &parsed
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			utils.Fail(c, "Invalid offset parameter", http.StatusBadRequest, "offset must be non-negative")
			return
		}
		offset = parsed
	}

	resp, err := l.lessonService.GetMonthLeaderboard(c.Request.Context(), monthKey, limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch month leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) ListMyEnrollments(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	status := c.Query("status")
	limit := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	resp, err := l.lessonService.ListMyEnrollments(c.Request.Context(), userID, email, sessionID, status, limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch enrollments", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) EnrollCourse(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}
	var req dto.CourseEnrollmentCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	resp, err := l.lessonService.EnrollCourse(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to enroll", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) GetEnrollment(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}
	enrollmentID := c.Param("enrollment_id")
	if enrollmentID == "" {
		utils.Fail(c, "enrollment_id is required", http.StatusBadRequest, "missing id")
		return
	}
	resp, err := l.lessonService.GetEnrollment(c.Request.Context(), enrollmentID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch enrollment", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) UpdateEnrollment(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}
	enrollmentID := c.Param("enrollment_id")
	if enrollmentID == "" {
		utils.Fail(c, "enrollment_id is required", http.StatusBadRequest, "missing id")
		return
	}
	var req dto.CourseEnrollmentUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	resp, err := l.lessonService.UpdateEnrollment(c.Request.Context(), enrollmentID, userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to update enrollment", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) CancelEnrollment(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}
	enrollmentID := c.Param("enrollment_id")
	if enrollmentID == "" {
		utils.Fail(c, "enrollment_id is required", http.StatusBadRequest, "missing id")
		return
	}
	resp, err := l.lessonService.CancelEnrollment(c.Request.Context(), enrollmentID, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to cancel enrollment", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) ListCourseLessonsByCourseID(c *gin.Context) {
	courseID := c.Param("course_id")
	if courseID == "" {
		utils.Fail(c, "course_id is required", http.StatusBadRequest, "missing course_id")
		return
	}
	resp, err := l.lessonService.ListCourseLessonsByCourseID(c.Request.Context(), courseID)
	if err != nil {
		utils.Fail(c, "Unable to fetch course lessons", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) CreateCourseLesson(c *gin.Context) {
	var req dto.CourseLessonCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	resp, err := l.lessonService.CreateCourseLesson(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, "Unable to create course lesson", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) UpdateCourseLesson(c *gin.Context) {
	rowID := c.Param("row_id")
	if rowID == "" {
		utils.Fail(c, "row_id is required", http.StatusBadRequest, "missing row_id")
		return
	}
	var req dto.CourseLessonUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	resp, err := l.lessonService.UpdateCourseLesson(c.Request.Context(), rowID, req)
	if err != nil {
		utils.Fail(c, "Unable to update course lesson", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}

func (l *LessonController) DeleteCourseLesson(c *gin.Context) {
	rowID := c.Param("row_id")
	if rowID == "" {
		utils.Fail(c, "row_id is required", http.StatusBadRequest, "missing row_id")
		return
	}
	resp, err := l.lessonService.DeleteCourseLesson(c.Request.Context(), rowID)
	if err != nil {
		utils.Fail(c, "Unable to delete course lesson", http.StatusBadGateway, err.Error())
		return
	}
	respondWithServiceResponse(c, resp)
}
