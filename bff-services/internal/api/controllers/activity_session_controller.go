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

// ActivitySessionController handles activity session HTTP requests
type ActivitySessionController struct {
	userService services.UserService
}

// NewActivitySessionController creates a new activity session controller
func NewActivitySessionController(userService services.UserService) *ActivitySessionController {
	return &ActivitySessionController{
		userService: userService,
	}
}

// StartSession starts a user activity session
// @Summary Start Activity Session
// @Description Start tracking user activity time for a session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.StartSessionRequest true "Session start data"
// @Security ApiKeyAuth
// @Success 200 {object} utils.HTTPResponse{data=dto.ActivitySessionResponse}
// @Failure 400 {object} utils.HTTPResponse
// @Failure 401 {object} utils.HTTPResponse
// @Failure 404 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /session/start [post]
func (c *ActivitySessionController) StartSession(ctx *gin.Context) {
	// Get user context from AuthRequired middleware
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	var req dto.StartSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request body", http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utils.Fail(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	response, err := c.userService.StartActivitySession(
		ctx,
		req,
		userID,
		email,
		sessionID,
	)
	if err != nil {
		utils.Fail(ctx, "Failed to start session", http.StatusInternalServerError, err.Error())
		return
	}

	respondWithServiceResponse(ctx, response)
}

// EndSession ends a user activity session
// @Summary End Activity Session
// @Description Stop tracking user activity time and calculate duration
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.EndSessionRequest true "Session end data"
// @Security ApiKeyAuth
// @Success 200 {object} utils.HTTPResponse{data=dto.ActivitySessionResponse}
// @Failure 400 {object} utils.HTTPResponse
// @Failure 401 {object} utils.HTTPResponse
// @Failure 404 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /session/end [post]
func (c *ActivitySessionController) EndSession(ctx *gin.Context) {
	// Get user context from AuthRequired middleware
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	var req dto.EndSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request body", http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utils.Fail(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	response, err := c.userService.EndActivitySession(
		ctx,
		req,
		userID,
		email,
		sessionID,
	)
	if err != nil {
		utils.Fail(ctx, "Failed to end session", http.StatusInternalServerError, err.Error())
		return
	}

	respondWithServiceResponse(ctx, response)
}

// GetSessions retrieves user's activity sessions
// @Summary Get User Activity Sessions
// @Description Get paginated list of user's activity sessions
// @Tags sessions
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page limit" default(20)
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Security ApiKeyAuth
// @Success 200 {object} utils.HTTPResponse{data=[]dto.ActivitySessionResponse}
// @Failure 400 {object} utils.HTTPResponse
// @Failure 401 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /sessions [get]
func (c *ActivitySessionController) GetSessions(ctx *gin.Context) {
	// Get user context from AuthRequired middleware
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	// Parse query parameters
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	var startDate, endDate *time.Time
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		start, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			startDate = &start
		}
	}
	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		end, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			endDate = &end
		}
	}

	response, err := c.userService.GetActivitySessions(
		ctx,
		userID,
		email,
		sessionID,
		page,
		limit,
		startDate,
		endDate,
	)
	if err != nil {
		utils.Fail(ctx, "Failed to get sessions", http.StatusInternalServerError, err.Error())
		return
	}

	respondWithServiceResponse(ctx, response)
}

// GetSessionStats retrieves session statistics for a user
// @Summary Get Session Statistics
// @Description Get aggregated statistics about user activity sessions
// @Tags sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.HTTPResponse{data=dto.SessionStatsResponse}
// @Failure 401 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /sessions/stats [get]
func (c *ActivitySessionController) GetSessionStats(ctx *gin.Context) {
	// Get user context from AuthRequired middleware
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	response, err := c.userService.GetSessionStats(
		ctx,
		userID,
		email,
		sessionID,
	)
	if err != nil {
		utils.Fail(ctx, "Failed to get session statistics", http.StatusInternalServerError, err.Error())
		return
	}

	respondWithServiceResponse(ctx, response)
}

// UpdateSession updates session information for an active session
// @Summary Update Session Information
// @Description Update user agent and IP address for an active session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.UpdateSessionRequest true "Update session data"
// @Security ApiKeyAuth
// @Success 200 {object} utils.HTTPResponse
// @Failure 400 {object} utils.HTTPResponse
// @Failure 401 {object} utils.HTTPResponse
// @Failure 404 {object} utils.HTTPResponse
// @Failure 500 {object} utils.HTTPResponse
// @Router /sessions/update [post]
func (c *ActivitySessionController) UpdateSession(ctx *gin.Context) {
	// Get user context from AuthRequired middleware
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	var req dto.UpdateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request body", http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		utils.Fail(ctx, "Validation failed", http.StatusBadRequest, err.Error())
		return
	}

	response, err := c.userService.UpdateActivitySession(
		ctx,
		req,
		userID,
		email,
		sessionID,
	)
	if err != nil {
		utils.Fail(ctx, "Failed to update session", http.StatusInternalServerError, err.Error())
		return
	}

	respondWithServiceResponse(ctx, response)
}