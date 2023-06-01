package v2_7_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func SetupCore(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS core;`); err != nil {
		return errors.Wrap(err, "error creating core schema")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION core.set_created_at_to_now() RETURNS TRIGGER AS $$
		BEGIN
			NEW.created_at = now();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`); err != nil {
		return errors.Wrap(err, "error creating set_created_at_to_now trigger function")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION core.set_updated_at_to_now() RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = now();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`); err != nil {
		return errors.Wrap(err, "error creating set_updated_at_to_now trigger function")
	}

	if err := SetupProjectsTable(ctx, tx); err != nil {
		return errors.Wrap(err, "error creating projects table")
	}
	return nil
}

func SetupProjectsTable(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS core.projects (
			id bigserial PRIMARY KEY,
			name text UNIQUE NOT NULL,
			description text,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL
		);
	`); err != nil {
		return errors.Wrap(err, "error creating projects table")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_created_at
			BEFORE INSERT ON core.projects
			FOR EACH ROW EXECUTE PROCEDURE core.set_created_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "error creating set_created_at trigger")
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TRIGGER set_updated_at
			BEFORE INSERT OR UPDATE ON core.projects
			FOR EACH ROW EXECUTE PROCEDURE core.set_updated_at_to_now();
	`); err != nil {
		return errors.Wrap(err, "error creating set_updated_at trigger")
	}
	return nil
}
