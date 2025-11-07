package utils

import "github.com/google/uuid"

func ErrString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func NormalizeUUIDOrString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case uuid.UUID:
		return t.String()
	default:
		return ""
	}
}

func NormalizeString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
