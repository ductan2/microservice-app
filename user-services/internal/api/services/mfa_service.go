package services

import (
	"context"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"

	"github.com/google/uuid"
)

type MFAService interface {
	SetupMFA(ctx context.Context, userID uuid.UUID, mfaType string) (*dto.MFASetupResponse, error)
	VerifyMFASetup(ctx context.Context, userID, methodID uuid.UUID, code string) error
	VerifyMFALogin(ctx context.Context, userID uuid.UUID, code string) error
	DisableMFA(ctx context.Context, userID, methodID uuid.UUID, password string) error
	GetUserMFAMethods(ctx context.Context, userID uuid.UUID) ([]dto.MFASetupResponse, error)
}

type mfaService struct {
	mfaRepo  repositories.MFARepository
	userRepo repositories.UserRepository
}

func NewMFAService(
	mfaRepo repositories.MFARepository,
	userRepo repositories.UserRepository,
) MFAService {
	return &mfaService{
		mfaRepo:  mfaRepo,
		userRepo: userRepo,
	}
}

func (s *mfaService) SetupMFA(ctx context.Context, userID uuid.UUID, mfaType string) (*dto.MFASetupResponse, error) {
	// TODO: implement TOTP secret generation or WebAuthn challenge
	return nil, nil
}

func (s *mfaService) VerifyMFASetup(ctx context.Context, userID, methodID uuid.UUID, code string) error {
	// TODO: implement verification
	return nil
}

func (s *mfaService) VerifyMFALogin(ctx context.Context, userID uuid.UUID, code string) error {
	// TODO: implement login MFA verification
	return nil
}

func (s *mfaService) DisableMFA(ctx context.Context, userID, methodID uuid.UUID, password string) error {
	// TODO: implement
	return nil
}

func (s *mfaService) GetUserMFAMethods(ctx context.Context, userID uuid.UUID) ([]dto.MFASetupResponse, error) {
	// TODO: implement
	return nil, nil
}
