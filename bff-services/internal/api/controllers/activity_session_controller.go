package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"
)

// ActivitySessionController handles activity session operations.
type ActivitySessionController struct {
	userService services.UserService
}

// NewActivitySessionController constructs a new ActivitySessionController.
func NewActivitySessionController(userService services.UserService) *ActivitySessionController {
	return &ActivitySessionController{
		userService: userService,
	}
}

func (a *ActivitySessionController) StartSession(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utils.Fail(c, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := a.userService.StartActivitySession(c, req, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to start session", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (a *ActivitySessionController) EndSession(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.EndSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utils.Fail(c, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := a.userService.EndActivitySession(c, req, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to end session", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (a *ActivitySessionController) GetSessions(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		start, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			startDate = &start
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		end, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			endDate = &end
		}
	}

	resp, err := a.userService.GetActivitySessions(c, userID, email, sessionID, page, limit, startDate, endDate)
	if err != nil {
		utils.Fail(c, "Unable to fetch sessions", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (a *ActivitySessionController) GetSessionStats(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := a.userService.GetSessionStats(c, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch session statistics", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (a *ActivitySessionController) UpdateSession(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utils.Fail(c, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := a.userService.UpdateActivitySession(c, req, userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to update session", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
