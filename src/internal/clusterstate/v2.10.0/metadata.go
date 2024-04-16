package v2_10_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func addMetadataToProjects(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE core.projects ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object')`); err != nil {
		return errors.Wrap(err, "add metadata column to core.projects")
	}
	return nil
}

func addMetadataToCommits(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE pfs.commits ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object')`); err != nil {
		return errors.Wrap(err, "add metadata column to pfs.commits")
	}
	return nil
}

func addMetadataToBranches(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE pfs.branches ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object')`); err != nil {
		return errors.Wrap(err, "add metadata column to pfs.branches")
	}
	return nil
}

func addMetadataToRepos(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE pfs.repos ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object')`); err != nil {
		return errors.Wrap(err, "add metadata column to pfs.repos")
	}
	return nil
}

func addClusterMetadata(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	// This adds a cluster metadata table that can only have one row.  The PRIMARY KEY NOT NULL
	// constraint on onlyonerow means that there can only be two rows, one true and one false.
	// The CHECK constraint ensures that there can only be a row with the primary key true.
	// This limits the table to one row; since Pachyderm only addresses one cluster, this is the
	// right number of rows for storing cluster metadata.  Don't delete this row unless you're
	// deleting cluster metadata entirely, however, as the rest of the code assumes it exists.
	if _, err := tx.ExecContext(ctx, `CREATE TABLE core.cluster_metadata (
		onlyonerow BOOLEAN PRIMARY KEY NOT NULL DEFAULT true,
		metadata JSONB NOT NULL DEFAULT '{}',
		CONSTRAINT onlyonerow CHECK (onlyonerow=true)
	);
	INSERT INTO core.cluster_metadata (onlyonerow, metadata) VALUES (true, '{}')`); err != nil {
		return errors.Wrap(err, "add core.cluster_metadata table and first row")
	}
	return nil
}
