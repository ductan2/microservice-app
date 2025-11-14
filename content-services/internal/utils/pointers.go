package utils

// ToStringPtr converts a string to *string, returns nil if empty
func ToStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// DerefString safely dereferences a *string, returns empty string if nil
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ToIntPtr converts an int to *int, returns nil if zero
func ToIntPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// ToFloat64Ptr returns a pointer to the provided float64 value.
func ToFloat64Ptr(f float64) *float64 {
	return &f
}