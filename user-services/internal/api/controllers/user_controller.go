package controllers

import (
	"errors"
	"net/http"
	"strings"

	"user-services/internal/api/dto"
	"user-services/internal/api/middleware"
	"user-services/internal/api/services"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	authService        *services.AuthService
	profileService     services.UserProfileService
	currentUserService services.CurrentUserService
	userService        services.UserService
	sessionService     services.SessionService
}

func NewUserController(
	authService *services.AuthService,
	profileService services.UserProfileService,
	currentUserService services.CurrentUserService,
	userService services.UserService,
	sessionService services.SessionService,
) *UserController {
	return &UserController{
		authService:        authService,
		profileService:     profileService,
		currentUserService: currentUserService,
		userService:        userService,
		sessionService:     sessionService,
	}
}

// RegisterUser handles user registration
// POST /users/register
func (c *UserController) RegisterUser(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	result, err := c.authService.Register(ctx.Request.Context(), email, req.Password, req.Name)
	if err != nil {
		utils.Fail(ctx, "Failed to register", http.StatusBadRequest, err.Error())
		return
	}

	// Return success message without token - user needs to verify email
	utils.Created(ctx, gin.H{
		"message": "Registration successful! Please check your email to verify your account.",
		"email":   result.User.Email,
	})
}

// LoginUser handles user login
// POST /users/login
func (c *UserController) LoginUser(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	userAgent := ctx.GetHeader("User-Agent")
	ipAddr := ctx.ClientIP()

	result, err := c.authService.Login(ctx.Request.Context(), email, req.Password, req.MFACode, userAgent, ipAddr)
	if err != nil {
		// Handle specific error types
		switch err {
		case utils.ErrInvalidMFACode:
			utils.Fail(ctx, "Invalid MFA code", http.StatusUnauthorized, err.Error())
		case utils.ErrEmailNotVerified:
			utils.Fail(ctx, "Email not verified", http.StatusUnauthorized, err.Error())
		case utils.ErrInvalidCredentials:
			utils.Fail(ctx, "Invalid email or password", http.StatusUnauthorized, err.Error())
		default:
			utils.Fail(ctx, "Internal server error", http.StatusInternalServerError, err.Error())
		}
		return
	}

	response := dto.AuthResponse{
		AccessToken:  result.Token,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
		User:         modelToPublicUser(result.User),
	}
	utils.Success(ctx, response)
}

// LogoutUser handles user logout
// POST /users/logout
func (c *UserController) LogoutUser(ctx *gin.Context) {
	utils.Success(ctx, gin.H{"message": "Logged out successfully"})
}

// VerifyUserEmail handles email verification
// GET /users/verify-email
func (c *UserController) VerifyUserEmail(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		utils.Fail(ctx, "Verification token is required", http.StatusBadRequest, "missing token")
		return
	}

	if err := c.authService.VerifyEmail(ctx.Request.Context(), token); err != nil {
		utils.Fail(ctx, "Email verification failed", http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(ctx, gin.H{
		"message": "Email verified successfully! You can now login.",
	})
}

// GetUserProfile gets the current user's profile (combines auth + profile data)
// GET /users/profile
func (c *UserController) GetUserProfile(ctx *gin.Context) {
	userIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "invalid user context")
		return
	}

	// Return combined {user + profile} from users table with preload
	profile, err := c.currentUserService.GetPublicUserByID(ctx.Request.Context(), userID.String())
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve user", http.StatusInternalServerError, err.Error())
		return
	}

	sessions, err := c.sessionService.GetUserSessions(ctx.Request.Context(), userID)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve user sessions", http.StatusInternalServerError, err.Error())
		return
	}

	if sessionIDValue, exists := ctx.Get(middleware.ContextSessionIDKey()); exists {
		if currentSessionID, ok := sessionIDValue.(uuid.UUID); ok {
			for i := range sessions {
				if sessions[i].ID == currentSessionID {
					sessions[i].IsCurrent = true
				}
			}
		}
	}

	profile.Sessions = sessions

	utils.Success(ctx, profile)
}

// UpdateUserProfile updates the current user's profile
// PUT /users/profile
func (c *UserController) UpdateUserProfile(ctx *gin.Context) {
	userIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "invalid user context")
		return
	}

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if err := c.profileService.UpdateProfile(ctx.Request.Context(), userID, &req); err != nil {
		if errors.Is(err, services.ErrProfileNotFound) {
			utils.Fail(ctx, "Profile not found", http.StatusNotFound, nil)
			return
		}

		utils.Fail(ctx, "Failed to update profile", http.StatusInternalServerError, err.Error())
		return
	}

	// Return combined {user + profile}
	profile, err := c.currentUserService.GetPublicUserByID(ctx.Request.Context(), userID.String())
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve user", http.StatusInternalServerError, err.Error())
		return
	}

	sessions, err := c.sessionService.GetUserSessions(ctx.Request.Context(), userID)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve user sessions", http.StatusInternalServerError, err.Error())
		return
	}

	if sessionIDValue, exists := ctx.Get(middleware.ContextSessionIDKey()); exists {
		if currentSessionID, ok := sessionIDValue.(uuid.UUID); ok {
			for i := range sessions {
				if sessions[i].ID == currentSessionID {
					sessions[i].IsCurrent = true
				}
			}
		}
	}

	profile.Sessions = sessions

	utils.Success(ctx, profile)
}

// ListAllUsers lists all users with pagination (admin function)
// GET /users
func (c *UserController) ListAllUsers(ctx *gin.Context) {
	userIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "invalid user context")
		return
	}
	if userID.String() == "" {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	isAdmin, err := c.currentUserService.IsAdmin(ctx.Request.Context(), userID.String())
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve user", http.StatusInternalServerError, err.Error())
		return
	}
	if !isAdmin {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	var req dto.ListUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.Fail(ctx, "Invalid request parameters", http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.userService.ListUsers(ctx.Request.Context(), req)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve users", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, result)
}

// GetUserByID gets a specific user by ID (combines auth + profile data)
// GET /users/:id
func (c *UserController) GetUserByID(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		utils.Fail(ctx, "User ID is required", http.StatusBadRequest, "missing user ID")
		return
	}
	// Return combined {user + profile} from users table with preload
	user, err := c.currentUserService.GetPublicUserByID(ctx.Request.Context(), userID)
	if err != nil {
		utils.Fail(ctx, "Failed to retrieve user", http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(ctx, user)
}

// Helper function to convert model to public user DTO
func modelToPublicUser(user models.User) dto.PublicUser {
	return dto.PublicUser{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		Role:          user.Role,
		CreatedAt:     user.CreatedAt,
		Profile: &dto.UserProfile{
			DisplayName: user.Profile.DisplayName,
			AvatarURL:   user.Profile.AvatarURL,
			Locale:      user.Profile.Locale,
			TimeZone:    user.Profile.TimeZone,
		},
		UpdatedAt: user.UpdatedAt,
	}
}

// UpdateUserRole updates a user's role (admin only; internal auth)
// PUT /users/:id/role
func (c *UserController) UpdateUserRole(ctx *gin.Context) {
	actorIDValue, exists := ctx.Get(middleware.ContextUserIDKey())
	if !exists {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	actorID, ok := actorIDValue.(uuid.UUID)
	if !ok {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "invalid user context")
		return
	}

	targetID := ctx.Param("id")
	if targetID == "" {
		utils.Fail(ctx, "User ID is required", http.StatusBadRequest, "missing user ID")
		return
	}

	var req dto.UpdateUserRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	// Authorization: only admin or super-admin can update roles
	actor, err := c.currentUserService.GetPublicUserByID(ctx.Request.Context(), actorID.String())
	if err != nil {
		utils.Fail(ctx, "Unauthorized", http.StatusUnauthorized, "actor not found")
		return
	}
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleSuperAdmin {
		utils.Fail(ctx, "Forbidden", http.StatusForbidden, "insufficient role")
		return
	}
	if req.Role == models.RoleSuperAdmin && actor.Role != models.RoleSuperAdmin {
		utils.Fail(ctx, "Forbidden", http.StatusForbidden, "only super-admin can assign super-admin")
		return
	}

	updated, err := c.userService.UpdateUserRole(ctx.Request.Context(), targetID, req.Role)
	if err != nil {
		utils.Fail(ctx, "Failed to update role", http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(ctx, updated)
}
