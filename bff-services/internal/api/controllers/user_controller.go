package controllers

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"bff-services/internal/api/dto"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"
	"bff-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	userService   services.UserService
	lessonService services.LessonService
}

func NewUserController(userService services.UserService, lessonService services.LessonService) *UserController {
	return &UserController{
		userService:   userService,
		lessonService: lessonService,
	}
}

// Authentication methods
func (u *UserController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := u.userService.Register(c.Request.Context(), req)
	if err != nil {
		utils.Fail(c, "Unable to register user", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (u *UserController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := u.userService.Login(c.Request.Context(), req, c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		utils.Fail(c, "Unable to login", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (u *UserController) Logout(c *gin.Context) {
	token, ok := requireBearerToken(c)
	if !ok {
		return
	}

	resp, err := u.userService.Logout(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to logout", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (u *UserController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.Fail(c, "Verification token is required", http.StatusBadRequest, "missing token")
		return
	}

	resp, err := u.userService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		utils.Fail(c, "Unable to verify email", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// Profile methods
func (u *UserController) GetProfile(c *gin.Context) {
	userID, email, sessionID, ok := getUserContextFromMiddleware(c)
	if !ok {
		return
	}

	resp, err := u.userService.GetProfileWithContext(c.Request.Context(), userID, email, sessionID)
	if err != nil {
		utils.Fail(c, "Unable to fetch profile", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

func (u *UserController) UpdateProfile(c *gin.Context) {
	userID, email, sessionID, ok := getUserContextFromMiddleware(c)
	if !ok {
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	resp, err := u.userService.UpdateProfileWithContext(c.Request.Context(), userID, email, sessionID, req)
	if err != nil {
		utils.Fail(c, "Unable to update profile", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(c, resp)
}

// Users management methods
func (u *UserController) AssignRoleToUser(ctx *gin.Context) {
	actorID, actorEmail, sessionID, ok := getUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	targetUserID := ctx.Param("id")
	if targetUserID == "" {
		utils.Fail(ctx, "User ID is required", http.StatusBadRequest, "missing user ID")
		return
	}
	if _, err := uuid.Parse(targetUserID); err != nil {
		utils.Fail(ctx, "Invalid user ID", http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UserRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}
	req.RoleName = strings.TrimSpace(req.RoleName)
	if req.RoleName == "" {
		utils.Fail(ctx, "Role name is required", http.StatusBadRequest, "role_name is required")
		return
	}

	resp, err := u.userService.AssignRoleWithContext(ctx.Request.Context(), actorID, actorEmail, sessionID, targetUserID, req)
	if err != nil {
		utils.Fail(ctx, "Failed to assign role", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

func (u *UserController) RemoveRoleFromUser(ctx *gin.Context) {
	actorID, actorEmail, sessionID, ok := getUserContextFromMiddleware(ctx)
	if !ok {
		return
	}

	targetUserID := ctx.Param("id")
	if targetUserID == "" {
		utils.Fail(ctx, "User ID is required", http.StatusBadRequest, "missing user ID")
		return
	}
	if _, err := uuid.Parse(targetUserID); err != nil {
		utils.Fail(ctx, "Invalid user ID", http.StatusBadRequest, err.Error())
		return
	}

	var req dto.UserRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}
	req.RoleName = strings.TrimSpace(req.RoleName)
	if req.RoleName == "" {
		utils.Fail(ctx, "Role name is required", http.StatusBadRequest, "role_name is required")
		return
	}

	resp, err := u.userService.RemoveRoleWithContext(ctx.Request.Context(), actorID, actorEmail, sessionID, targetUserID, req)
	if err != nil {
		utils.Fail(ctx, "Failed to remove role", http.StatusBadGateway, err.Error())
		return
	}

	respondWithServiceResponse(ctx, resp)
}

func (u *UserController) ListUsersWithProgress(ctx *gin.Context) {
	// Get query parameters
	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("page_size", "20")
	status := ctx.Query("status")
	search := ctx.Query("search")

	// Call user service to get list of users
	userResp, err := u.userService.GetUsers(ctx.Request.Context(), page, pageSize, status, search)
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
			Page       int            `json:"page"`
			PageSize   int            `json:"page_size"`
			Total      int            `json:"total"`
			TotalPages int            `json:"total_pages"`
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
			if pointsResp, err := u.lessonService.GetUserPoints(ctx.Request.Context(), userData.ID); err == nil && pointsResp.StatusCode == http.StatusOK {
				var pointsData dto.PointsData
				if err := json.Unmarshal(pointsResp.Body, &pointsData); err == nil {
					points = pointsData.Lifetime
				}
			}

			// Fetch streak
			if streakResp, err := u.lessonService.GetUserStreak(ctx.Request.Context(), userData.ID); err == nil && streakResp.StatusCode == http.StatusOK {
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

func (u *UserController) GetUserById(ctx *gin.Context) {
	// Extract user context from middleware
	userIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "user context not found")
		return
	}
	emailValue, exists := ctx.Get(middleware.ContextUserEmailKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "email context not found")
		return
	}
	sessionIDValue, exists := ctx.Get(middleware.ContextSessionIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "session context not found")
		return
	}

	userID := normalizeUUIDOrString(userIDValue)
	email := normalizeString(emailValue)
	sessionID := normalizeUUIDOrString(sessionIDValue)

	// Call user service using internal headers
	userResp, err := u.userService.GetProfileWithContext(ctx.Request.Context(), userID, email, sessionID)
	if err != nil || userResp == nil {
		utils.Fail(ctx, "Failed to fetch profile", http.StatusBadGateway, errString(err))
		return
	}
	if userResp.StatusCode != http.StatusOK {
		ctx.Data(userResp.StatusCode, "application/json", userResp.Body)
		return
	}
	// Fetch points and streak concurrently, but don't reshape; return raw bodies
	var (
		pointsBody json.RawMessage
		streakBody json.RawMessage
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if resp, err := u.lessonService.GetUserPoints(ctx.Request.Context(), userID); err == nil && resp.StatusCode == http.StatusOK {
			pointsBody = json.RawMessage(resp.Body)
		}
	}()
	go func() {
		defer wg.Done()
		if resp, err := u.lessonService.GetUserStreak(ctx.Request.Context(), userID); err == nil && resp.StatusCode == http.StatusOK {
			streakBody = json.RawMessage(resp.Body)
		}
	}()
	wg.Wait()

	// Extract only inner data from user-service envelope {status,data}
	var userEnvelope map[string]interface{}
	_ = json.Unmarshal(userResp.Body, &userEnvelope)
	userInner := userEnvelope["data"]

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user":   userInner,
			"points": pointsBody,
			"streak": streakBody,
		},
	})
}

// Helper functions
func getUserContextFromMiddleware(c *gin.Context) (userID, email, sessionID string, ok bool) {
	userIDValue, exists := c.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "user context not found")
		return "", "", "", false
	}

	emailValue, exists := c.Get(middleware.ContextUserEmailKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "email context not found")
		return "", "", "", false
	}

	sessionIDValue, exists := c.Get(middleware.ContextSessionIDKey())
	if !exists {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "session context not found")
		return "", "", "", false
	}

	// Convert UUID to string for internal communication
	userIDUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid user ID type")
		return "", "", "", false
	}

	emailStr, ok := emailValue.(string)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid email type")
		return "", "", "", false
	}

	sessionIDUUID, ok := sessionIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(c, "Unauthorized", http.StatusUnauthorized, "invalid session ID type")
		return "", "", "", false
	}

	return userIDUUID.String(), emailStr, sessionIDUUID.String(), true
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func normalizeUUIDOrString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case uuid.UUID:
		return t.String()
	default:
		return ""
	}
}

func normalizeString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
