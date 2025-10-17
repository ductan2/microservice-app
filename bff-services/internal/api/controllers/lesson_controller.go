package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type LessonController struct {
	lessonService services.LessonService
}

func NewLessonController(lessonService services.LessonService) *LessonController {
	return &LessonController{lessonService: lessonService}
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

func (l *LessonController) GetUserPreferences(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.GetUserPreferences(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user preferences", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) CreateUserPreferences(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.DimUserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}
	req.UserID = userID

	resp, err := l.lessonService.CreateUserPreferences(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to create user preferences", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) UpdateUserPreferences(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.DimUserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	if req.Locale == nil && req.LevelHint == nil {
		utils.Fail(c, "At least one field must be provided", http.StatusBadRequest, "empty update payload")
		return
	}

	resp, err := l.lessonService.UpdateUserPreferences(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to update user preferences", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) UpdateUserLocale(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.DimUserLocaleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := l.lessonService.UpdateUserLocale(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to update user locale", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LessonController) DeleteUserPreferences(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.lessonService.DeleteUserPreferences(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to delete user preferences", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
