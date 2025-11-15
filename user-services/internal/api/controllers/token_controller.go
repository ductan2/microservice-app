package controllers

import (
	"net/http"
	"strings"

	"user-services/internal/api/dto"
	"user-services/internal/api/middleware"
	"user-services/internal/api/services"
	"user-services/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TokenController struct {
	tokenService services.TokenService
	rateLimiter  middleware.RateLimiter
}

func NewTokenController(tokenService services.TokenService, rateLimiter middleware.RateLimiter) *TokenController {
	return &TokenController{
		tokenService: tokenService,
		rateLimiter:  rateLimiter,
	}
}

// RefreshToken handles token refresh requests
// POST /auth/refresh
func (c *TokenController) RefreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	// Basic validation
	if strings.TrimSpace(req.RefreshToken) == "" {
		utils.Fail(ctx, "Refresh token is required", http.StatusBadRequest, "missing refresh token")
		return
	}

	result, err := c.tokenService.RefreshAccessToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		// Record failed attempt if it's an invalid token (could be brute force)
		if c.rateLimiter != nil && strings.Contains(err.Error(), "invalid") {
			// Use a generic identifier for refresh token attempts since we don't have user context
			c.rateLimiter.RecordFailedAttempt(ctx.Request.Context(), "refresh_token:"+ctx.ClientIP())
		}

		utils.Fail(ctx, "Invalid or expired refresh token", http.StatusUnauthorized, err.Error())
		return
	}

	// Reset failed attempts on successful refresh
	if c.rateLimiter != nil {
		c.rateLimiter.ResetFailedAttempts(ctx.Request.Context(), "refresh_token:"+ctx.ClientIP())
	}

	response := dto.RefreshTokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresAt:    result.ExpiresAt,
	}

	utils.Success(ctx, response)
}

// RevokeToken handles token revocation
// POST /auth/revoke
func (c *TokenController) RevokeToken(ctx *gin.Context) {
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

	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, "Invalid request data", http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		utils.Fail(ctx, "Refresh token is required", http.StatusBadRequest, "missing refresh token")
		return
	}

	err := c.tokenService.RevokeRefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.Fail(ctx, "Failed to revoke token", http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(ctx, gin.H{
		"message": "Token revoked successfully",
		"user_id": userID.String(),
	})
}