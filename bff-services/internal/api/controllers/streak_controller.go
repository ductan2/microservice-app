package controllers

import (
	"io"
	"net/http"
	"strconv"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type StreakController struct {
	streakService services.StreakService
}

func NewStreakController(streakService services.StreakService) *StreakController {
	return &StreakController{streakService: streakService}
}

func (s *StreakController) GetMyStreak(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := s.streakService.GetMyStreak(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user streak", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *StreakController) CheckMyStreak(c *gin.Context) {
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

	resp, err := s.streakService.CheckMyStreak(c.Request.Context(), userID, email, sessionID, payload)
	if err != nil {
		utils.Fail(c, "Unable to check streak", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *StreakController) GetMyStreakStatus(c *gin.Context) {
	userID, email, sessionID, ok := middleware.GetUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := s.streakService.GetMyStreakStatus(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch streak status", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *StreakController) GetStreakLeaderboard(c *gin.Context) {
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

	resp, err := s.streakService.GetStreakLeaderboard(c.Request.Context(), userID, email, sessionID, limit)
	if err != nil {
		utils.Fail(c, "Unable to fetch streak leaderboard", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (s *StreakController) GetStreakByUserID(c *gin.Context) {
	if _, _, _, ok := middleware.GetUserContextFromMiddleware(c); !ok {
		return
	}

	targetID := c.Param("user_id")
	if targetID == "" {
		utils.Fail(c, "User ID is required", http.StatusBadRequest, "missing user_id path parameter")
		return
	}

	resp, err := s.streakService.GetUserStreak(c.Request.Context(), targetID)
	if err != nil {
		utils.Fail(c, "Unable to fetch user streak", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}
