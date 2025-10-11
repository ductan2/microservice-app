package dto

// UpdateProfileRequest represents payload for updating user profile.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty" binding:"omitempty,min=2"`
	AvatarURL   string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	Locale      string `json:"locale,omitempty" binding:"omitempty,len=2"`
	TimeZone    string `json:"time_zone,omitempty"`
}

type UserWithProgressResponse struct {
	ID            string      `json:"id"`
	Email         string      `json:"email"`
	Status        string      `json:"status"`
	Role          string      `json:"role"`
	CreatedAt     string      `json:"created_at"`
	LastLoginAt   string      `json:"last_login_at"`
	LastLoginIP   string      `json:"last_login_ip"`
	LockoutUntil  string      `json:"lockout_until"`
	DeletedAt     string      `json:"deleted_at"`
	Profile       UserProfile `json:"profile"`
	EmailVerified bool        `json:"email_verified"`
	Points        int         `json:"points"`
	Streak        int         `json:"streak"`
}

type UserData struct {
	ID            string       `json:"id"`
	Email         string       `json:"email"`
	Status        string       `json:"status"`
	Role          string       `json:"role"`
	CreatedAt     string       `json:"created_at"`
	LastLoginAt   string       `json:"last_login_at"`
	LastLoginIP   string       `json:"last_login_ip"`
	LockoutUntil  string       `json:"lockout_until"`
	DeletedAt     string       `json:"deleted_at"`
	Profile       *UserProfile `json:"profile"`
	EmailVerified bool         `json:"email_verified"`
	UpdatedAt     string       `json:"updated_at"`
}

// UserProfile for non-auth data
type UserProfile struct {
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Locale      string `json:"locale"`
	TimeZone    string `json:"time_zone"`
	UpdatedAt   string `json:"updated_at"`
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

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=student teacher admin super-admin"`
}