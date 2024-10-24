package v2_12_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func createSnapshotSchema(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `create schema recovery`); err != nil {
		return errors.Wrap(err, "create recovery schema")
	}
	if _, err := tx.ExecContext(ctx, `create table recovery.snapshots (
		id bigserial not null primary key,
		chunkset bigint not null,
		sql_dump_pin bigint null,
		metadata jsonb not null default '{}',
		pachyderm_version text not null,
		created_at timestamptz not null default now()
	)`); err != nil {
		return errors.Wrap(err, "create recovery.snapshots table")
	}
	return nil
}
