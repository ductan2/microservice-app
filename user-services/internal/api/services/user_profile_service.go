package services

import (
	"context"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type UserProfileService interface {
	CreateProfile(ctx context.Context, userID uuid.UUID) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) error
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
}

type userProfileService struct {
	profileRepo repositories.UserProfileRepository
}

func NewUserProfileService(profileRepo repositories.UserProfileRepository) UserProfileService {
	return &userProfileService{
		profileRepo: profileRepo,
	}
}

func (s *userProfileService) CreateProfile(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}

func (s *userProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserProfile, error) {
	// TODO: implement
	return nil, nil
}

func (s *userProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) error {
	// TODO: implement
	return nil
}

func (s *userProfileService) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}
