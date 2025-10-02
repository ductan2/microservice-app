package dto

import "github.com/google/uuid"

// MFASetupRequest represents the request to setup MFA
type MFASetupRequest struct {
	Type  string `json:"type" binding:"required,oneof=totp webauthn"` // totp or webauthn
	Label string `json:"label"`                                        // optional label for the method
}

// MFASetupResponse represents the response after MFA setup initiation
type MFASetupResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Label     string    `json:"label,omitempty"`
	Secret    string    `json:"secret,omitempty"`     // TOTP secret (base32)
	QRCodeURL string    `json:"qr_code_url,omitempty"` // TOTP QR code data URL
	AddedAt   string    `json:"added_at,omitempty"`
}

// MFAVerifyRequest represents the request to verify MFA setup
type MFAVerifyRequest struct {
	MethodID uuid.UUID `json:"method_id" binding:"required"`
	Code     string    `json:"code" binding:"required,len=6"` // 6-digit TOTP code
}

// MFADisableRequest represents the request to disable MFA
type MFADisableRequest struct {
	MethodID uuid.UUID `json:"method_id" binding:"required"`
	Password string    `json:"password" binding:"required"` // require password for security
}

// MFALoginRequest represents MFA verification during login
type MFALoginRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}