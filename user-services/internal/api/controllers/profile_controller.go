package controllers

import (
	"errors"
	"net/http"

	"user-services/internal/api/dto"
	"user-services/internal/api/middleware"
	"user-services/internal/api/services"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileController struct {
	profileService     services.UserProfileService
	currentUserService services.CurrentUserService
}

func NewProfileController(profileService services.UserProfileService, currentUserService services.CurrentUserService) *ProfileController {
	return &ProfileController{
		profileService:     profileService,
		currentUserService: currentUserService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Tags profile
// @Produce json
// @Success 200 {object} dto.UserProfile
// @Router /profile [get]
func (c *ProfileController) GetProfile(ctx *gin.Context) {
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

	utils.Success(ctx, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Tags profile
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update Profile Request"
// @Success 200 {object} dto.UserProfile
// @Router /profile [put]
func (c *ProfileController) UpdateProfile(ctx *gin.Context) {
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

	utils.Success(ctx, profile)
}
