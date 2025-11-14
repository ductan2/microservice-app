package utils

import (
	"context"
	"errors"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// UserIDFromContext extracts user ID from GraphQL context, returns error if not authenticated
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	id, ok, err := UserIDFromContextOptional(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	if !ok {
		return uuid.Nil, gqlerror.Errorf("authentication required")
	}
	return id, nil
}

// UserIDFromContextOptional extracts user ID from GraphQL context, returns (nil, false, nil) if not present
func UserIDFromContextOptional(ctx context.Context) (uuid.UUID, bool, error) {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return uuid.Nil, false, gqlerror.Errorf("missing request context")
	}

	userID := strings.TrimSpace(opCtx.Headers.Get("X-User-ID"))
	if userID == "" {
		return uuid.Nil, false, nil
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, false, gqlerror.Errorf("invalid user id")
	}

	return id, true, nil
}

// ValidateUUID validates and parses a UUID string
func ValidateUUID(idStr string) (uuid.UUID, error) {
	if strings.TrimSpace(idStr) == "" {
		return uuid.Nil, errors.New("empty UUID")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid UUID format")
	}

	return id, nil
}