package utils

import (
	"strings"
)

// CloneBody creates a deep copy of a body map[string]any
func CloneBody(body map[string]any) map[string]any {
	if body == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(body))
	for k, v := range body {
		cloned[k] = v
	}

	return cloned
}

// TrimAndValidateString trims whitespace from a string pointer and validates it's not empty
func TrimAndValidateString(s *string) *string {
	if s == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

// ValidateSearchString validates and trims a search string
func ValidateSearchString(search *string) string {
	if search == nil {
		return ""
	}
	return strings.TrimSpace(*search)
}