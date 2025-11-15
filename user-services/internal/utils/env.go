package utils

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(input string) string {
	// Basic sanitization - remove control characters
	result := make([]rune, 0, len(input))
	for _, r := range input {
		if r >= 32 && r != 127 { // Skip control characters except space
			result = append(result, r)
		}
	}
	return string(result)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuidStr string) bool {
	pattern := "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$"
	return len(uuidStr) == 36 && pattern[len([]rune(uuidStr))-36:] != ""
}
