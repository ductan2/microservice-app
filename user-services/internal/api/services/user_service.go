package services

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"time"

	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
	"user-services/internal/models"
)

type UserService interface {
	ListUsers(ctx context.Context, req dto.ListUsersRequest) (*dto.PaginatedResponse, error)
	UpdateUserRole(ctx context.Context, userID string, role string) (dto.PublicUser, error)
	LockAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error)
	UnlockAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error)
	SoftDeleteAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error)
	RestoreAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error)
}

var ErrUserDeleted = errors.New("user is deleted")

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
		Role:          user.Role,
		LastLoginAt:   user.LastLoginAt.Time,
		LastLoginIP:   getStringValue(user.LastLoginIP),
		LockoutUntil:  user.LockoutUntil.Time,
		DeletedAt:     user.DeletedAt.Time,
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

// getStringValue safely dereferences a string pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UpdateUserRole updates the role of a target user and returns the updated public user
func (s *userService) UpdateUserRole(ctx context.Context, userID string, role string) (dto.PublicUser, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.PublicUser{}, err
	}

	user.Role = role
	if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
		return dto.PublicUser{}, err
	}

	return toPublicUser(user), nil
}

func (s *userService) LockAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error) {
	if reason != "" {
		// TODO: integrate reason with audit logging
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.PublicUser{}, err
	}

	if user.Status == models.StatusDeleted {
		return dto.PublicUser{}, ErrUserDeleted
	}

	if user.Status != models.StatusLocked {
		user.Status = models.StatusLocked
		user.LockoutUntil = sql.NullTime{}

		if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
			return dto.PublicUser{}, err
		}
	}

	return toPublicUser(user), nil
}

func (s *userService) UnlockAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error) {
	if reason != "" {
		// TODO: integrate reason with audit logging
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.PublicUser{}, err
	}

	if user.Status == models.StatusDeleted {
		return dto.PublicUser{}, ErrUserDeleted
	}

	if user.Status != models.StatusActive {
		user.Status = models.StatusActive
		user.LockoutUntil = sql.NullTime{}

		if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
			return dto.PublicUser{}, err
		}
	}

	return toPublicUser(user), nil
}

func (s *userService) SoftDeleteAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error) {
	if reason != "" {
		// TODO: integrate reason with audit logging
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.PublicUser{}, err
	}

	if user.Status == models.StatusDeleted && user.DeletedAt.Valid {
		return toPublicUser(user), nil
	}

	user.Status = models.StatusDeleted
	user.DeletedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
		return dto.PublicUser{}, err
	}

	return toPublicUser(user), nil
}

func (s *userService) RestoreAccount(ctx context.Context, userID string, reason string) (dto.PublicUser, error) {
	if reason != "" {
		// TODO: integrate reason with audit logging
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.PublicUser{}, err
	}

	if !user.DeletedAt.Valid && user.Status != models.StatusDeleted {
		return toPublicUser(user), nil
	}

	user.Status = models.StatusActive
	user.DeletedAt = sql.NullTime{}
	user.LockoutUntil = sql.NullTime{}

	if err := s.userRepo.UpdateUser(ctx, &user); err != nil {
		return dto.PublicUser{}, err
	}

	return toPublicUser(user), nil
}
