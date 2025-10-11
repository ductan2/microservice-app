package utils

import (
	"bff-services/internal/types"
	"bytes"
)

func IsBodyEmpty(r *types.HTTPResponse) bool {
	if r == nil {
		return true
	}
	return len(bytes.TrimSpace(r.Body)) == 0
}
