package services

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"user-services/internal/api/dto"
	"user-services/internal/api/repositories"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type MFAService interface {
	SetupMFA(ctx context.Context, userID uuid.UUID, mfaType string, label string) (*dto.MFASetupResponse, error)
	VerifyMFASetup(ctx context.Context, userID, methodID uuid.UUID, code string) error
	VerifyMFALogin(ctx context.Context, userID uuid.UUID, code, secret string) error
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

func (s *mfaService) SetupMFA(ctx context.Context, userID uuid.UUID, mfaType string, label string) (*dto.MFASetupResponse, error) {
	if mfaType == "totp" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "MyApp",
			AccountName: label,
			Period:      30,
			SecretSize:  20,
			Algorithm:   otp.AlgorithmSHA1,
		})
		if err != nil {
			return nil, err
		}

		m := &models.MFAMethod{
			ID:     uuid.New(),
			UserID: userID,
			Type:   "totp",
			Label:  label,
			Secret: key.Secret(),
			LastUsedAt: sql.NullTime{},
		}

		if err := s.mfaRepo.Create(ctx, m); err != nil {
			return nil, err
		}

		return &dto.MFASetupResponse{
			ID:        m.ID,
			Type:      m.Type,
			Label:     m.Label,
			Secret:    key.Secret(),
			QRCodeURL: key.URL(),
		}, nil
	}
	return nil, errors.New("unsupported mfa type")
}

func (s *mfaService) VerifyMFASetup(ctx context.Context, userID, methodID uuid.UUID, code string) error {
	m, err := s.mfaRepo.GetByID(ctx, methodID)
	if err != nil || m.UserID != userID {
		return errors.New("method not found")
	}
	if m.Type == "totp" {
		if !totp.Validate(code, m.Secret) {
			return errors.New("invalid code")
		}
		m.LastUsedAt = sql.NullTime{Time: time.Now(), Valid: true}
		if err := s.mfaRepo.UpdateLastUsed(ctx, m.ID); err != nil {
			return err
		}
		return nil
	}
	return errors.New("unsupported type")
}

func (s *mfaService) VerifyMFALogin(ctx context.Context, userID uuid.UUID, code, secret string) error {
	if !totp.Validate(code, secret) {
		return errors.New("invalid MFA code")
	}
	return nil
}

func (s *mfaService) DisableMFA(ctx context.Context, userID, methodID uuid.UUID, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return errors.New("invalid password")
	}
	return s.mfaRepo.Delete(ctx, methodID)
}

// Get List MFA of a user
func (s *mfaService) GetUserMFAMethods(ctx context.Context, userID uuid.UUID) ([]dto.MFASetupResponse, error) {
	methods, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.MFASetupResponse, 0)
	for _, m := range methods {
		resp = append(resp, dto.MFASetupResponse{
			ID:     m.ID,
			Type:   m.Type,
			Label:  m.Label,
			AddedAt: m.AddedAt.Format(time.RFC3339),
		})
	}
	return resp, nil
}