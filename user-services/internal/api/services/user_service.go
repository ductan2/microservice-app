package services

import (
	"context"
	"math"

	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
	"user-services/internal/models"
)

type UserService interface {
	ListUsers(ctx context.Context, req dto.ListUsersRequest) (*dto.PaginatedResponse, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) ListUsers(ctx context.Context, req dto.ListUsersRequest) (*dto.PaginatedResponse, error) {
	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := s.userRepo.ListUsers(ctx, page, pageSize, req.Status, req.Search)
	if err != nil {
		return nil, err
	}

	// Convert to PublicUser DTOs
	publicUsers := make([]dto.PublicUser, len(users))
	for i, user := range users {
		publicUsers[i] = toPublicUser(user)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &dto.PaginatedResponse{
		Data:       publicUsers,
		Page:       page,
		PageSize:   pageSize,
		Total:      int(total),
		TotalPages: totalPages,
	}, nil
}

func toPublicUser(user models.User) dto.PublicUser {
	publicUser := dto.PublicUser{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	// Include profile if available
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
