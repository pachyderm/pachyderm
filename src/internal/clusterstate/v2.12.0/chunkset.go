package v2_12_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func createChunksetSchema(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "createChunksetSchema")
	tx := env.Tx
	_, err := tx.ExecContext(ctx, `
		create table storage.chunksets (
		   id bigserial not null primary key,
		   created_at timestamptz not null default now()
		);
	`)
	if err != nil {
		return errors.Wrap(err, "create Chunkset schema")
	}
	return nil
}
