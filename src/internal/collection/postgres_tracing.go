// This provides a wrapper implementing the sqlx.ExtContext interface that adds
// a span to any running trace if tracing is in progress

package collection

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
)

type tracingExtContext struct {
	inner sqlx.ExtContext
}

func (c tracingExtContext) DriverName() string {
	return c.inner.DriverName()
}

func (c tracingExtContext) Rebind(s string) string {
	return c.inner.Rebind(s)
}

func (c tracingExtContext) BindNamed(s string, i interface{}) (string, []interface{}, error) {
	return c.inner.BindNamed(s, i)
}

func (c tracingExtContext) QueryContext(ctx context.Context, query string, args ...interface{}) (_ *sql.Rows, retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/postgres/QueryContext", "query", query)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return c.inner.QueryContext(ctx, query, args...)
}

func (c tracingExtContext) QueryxContext(ctx context.Context, query string, args ...interface{}) (_ *sqlx.Rows, retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/postgres/QueryxContext", "query", query)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return c.inner.QueryxContext(ctx, query, args...)
}

func (c tracingExtContext) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/postgres/QueryRowxContext", "query", query)
	defer func() {
		tracing.FinishAnySpan(span)
	}()
	return c.inner.QueryRowxContext(ctx, query, args...)
}

func (c tracingExtContext) ExecContext(ctx context.Context, query string, args ...interface{}) (_ sql.Result, retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/postgres/ExecContext", "query", query)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return c.inner.ExecContext(ctx, query, args...)
}
