package v2_12_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func createPachydermRestartSchema(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `create schema admin`); err != nil {
		return errors.Wrap(err, "create admin schema")
	}
	if _, err := tx.ExecContext(ctx, `create table admin.restarts (
		id bigserial not null primary key,
		restart_at timestamptz not null,
		reason text,
		created_by text not null,
		created_at timestamptz not null default now()
	)`); err != nil {
		return errors.Wrap(err, "create admin.restarts table")
	}
	return nil
}
