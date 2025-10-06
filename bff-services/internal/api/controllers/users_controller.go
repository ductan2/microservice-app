package controllers

import (
	"encoding/json"
	"net/http"
	"sync"

	"bff-services/internal/api/dto"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	userService   services.UserService
	lessonService services.LessonService
}

func NewUsersController(userService services.UserService, lessonService services.LessonService) *UsersController {
	return &UsersController{
		userService:   userService,
		lessonService: lessonService,
	}
}


// ListUsersWithProgress returns list of users with their points and streak
func (c *UsersController) ListUsersWithProgress(ctx *gin.Context) {
	// Get query parameters
	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("page_size", "20")
	status := ctx.Query("status")
	search := ctx.Query("search")

	// Call user service to get list of users
	userResp, err := c.userService.GetUsers(ctx.Request.Context(), page, pageSize, status, search)
	if err != nil {
		utils.Fail(ctx, "Failed to fetch users", http.StatusInternalServerError, err.Error())
		return
	}

	if userResp.StatusCode != http.StatusOK {
		ctx.Data(userResp.StatusCode, "application/json", userResp.Body)
		return
	}

	// Parse user service response
	var usersResponse struct {
		Status string `json:"status"`
		Data   struct {
			Data       []dto.UserData `json:"data"`
			Page       int        `json:"page"`
			PageSize   int        `json:"page_size"`
			Total      int        `json:"total"`
			TotalPages int        `json:"total_pages"`
		} `json:"data"`
	}

	if err := json.Unmarshal(userResp.Body, &usersResponse); err != nil {
		utils.Fail(ctx, "Failed to parse users data", http.StatusInternalServerError, err.Error())
		return
	}

	users := usersResponse.Data.Data
	result := make([]dto.UserWithProgressResponse, len(users))

	// Fetch points and streak for each user concurrently
	var wg sync.WaitGroup
	for i, user := range users {
		wg.Add(1)
		go func(index int, userData dto.UserData) {
			defer wg.Done()

			points := 0
			streak := 0

			// Fetch points
			if pointsResp, err := c.lessonService.GetUserPoints(ctx.Request.Context(), userData.ID); err == nil && pointsResp.StatusCode == http.StatusOK {
				var pointsData dto.PointsData
				if err := json.Unmarshal(pointsResp.Body, &pointsData); err == nil {
					points = pointsData.Lifetime
				}
			}

			// Fetch streak
			if streakResp, err := c.lessonService.GetUserStreak(ctx.Request.Context(), userData.ID); err == nil && streakResp.StatusCode == http.StatusOK {
				var streakData dto.StreakData
				if err := json.Unmarshal(streakResp.Body, &streakData); err == nil {
					streak = streakData.CurrentLen
				}
			}

			profile := dto.UserProfile{}
			if userData.Profile != nil {
				profile = dto.UserProfile{
					DisplayName: userData.Profile.DisplayName,
					AvatarURL:   userData.Profile.AvatarURL,
				}
			}

			result[index] = dto.UserWithProgressResponse{
				ID:        userData.ID,
				Email:     userData.Email,
				Status:    userData.Status,
				CreatedAt: userData.CreatedAt,
				Profile:   profile,
				Points:    points,
				Streak:    streak,
			}
		}(i, user)
	}

	wg.Wait()

	// Return aggregated response with pagination info
	response := gin.H{
		"status": "success",
		"data": gin.H{
			"users":       result,
			"page":        usersResponse.Data.Page,
			"page_size":   usersResponse.Data.PageSize,
			"total":       usersResponse.Data.Total,
			"total_pages": usersResponse.Data.TotalPages,
		},
	}

	ctx.JSON(http.StatusOK, response)
}
