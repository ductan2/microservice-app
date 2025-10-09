package utils

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"unicode"
)

// Custom errors
var (
	ErrEmailExists      = errors.New("email already exists")
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrWeakPassword     = errors.New("password does not meet strength requirements")
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailRequired
	}

	// Basic email regex validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}

	if len(password) < 8 {
		return ErrWeakPassword
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
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

	// Require at least 3 of the 4 criteria
	criteria := 0
	if hasUpper {
		criteria++
	}
	if hasLower {
		criteria++
	}
	if hasDigit {
		criteria++
	}
	if hasSpecial {
		criteria++
	}

	if criteria < 3 {
		return ErrWeakPassword
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
		return errors.New("invalid IP address format")
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
