package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"user-services/internal/api/repositories"
	"user-services/internal/cache"
	"user-services/internal/config"
	"user-services/internal/models"
	"user-services/internal/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	UserRepo         repositories.UserRepository
	UserProfileRepo  repositories.UserProfileRepository
	AuditLogRepo     repositories.AuditLogRepository
	OutboxRepo       repositories.OutboxRepository
	SessionRepo      repositories.SessionRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	MFARepo          repositories.MFARepository
	LoginAttemptRepo repositories.LoginAttemptRepository
	SessionCache     *cache.SessionCache
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	userRepo repositories.UserRepository,
	userProfileRepo repositories.UserProfileRepository,
	auditLogRepo repositories.AuditLogRepository,
	outboxRepo repositories.OutboxRepository,
	sessionRepo repositories.SessionRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	mfaRepo repositories.MFARepository,
	loginAttemptRepo repositories.LoginAttemptRepository,
	sessionCache *cache.SessionCache,
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
		SessionCache:     sessionCache,
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

	// 7. Generate email verification token
	verificationToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return AuthResult{}, err
	}

	// Hash the token for storage
	tokenHash := utils.HashToken(verificationToken)

	// Save verification token to user
	user.EmailVerificationToken = tokenHash
	user.EmailVerificationExpiry = sql.NullTime{
		Time:  time.Now().Add(24 * time.Hour), // Expires in 24 hours
		Valid: true,
	}
	if err := s.UserRepo.UpdateUser(ctx, &user); err != nil {
		return AuthResult{}, err
	}

	// 8. Build verification link
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, verificationToken)

	// 9. Create outbox event for email verification
	payloadData := map[string]any{
		"user_id":           user.ID,
		"email":             user.Email,
		"name":              name,
		"verification_link": verificationLink,
		"verificationLink":  verificationLink, // alternative key
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		// Log error but don't fail registration
		// In production, you might want to use a proper logger
	} else {
		outboxEvent := &models.Outbox{
			AggregateID: user.ID,
			Topic:       "user.email_verification",
			Type:        "EmailVerificationRequested",
			Payload:     payloadBytes,
			CreatedAt:   time.Now(),
		}

		if err := s.OutboxRepo.Create(ctx, outboxEvent); err != nil {
			// Log error but don't fail registration
			// In production, you might want to use a proper logger
		}
	}

	// 10. Return result WITHOUT token (user needs to verify email first)
	return AuthResult{
		User: user,
		// No token until email is verified
	}, nil
}

// Login authenticates a user and returns auth result
func (s *AuthService) Login(ctx context.Context, email, password, mfaCode, userAgent, ipAddr string) (AuthResult, error) {
	// 1) Find user by email
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Log failed attempt (no user id)
		_ = s.logLoginAttempt(ctx, nil, email, ipAddr, false, "invalid_credentials")
		return AuthResult{}, utils.ErrInvalidCredentials
	}
	if !user.EmailVerified {
		return AuthResult{}, utils.ErrEmailNotVerified
	}

	// 2) Verify password
	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "invalid_credentials")
		return AuthResult{}, utils.ErrInvalidCredentials
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

	// 4.1) Store session in Redis with TTL matching JWT expiration
	jwtConfig := config.GetJWTConfig()
	sessionData := cache.SessionData{
		UserID:    user.ID,
		Email:     user.Email,
		UserAgent: userAgent,
		IPAddr:    ipAddr,
		CreatedAt: session.CreatedAt,
	}
	if err := s.SessionCache.StoreSession(ctx, session.ID, sessionData, jwtConfig.ExpiresIn); err != nil {
		// Log error but don't fail login - session still exists in DB
		// In production, you might want to use a proper logger
		fmt.Printf("Warning: failed to store session in Redis: %v\n", err)
	}

	// 5) Issue tokens: access JWT and refresh token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, session.ID)
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

// VerifyEmail verifies user's email with the provided token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// 1. Hash the token
	tokenHash := utils.HashToken(token)

	// 2. Find user with this token
	user, err := s.UserRepo.GetByVerificationToken(ctx, tokenHash)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	// 3. Check if token is expired
	if user.EmailVerificationExpiry.Valid && user.EmailVerificationExpiry.Time.Before(time.Now()) {
		return errors.New("verification token has expired")
	}

	// 4. Check if already verified
	if user.EmailVerified {
		return nil // Already verified, silently succeed
	}

	// 5. Mark email as verified and clear token
	user.EmailVerified = true
	user.EmailVerificationToken = ""
	user.EmailVerificationExpiry = sql.NullTime{Valid: false}

	if err := s.UserRepo.UpdateUser(ctx, user); err != nil {
		return err
	}

	// 6. Log audit event
	auditLog := &models.AuditLog{
		UserID: &user.ID,
		Action: "email.verified",
		Metadata: map[string]any{
			"email": user.Email,
		},
		CreatedAt: time.Now(),
	}
	_ = s.AuditLogRepo.Create(ctx, auditLog)

	// 7. Send welcome email after verification
	payloadData := map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
		"name":    user.Profile.DisplayName,
	}
	payloadBytes, _ := json.Marshal(payloadData)
	outboxEvent := &models.Outbox{
		AggregateID: user.ID,
		Topic:       "user.created",
		Type:        "UserCreated",
		Payload:     payloadBytes,
		CreatedAt:   time.Now(),
	}
	_ = s.OutboxRepo.Create(ctx, outboxEvent)

	return nil
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
