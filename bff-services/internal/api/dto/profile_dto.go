package dto

// UpdateProfileRequest represents payload for updating user profile.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	AvatarURL   string `json:"avatar_url" binding:"omitempty,url"`
	Locale      string `json:"locale" binding:"omitempty"`
	TimeZone    string `json:"time_zone" binding:"omitempty"`
}
