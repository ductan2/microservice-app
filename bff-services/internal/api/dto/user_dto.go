package dto

// UpdateProfileRequest represents payload for updating user profile.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	AvatarURL   string `json:"avatar_url" binding:"omitempty,url"`
	Locale      string `json:"locale" binding:"omitempty"`
	TimeZone    string `json:"time_zone" binding:"omitempty"`
}

type UserWithProgressResponse struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	Status    string      `json:"status"`
	CreatedAt string      `json:"created_at"`
	Profile   UserProfile `json:"profile"`
	Points    int         `json:"points"`
	Streak    int         `json:"streak"`
}

type UserProfile struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type UserData struct {
	ID            string       `json:"id"`
	Email         string       `json:"email"`
	Status        string       `json:"status"`
	CreatedAt     string       `json:"created_at"`
	Profile       *UserProfile `json:"profile"`
	EmailVerified bool         `json:"email_verified"`
	UpdatedAt     string       `json:"updated_at"`
}

type PointsData struct {
	UserID   string `json:"user_id"`
	Lifetime int    `json:"lifetime"`
	Weekly   int    `json:"weekly"`
	Monthly  int    `json:"monthly"`
}

type StreakData struct {
	UserID     string `json:"user_id"`
	CurrentLen int    `json:"current_len"`
	LongestLen int    `json:"longest_len"`
	LastDay    string `json:"last_day"`
}
