package controllers

import (
	"net/http"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

// DailyController handles daily activity tracking operations.
type DailyController struct {
	dailyService services.DailyService
}

// NewDailyController constructs a new DailyController.
func NewDailyController(dailyService services.DailyService) *DailyController {
	return &DailyController{dailyService: dailyService}
}

func (d *DailyController) GetDailyActivityToday(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := d.dailyService.GetDailyActivityToday(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch today's activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (d *DailyController) GetDailyActivityByDate(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	activityDate := c.Param("activity_date")
	if activityDate == "" {
		utils.Fail(c, "activity_date is required", http.StatusBadRequest, "missing activity date")
		return
	}

	resp, err := d.dailyService.GetDailyActivityByDate(c.Request.Context(), userID, email, sessionID, activityDate)
	if err != nil {
		utils.Fail(c, "Unable to fetch activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (d *DailyController) GetDailyActivityRange(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := d.dailyService.GetDailyActivityRange(
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

func (d *DailyController) GetDailyActivityWeek(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := d.dailyService.GetDailyActivityWeek(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch weekly activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (d *DailyController) GetDailyActivityMonth(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := d.dailyService.GetDailyActivityMonth(
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

func (d *DailyController) GetDailyActivitySummary(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := d.dailyService.GetDailyActivitySummary(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch activity summary", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (d *DailyController) IncrementDailyActivity(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.DailyActivityIncrementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request payload", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := d.dailyService.IncrementDailyActivity(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to update activity", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
