package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user-services/internal/api/dto"
	"user-services/internal/api/services"
	"user-services/internal/utils"
)

// ActivitySessionController handles activity session HTTP requests
type ActivitySessionController struct {
	service services.ActivitySessionService
}

// NewActivitySessionController creates a new activity session controller
func NewActivitySessionController(service services.ActivitySessionService) *ActivitySessionController {
	return &ActivitySessionController{
		service: service,
	}
}

// StartSession starts a user activity session
// @Summary Start Activity Session
// @Description Start tracking user activity time for a session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.StartSessionRequest true "Session start data"
// @Success 200 {object} types.HTTPResponse{data=dto.ActivitySessionResponse}
// @Failure 400 {object} types.HTTPResponse
// @Failure 401 {object} types.HTTPResponse
// @Failure 404 {object} types.HTTPResponse
// @Failure 500 {object} types.HTTPResponse
// @Router /sessions/start [post]
func (c *ActivitySessionController) StartSession(ctx *gin.Context) {
	// Get userID from context (set by InternalAuthRequired middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, "Unauthorized: User ID not found in context", http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized: Invalid user ID type in context", http.StatusUnauthorized, "invalid user ID type")
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

	session, err := c.service.StartActivitySession(ctx, &req, userIDUUID, ctx)
	if err != nil {
		utils.Fail(ctx, "Failed to start session", http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(ctx, session)
}

// EndSession ends a user activity session
// @Summary End Activity Session
// @Description Stop tracking user activity time and calculate duration
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.EndSessionRequest true "Session end data"
// @Success 200 {object} types.HTTPResponse{data=dto.ActivitySessionResponse}
// @Failure 400 {object} types.HTTPResponse
// @Failure 401 {object} types.HTTPResponse
// @Failure 404 {object} types.HTTPResponse
// @Failure 500 {object} types.HTTPResponse
// @Router /sessions/end [post]
func (c *ActivitySessionController) EndSession(ctx *gin.Context) {
	// Get userID from context (set by InternalAuthRequired middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, "Unauthorized: User ID not found in context", http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized: Invalid user ID type in context", http.StatusUnauthorized, "invalid user ID type")
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

	session, err := c.service.EndActivitySession(ctx, &req, userIDUUID)
	if err != nil {
		utils.Fail(ctx, "Failed to end session", http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(ctx, session)
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
// @Success 200 {object} types.HTTPResponse{data=[]dto.ActivitySessionResponse}
// @Failure 400 {object} types.HTTPResponse
// @Failure 401 {object} types.HTTPResponse
// @Failure 500 {object} types.HTTPResponse
// @Router /sessions [get]
func (c *ActivitySessionController) GetSessions(ctx *gin.Context) {
	// Get userID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, "Unauthorized: User ID not found in context", http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized: Invalid user ID type in context", http.StatusUnauthorized, "invalid user ID type")
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

	sessions, total, err := c.service.GetActivitySessions(ctx, userIDUUID, page, limit, startDate, endDate)
	if err != nil {
		utils.Fail(ctx, "Failed to get sessions", http.StatusInternalServerError, err.Error())
		return
	}

	// Create pagination metadata
	responseData := map[string]any{
		"sessions": sessions,
		"metadata": map[string]any{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	utils.Success(ctx, responseData)
}

// GetSessionStats retrieves session statistics for a user
// @Summary Get Session Statistics
// @Description Get aggregated statistics about user activity sessions
// @Tags sessions
// @Accept json
// @Produce json
// @Success 200 {object} types.HTTPResponse{data=dto.SessionStatsResponse}
// @Failure 401 {object} types.HTTPResponse
// @Failure 500 {object} types.HTTPResponse
// @Router /sessions/stats [get]
func (c *ActivitySessionController) GetSessionStats(ctx *gin.Context) {
	// Get userID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, "Unauthorized: User ID not found in context", http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized: Invalid user ID type in context", http.StatusUnauthorized, "invalid user ID type")
		return
	}

	stats, err := c.service.GetSessionStats(ctx, userIDUUID)
	if err != nil {
		utils.Fail(ctx, "Failed to get session statistics", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, stats)
}

// UpdateSession updates session information for an active session
// @Summary Update Session Information
// @Description Update user agent and IP address for an active session
// @Tags sessions
// @Accept json
// @Produce json
// @Param session body dto.UpdateSessionRequest true "Update session data"
// @Success 200 {object} types.HTTPResponse
// @Failure 400 {object} types.HTTPResponse
// @Failure 401 {object} types.HTTPResponse
// @Failure 404 {object} types.HTTPResponse
// @Failure 500 {object} types.HTTPResponse
// @Router /sessions/update [post]
func (c *ActivitySessionController) UpdateSession(ctx *gin.Context) {
	// Get userID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, "Unauthorized: User ID not found in context", http.StatusUnauthorized, "unauthorized")
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized: Invalid user ID type in context", http.StatusUnauthorized, "invalid user ID type")
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

	err := c.service.UpdateActiveSession(ctx, req.SessionID, req.UserAgent, req.IPAddr, userIDUUID, ctx)
	if err != nil {
		if err.Error() == "active session not found" {
			utils.Fail(ctx, "Active session not found", http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "unauthorized to access this session" {
			utils.Fail(ctx, "Unauthorized to access this session", http.StatusUnauthorized, err.Error())
			return
		}

		utils.Fail(ctx, "Failed to update session", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, map[string]string{"message": "Session updated successfully"})
}