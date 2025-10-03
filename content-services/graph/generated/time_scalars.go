package generated

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

func (ec *executionContext) unmarshalInputTime(ctx context.Context, v any) (time.Time, error) {
	return graphql.UnmarshalTime(v)
}

func (ec *executionContext) _Time(ctx context.Context, sel ast.SelectionSet, v *time.Time) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	return graphql.MarshalTime(*v)
}
