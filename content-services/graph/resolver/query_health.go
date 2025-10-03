package resolver

import "context"

// Health is the resolver for the health field.
func (r *queryResolver) Health(ctx context.Context) (string, error) {
	return "ok", nil
}
