package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"user-services/internal/api/dto"
	"user-services/internal/api/services"
	"user-services/internal/types"
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
		ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
			Success: false,
			Message: "Unauthorized: User ID not found in context",
			Error:   "unauthorized",
		})
		return
	}

	userIDUUID := userID.(uuid.UUID)

	var req dto.StartSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	session, err := c.service.StartActivitySession(ctx, &req, userIDUUID)
	if err != nil {
		if _, ok := err.(*services.ValidationError); ok {
			ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
				Success: false,
				Message: "Validation failed",
				Error:   err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
			Success: false,
			Message: "Failed to start session",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.HTTPResponse{
		Success: true,
		Message: "Session started successfully",
		Data:    session,
	})
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
		ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
			Success: false,
			Message: "Unauthorized: User ID not found in context",
			Error:   "unauthorized",
		})
		return
	}

	userIDUUID := userID.(uuid.UUID)

	var req dto.EndSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	session, err := c.service.EndActivitySession(ctx, &req, userIDUUID)
	if err != nil {
		if sessionErr, ok := err.(*services.SessionError); ok {
			switch sessionErr.Code {
			case "SESSION_NOT_FOUND":
				ctx.JSON(http.StatusNotFound, types.HTTPResponse{
					Success: false,
					Message: "Active session not found",
					Error:   sessionErr.Error(),
				})
				return
			case "UNAUTHORIZED":
				ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
					Success: false,
					Message: "Unauthorized to access this session",
					Error:   sessionErr.Error(),
				})
				return
			default:
				ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
					Success: false,
					Message: "Failed to end session",
					Error:   sessionErr.Error(),
				})
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
			Success: false,
			Message: "Failed to end session",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.HTTPResponse{
		Success: true,
		Message: "Session ended successfully",
		Data:    session,
	})
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
		ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
			Success: false,
			Message: "Unauthorized: User ID not found in context",
			Error:   "unauthorized",
		})
		return
	}

	userIDUUID := userID.(uuid.UUID)

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
		ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
			Success: false,
			Message: "Failed to get sessions",
			Error:   err.Error(),
		})
		return
	}

	// Create pagination metadata
	metadata := map[string]interface{}{
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	}

	ctx.JSON(http.StatusOK, types.HTTPResponse{
		Success:  true,
		Message:  "Sessions retrieved successfully",
		Data:     sessions,
		Metadata: metadata,
	})
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
		ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
			Success: false,
			Message: "Unauthorized: User ID not found in context",
			Error:   "unauthorized",
		})
		return
	}

	userIDUUID := userID.(uuid.UUID)

	stats, err := c.service.GetSessionStats(ctx, userIDUUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
			Success: false,
			Message: "Failed to get session statistics",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.HTTPResponse{
		Success: true,
		Message: "Session statistics retrieved successfully",
		Data:    stats,
	})
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
		ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
			Success: false,
			Message: "Unauthorized: User ID not found in context",
			Error:   "unauthorized",
		})
		return
	}

	userIDUUID := userID.(uuid.UUID)

	var req dto.UpdateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, types.HTTPResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	err := c.service.UpdateActiveSession(ctx, req.SessionID, req.UserAgent, req.IPAddr, userIDUUID)
	if err != nil {
		if sessionErr, ok := err.(*services.SessionError); ok {
			switch sessionErr.Code {
			case "SESSION_NOT_FOUND":
				ctx.JSON(http.StatusNotFound, types.HTTPResponse{
					Success: false,
					Message: "Active session not found",
					Error:   sessionErr.Error(),
				})
				return
			case "UNAUTHORIZED":
				ctx.JSON(http.StatusUnauthorized, types.HTTPResponse{
					Success: false,
					Message: "Unauthorized to access this session",
					Error:   sessionErr.Error(),
				})
				return
			default:
				ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
					Success: false,
					Message: "Failed to update session",
					Error:   sessionErr.Error(),
				})
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, types.HTTPResponse{
			Success: false,
			Message: "Failed to update session",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, types.HTTPResponse{
		Success: true,
		Message: "Session updated successfully",
	})
}