package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"user-services/internal/api/repositories"
	"user-services/internal/cache"
	"user-services/internal/config"
	"user-services/internal/errors"
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
		return AuthResult{}, errors.ErrEmailExists
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

	// Save verification token to user with configurable expiry
	cfg := config.GetConfig()
	user.EmailVerificationToken = tokenHash
	user.EmailVerificationExpiry = sql.NullTime{
		Time:  time.Now().Add(cfg.Email.VerificationExpiry),
		Valid: true,
	}
	if err := s.UserRepo.UpdateUser(ctx, &user); err != nil {
		return AuthResult{}, err
	}

	// 8. Build verification link
	cfg = config.GetConfig()
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", cfg.Email.FrontendURL, verificationToken)

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
	// Step 1: Authenticate user credentials
	user, err := s.authenticateUser(ctx, email, password, ipAddr)
	if err != nil {
		return AuthResult{}, err
	}

	// Step 2: Verify MFA if required
	if err := s.verifyMFA(ctx, user, mfaCode, email, ipAddr); err != nil {
		return AuthResult{}, err
	}

	// Step 3: Create session and tokens
	authResult, err := s.createSessionAndTokens(ctx, user, userAgent, ipAddr)
	if err != nil {
		_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "session_creation_failed")
		return AuthResult{}, err
	}

	// Step 4: Log successful login
	_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, true, "success")
	_ = s.UserRepo.UpdateLastLogin(ctx, user.ID, time.Now(), ipAddr)

	return authResult, nil
}

// authenticateUser validates user credentials
func (s *AuthService) authenticateUser(ctx context.Context, email, password, ipAddr string) (*models.User, error) {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		_ = s.logLoginAttempt(ctx, nil, email, ipAddr, false, "invalid_credentials")
		return nil, errors.ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return nil, errors.ErrEmailNotVerified
	}

	// Check account status
	switch user.Status {
	case "locked":
		return nil, errors.ErrAccountLocked
	case "disabled":
		return nil, errors.ErrAccountDisabled
	case "deleted":
		return nil, errors.ErrUserNotFound
	}

	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "invalid_credentials")
		return nil, errors.ErrInvalidCredentials
	}

	return &user, nil
}

// verifyMFA checks MFA if required for the user
func (s *AuthService) verifyMFA(ctx context.Context, user *models.User, mfaCode, email, ipAddr string) error {
	totp, err := s.MFARepo.GetTOTPByUserID(ctx, user.ID)
	if err != nil || totp == nil {
		// No MFA setup, skip verification
		return nil
	}

	if mfaCode == "" {
		_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "mfa_required")
		return errors.NewAuthenticationError("MFA code required").WithCode("MFA_REQUIRED")
	}

	if !utils.VerifyTOTP(totp.Secret, mfaCode, time.Now()) {
		_ = s.logLoginAttempt(ctx, &user.ID, email, ipAddr, false, "mfa_invalid")
		return errors.ErrInvalidMFACode
	}

	// Update last used timestamp
	_ = s.MFARepo.UpdateLastUsed(ctx, totp.ID)
	return nil
}

// createSessionAndTokens creates a session and generates JWT tokens
func (s *AuthService) createSessionAndTokens(ctx context.Context, user *models.User, userAgent, ipAddr string) (AuthResult, error) {
	cfg := config.GetConfig()

	// Create session in database
	sanitizedIP := utils.SanitizeIPAddress(ipAddr)
	var ipAddrPtr *string
	if sanitizedIP != "" {
		ipAddrPtr = &sanitizedIP
	}

	session := &models.Session{
		UserID:    user.ID,
		UserAgent: userAgent,
		IPAddr:    ipAddrPtr,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(cfg.Session.Expiry),
	}

	if err := s.SessionRepo.Create(ctx, session); err != nil {
		return AuthResult{}, err
	}

	// Store session in Redis
	if err := s.storeSessionInCache(ctx, session, user, userAgent, ipAddr); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to store session in Redis: %v\n", err)
	}

	// Generate JWT token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, session.ID)
	if err != nil {
		return AuthResult{}, err
	}

	// Generate refresh token
	refreshToken, err := s.createRefreshToken(ctx, session.ID)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User:         *user,
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(cfg.JWT.ExpiresIn),
	}, nil
}

// storeSessionInCache stores session data in Redis cache
func (s *AuthService) storeSessionInCache(ctx context.Context, session *models.Session, user *models.User, userAgent, ipAddr string) error {
	cfg := config.GetConfig()

	sessionData := cache.SessionData{
		UserID:    user.ID,
		Email:     user.Email,
		UserAgent: userAgent,
		IPAddr:    utils.SanitizeIPAddress(ipAddr),
		CreatedAt: session.CreatedAt,
	}

	return s.SessionCache.StoreSession(ctx, session.ID, sessionData, cfg.JWT.ExpiresIn)
}

// createRefreshToken creates and stores a refresh token
func (s *AuthService) createRefreshToken(ctx context.Context, sessionID uuid.UUID) (string, error) {
	cfg := config.GetConfig()

	rawRefresh := uuid.NewString()
	refreshHash, err := utils.HashPassword(rawRefresh)
	if err != nil {
		return "", err
	}

	refresh := &models.RefreshToken{
		SessionID: sessionID,
		TokenHash: refreshHash,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(cfg.JWT.RefreshExpiresIn),
	}

	if err := s.RefreshTokenRepo.Create(ctx, refresh); err != nil {
		return "", err
	}

	return rawRefresh, nil
}

// VerifyEmail verifies user's email with the provided token
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return errors.NewValidationError("Verification token is required").WithCode("TOKEN_REQUIRED")
	}

	// 1. Hash the token
	tokenHash := utils.HashToken(token)

	// 2. Find user with this token
	user, err := s.UserRepo.GetByVerificationToken(ctx, tokenHash)
	if err != nil {
		return errors.InvalidVerificationToken
	}

	// 3. Check if token is expired
	if user.EmailVerificationExpiry.Valid && user.EmailVerificationExpiry.Time.Before(time.Now()) {
		return errors.InvalidVerificationToken.WithDetails(map[string]interface{}{
			"reason": "expired",
		})
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
		return errors.NewInternalError("Failed to update user verification status").WithCause(err)
	}

	// 6. Log audit event
	s.logAuditEvent(ctx, &user.ID, "email.verified", map[string]any{
		"email": user.Email,
	})

	// 7. Send welcome email after verification
	s.sendWelcomeEmail(ctx, user)

	return nil
}

// logAuditEvent is a helper to log audit events
func (s *AuthService) logAuditEvent(ctx context.Context, userID *uuid.UUID, action string, metadata map[string]any) {
	auditLog := &models.AuditLog{
		UserID:    userID,
		Action:    action,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
	_ = s.AuditLogRepo.Create(ctx, auditLog)
}

// sendWelcomeEmail creates an outbox event for welcome email
func (s *AuthService) sendWelcomeEmail(ctx context.Context, user *models.User) {
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
}

// logLoginAttempt helper
func (s *AuthService) logLoginAttempt(ctx context.Context, userID *uuid.UUID, email, ip string, success bool, reason string) error {
	var ipAddr *string
	if ip != "" {
		ipAddr = &ip
	}
	attempt := &models.LoginAttempt{
		UserID:  userID,
		Email:   email,
		IPAddr:  ipAddr,
		Success: success,
		Reason:  reason,
	}
	return s.LoginAttemptRepo.Create(ctx, attempt)
}
