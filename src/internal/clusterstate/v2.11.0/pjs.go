package v2_11_0

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func createPJSSchema(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "normalizeCommitDiffs")
	tx := env.Tx
	_, err := tx.ExecContext(ctx, `
		CREATE SCHEMA pjs; 
		CREATE TYPE pjs.job_error_code AS ENUM (
			'failed', 
			'disconnected',
			'canceled'
		);
		CREATE TABLE pjs.jobs (
				id BIGSERIAL PRIMARY KEY,
				
				-- These are all fileset IDs, referential integrity can be added
				-- later once stable fileset hashing is designed and implemented.
				input BYTEA,
				spec BYTEA NOT NULL,
				spec_hash BYTEA NOT NULL, -- the hash will be used to model the queues
				output BYTEA,
				
				parent BIGINT REFERENCES pjs.jobs(id),
				error pjs.job_error_code,
				
				-- error is mutually exclusive with output.
				CONSTRAINT jobs_output_or_error CHECK (num_nonnulls(error, output) <= 1),
				
				queued timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
				processing timestamptz,
				done timestamptz		
		);
	`)
	if err != nil {
		return errors.Wrap(err, "create PJS schema")
	}
	return nil
}
