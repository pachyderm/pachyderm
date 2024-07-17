package v2_11_0

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func createPJSSchema(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "createPJSSchema")
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
				parent BIGINT REFERENCES pjs.jobs(id),
				
				program BYTEA NOT NULL,
				program_hash BYTEA NOT NULL, -- the hash will be used to model the queues
				
				error pjs.job_error_code,
				
				queued timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
				processing timestamptz,
				done timestamptz		
		);
		CREATE TYPE pjs.fileset_types AS ENUM (
			'input', 
			'output'
		);
		-- A job's input and output may be more than one fileset.
		-- An extra table makes looking up, inserting, and reordering those filesets more efficient.
		CREATE TABLE pjs.job_filesets (
			job_id BIGINT REFERENCES pjs.jobs(id) ON DELETE CASCADE,
			fileset_type pjs.fileset_types NOT NULL,
			array_position INT NOT NULL,
			fileset BYTEA NOT NULL,
			PRIMARY KEY(job_id, fileset_type, array_position)
		);
		-- a reverse index here should make it faster to look up whether a fileset is used by a job.
		CREATE INDEX ON pjs.job_filesets (
		    fileset, job_id
		);
	`)
	if err != nil {
		return errors.Wrap(err, "create PJS schema")
	}
	return nil
}
