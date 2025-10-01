package services

import (
	"context"
	"errors"
	"time"

	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrProfileNotFound = errors.New("profile not found")

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
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	return &dto.UserProfile{
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Locale:      profile.Locale,
		TimeZone:    profile.TimeZone,
	}, nil
}

func (s *userProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) error {
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProfileNotFound
		}
		return err
	}

	if req.DisplayName != "" {
		profile.DisplayName = req.DisplayName
	}
	if req.AvatarURL != "" {
		profile.AvatarURL = req.AvatarURL
	}
	if req.Locale != "" {
		profile.Locale = req.Locale
	}
	if req.TimeZone != "" {
		profile.TimeZone = req.TimeZone
	}

	profile.UpdatedAt = time.Now()

	return s.profileRepo.Update(ctx, profile)
}

func (s *userProfileService) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	// TODO: implement
	return nil
}
