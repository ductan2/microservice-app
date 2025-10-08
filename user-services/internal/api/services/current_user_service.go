package services

import (
	"context"

	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
)

type CurrentUserService interface {
	GetPublicUserByID(ctx context.Context, id string) (dto.PublicUser, error)
}

type currentUserService struct {
	userRepo repositories.UserRepository
}

func NewCurrentUserService(userRepo repositories.UserRepository) CurrentUserService {
	return &currentUserService{userRepo: userRepo}
}

func (s *currentUserService) GetPublicUserByID(ctx context.Context, id string) (dto.PublicUser, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return dto.PublicUser{}, err
	}
	return toPublicUser(user), nil
}
