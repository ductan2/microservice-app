package helpers

import (
	"user-services/internal/api/dto"
	"user-services/internal/models"
)

func ToPublicUser(user models.User) dto.PublicUser {

	publicUser := dto.PublicUser{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		Role:          user.Role,
		LastLoginAt:   user.LastLoginAt.Time,
		LastLoginIP:   *user.LastLoginIP,
		LockoutUntil:  user.LockoutUntil.Time,
		DeletedAt:     user.DeletedAt.Time,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	if user.Profile.UserID != (user.ID) || user.Profile.DisplayName != "" || user.Profile.AvatarURL != "" {
		publicUser.Profile = &dto.UserProfile{
			DisplayName: user.Profile.DisplayName,
			AvatarURL:   user.Profile.AvatarURL,
			Locale:      user.Profile.Locale,
			TimeZone:    user.Profile.TimeZone,
			UpdatedAt:   user.Profile.UpdatedAt,
		}
	}

	return publicUser
}
