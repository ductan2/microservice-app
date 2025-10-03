package dto

// MFASetupRequest represents payload to setup MFA.
type MFASetupRequest struct {
	Type  string `json:"type" binding:"required"`
	Label string `json:"label" binding:"required"`
}

// MFAVerifyRequest represents payload to verify an MFA method.
type MFAVerifyRequest struct {
	MethodID string `json:"method_id" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// MFADisableRequest represents payload to disable an MFA method.
type MFADisableRequest struct {
	MethodID string `json:"method_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}
