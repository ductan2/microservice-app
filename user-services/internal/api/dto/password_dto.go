package dto

// PasswordResetRequestDTO initiates password reset
type PasswordResetRequestDTO struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmDTO completes password reset
type PasswordResetConfirmDTO struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePasswordRequest for authenticated users
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
