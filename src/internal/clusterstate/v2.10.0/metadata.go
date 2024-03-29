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
