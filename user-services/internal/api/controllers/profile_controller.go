package controllers

import (
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	profileService services.UserProfileService
}

func NewProfileController(profileService services.UserProfileService) *ProfileController {
	return &ProfileController{
		profileService: profileService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Tags profile
// @Produce json
// @Success 200 {object} dto.UserProfile
// @Router /profile [get]
func (c *ProfileController) GetProfile(ctx *gin.Context) {
	// TODO: implement
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
	// TODO: implement
}
