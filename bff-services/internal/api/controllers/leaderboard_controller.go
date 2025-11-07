package controllers

import (
	"net/http"
	"strconv"

	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type LeaderboardController struct {
	leaderboardService services.LeaderboardService
}

func NewLeaderboardController(leaderboardService services.LeaderboardService) *LeaderboardController {
	return &LeaderboardController{leaderboardService: leaderboardService}
}

func (l *LeaderboardController) GetCurrentWeeklyLeaderboard(c *gin.Context) {
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

	resp, err := l.leaderboardService.GetCurrentWeeklyLeaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch weekly leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LeaderboardController) GetCurrentMonthlyLeaderboard(c *gin.Context) {
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

	resp, err := l.leaderboardService.GetCurrentMonthlyLeaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch monthly leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LeaderboardController) GetWeeklyLeaderboardHistory(c *gin.Context) {
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

	resp, err := l.leaderboardService.GetWeeklyLeaderboardHistory(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch weekly leaderboard history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LeaderboardController) GetMonthlyLeaderboardHistory(c *gin.Context) {
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

	resp, err := l.leaderboardService.GetMonthlyLeaderboardHistory(c.Request.Context(), limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch monthly leaderboard history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LeaderboardController) GetUserLeaderboardHistory(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := l.leaderboardService.GetUserLeaderboardHistory(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user leaderboard history", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LeaderboardController) GetWeekLeaderboard(c *gin.Context) {
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

	resp, err := l.leaderboardService.GetWeekLeaderboard(c.Request.Context(), weekKey, limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch week leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (l *LeaderboardController) GetMonthLeaderboard(c *gin.Context) {
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

	resp, err := l.leaderboardService.GetMonthLeaderboard(c.Request.Context(), monthKey, limit, offset)
	if err != nil {
		utils.Fail(c, "Unable to fetch month leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
