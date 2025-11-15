package utils

import (
	"net"
	"regexp"
	"strings"
	"unicode"

	"user-services/internal/config"
	customerrors "user-services/internal/errors"
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return customerrors.NewValidationError("Email is required").WithCode("EMAIL_REQUIRED")
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return customerrors.NewValidationError("Email is required").WithCode("EMAIL_REQUIRED")
	}

	// Basic email regex validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return customerrors.ErrInvalidEmail
	}

	return nil
}

// ValidatePassword validates password strength based on security configuration
func ValidatePassword(password string) error {
	cfg := config.GetConfig()

	if password == "" {
		return customerrors.NewValidationError("Password is required").WithCode("PASSWORD_REQUIRED")
	}

	if len(password) < cfg.Security.PasswordMinLength {
		return customerrors.ErrWeakPassword.WithDetails(map[string]interface{}{
			"min_length": cfg.Security.PasswordMinLength,
		})
	}

	// Check for character requirements based on configuration
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Validate requirements based on configuration
	missingRequirements := make([]string, 0)

	if cfg.Security.PasswordRequireUpper && !hasUpper {
		missingRequirements = append(missingRequirements, "uppercase letter")
	}
	if cfg.Security.PasswordRequireLower && !hasLower {
		missingRequirements = append(missingRequirements, "lowercase letter")
	}
	if cfg.Security.PasswordRequireDigit && !hasDigit {
		missingRequirements = append(missingRequirements, "digit")
	}
	if cfg.Security.PasswordRequireSpecial && !hasSpecial {
		missingRequirements = append(missingRequirements, "special character")
	}

	// If any requirements are not met, return validation error
	if len(missingRequirements) > 0 {
		return customerrors.ErrWeakPassword.WithDetails(map[string]interface{}{
			"missing_requirements": missingRequirements,
			"requirements": map[string]interface{}{
				"uppercase":   cfg.Security.PasswordRequireUpper,
				"lowercase":   cfg.Security.PasswordRequireLower,
				"digit":       cfg.Security.PasswordRequireDigit,
				"special":     cfg.Security.PasswordRequireSpecial,
				"min_length":  cfg.Security.PasswordMinLength,
			},
		})
	}

	return nil
}

// ValidateIPAddress validates if a string is a valid IP address
func ValidateIPAddress(ip string) error {
	if ip == "" {
		return nil // Empty IP is allowed
	}

	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil // Empty after trim is allowed
	}

	if net.ParseIP(ip) == nil {
		return customerrors.NewValidationError("invalid IP address format")
	}

	return nil
}

// SanitizeIPAddress returns a valid IP address or empty string if invalid
func SanitizeIPAddress(ip string) string {
	if ip == "" {
		return ""
	}

	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}

	// Parse the IP to validate it
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "" // Return empty string for invalid IPs
	}

	return parsedIP.String()
}
