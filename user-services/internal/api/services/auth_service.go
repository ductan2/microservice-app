package services

import (
	"context"
	"errors"
	"time"

	"user-services/internal/api/repositories"
	"user-services/internal/config"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	UserRepo         *repositories.UserRepository
	UserProfileRepo  repositories.UserProfileRepository
	AuditLogRepo     repositories.AuditLogRepository
	OutboxRepo       repositories.OutboxRepository
	SessionRepo      repositories.SessionRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	MFARepo          repositories.MFARepository
	LoginAttemptRepo repositories.LoginAttemptRepository
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	userRepo *repositories.UserRepository,
	userProfileRepo repositories.UserProfileRepository,
	auditLogRepo repositories.AuditLogRepository,
	outboxRepo repositories.OutboxRepository,
	sessionRepo repositories.SessionRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	mfaRepo repositories.MFARepository,
	loginAttemptRepo repositories.LoginAttemptRepository,
) *AuthService {
	return &AuthService{
		UserRepo:         userRepo,
		UserProfileRepo:  userProfileRepo,
		AuditLogRepo:     auditLogRepo,
		OutboxRepo:       outboxRepo,
		SessionRepo:      sessionRepo,
		RefreshTokenRepo: refreshTokenRepo,
		MFARepo:          mfaRepo,
		LoginAttemptRepo: loginAttemptRepo,
	}
}

// AuthResult contains user data and JWT token
type AuthResult struct {
	User         models.User
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}

// Register creates a new user account and returns auth result
func (s *AuthService) Register(ctx context.Context, email, password, name string) (AuthResult, error) {
	// 1. Validate input
	if err := utils.ValidateEmail(email); err != nil {
		return AuthResult{}, err
	}

	if err := utils.ValidatePassword(password); err != nil {
		return AuthResult{}, err
	}

	// 2. Check if email already exists
	exists, err := s.UserRepo.CheckEmailExists(ctx, email)
	if err != nil {
		return AuthResult{}, err
	}
	if exists {
		return AuthResult{}, utils.ErrEmailExists
	}

	// 3. Hash password with bcrypt
	hash, err := utils.HashPassword(password)
	if err != nil {
		return AuthResult{}, err
	}

	// 4. Create user in database
	user, err := s.UserRepo.CreateUser(ctx, email, hash)
	if err != nil {
		return AuthResult{}, err
	}

	// 5. Create user profile
	profile := &models.UserProfile{
		UserID:      user.ID,
		DisplayName: name,
		Locale:      "en",
		TimeZone:    "UTC",
		UpdatedAt:   time.Now(),
	}

	if err := s.UserProfileRepo.Create(ctx, profile); err != nil {
		// If profile creation fails, we should clean up the user
		// For now, we'll just return the error
		return AuthResult{}, err
	}

	// 6. Log audit event
	auditLog := &models.AuditLog{
		UserID: &user.ID,
		Action: "user.registered",
		Metadata: map[string]any{
			"email": user.Email,
			"name":  name,
		},
		CreatedAt: time.Now(),
	}

	if err := s.AuditLogRepo.Create(ctx, auditLog); err != nil {
		// Log error but don't fail registration
		// In production, you might want to use a proper logger
	}

	// 7. Create outbox event for user.created
	outboxEvent := &models.Outbox{
		AggregateID: user.ID,
		Topic:       "user.created",
		Type:        "UserCreated",
		Payload: map[string]any{
			"user_id": user.ID,
			"email":   user.Email,
			"name":    name,
		},
		CreatedAt: time.Now(),
	}

	if err := s.OutboxRepo.Create(ctx, outboxEvent); err != nil {
		// Log error but don't fail registration
		// In production, you might want to use a proper logger
	}

	// 8. Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return AuthResult{}, err
	}

	// For registration, we don't create session/refresh token yet
	// User will need to login to get full session
	return AuthResult{
		User:      user,
		Token:     token,
		ExpiresAt: time.Now().Add(config.GetJWTConfig().ExpiresIn),
	}, nil
}

// Login authenticates a user and returns auth result
func (s *AuthService) Login(ctx context.Context, email, password, mfaCode, userAgent, ipAddr string) (AuthResult, error) {
	// 1) Find user by email
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Log failed attempt (no user id)
		_ = s.logLoginAttempt(ctx, nil, email, ipAddr, false, "invalid_credentials")
		return AuthResult{}, errors.New("invalid email or password")
	}
	if user.EmailVerified == false {
		return AuthResult{}, errors.New("email not verified")
	}

	// 2) Verify password
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "invalid_credentials")
		return AuthResult{}, errors.New("invalid email or password")
	}

	// 3) If MFA enabled, verify code
	// For now, check if TOTP exists; if exists require mfaCode and verify
	if totp, err := s.MFARepo.GetTOTPByUserID(ctx, user.ID); err == nil && totp != nil {
		if mfaCode == "" || !utils.VerifyTOTP(totp.Secret, mfaCode, time.Now()) {
			_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "mfa_required_or_invalid")
			return AuthResult{}, utils.ErrInvalidMFACode
		}
		_ = s.MFARepo.UpdateLastUsed(ctx, totp.ID)
	}

	// 4) Create session
	session := &models.Session{
		UserID:    user.ID,
		UserAgent: userAgent,
		IPAddr:    ipAddr,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // default 30d session expiry
	}
	if err := s.SessionRepo.Create(ctx, session); err != nil {
		return AuthResult{}, err
	}

	// 5) Issue tokens: access JWT and refresh token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return AuthResult{}, err
	}

	// Create random refresh token and store hashed
	rawRefresh := uuid.NewString()
	refreshHash, err := utils.HashPassword(rawRefresh)
	if err != nil {
		return AuthResult{}, err
	}
	refresh := &models.RefreshToken{
		SessionID: session.ID,
		TokenHash: refreshHash,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), // 60d refresh
	}
	if err := s.RefreshTokenRepo.Create(ctx, refresh); err != nil {
		return AuthResult{}, err
	}

	// 6) Audit + outbox could be added here if needed

	// 7) Log success attempt
	_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, true, "success")

	// Build response
	return AuthResult{
		User:         user,
		Token:        accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    time.Now().Add(config.GetJWTConfig().ExpiresIn),
	}, nil
}

// logLoginAttempt helper
func (s *AuthService) logLoginAttempt(ctx context.Context, userID *uuid.UUID, email, ip string, success bool, reason string) error {
	attempt := &models.LoginAttempt{
		UserID:  userID,
		Email:   email,
		IPAddr:  ip,
		Success: success,
		Reason:  reason,
	}
	return s.LoginAttemptRepo.Create(ctx, attempt)
}
