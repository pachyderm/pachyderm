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
	if _, err := tx.ExecContext(ctx, `
create or replace function admin.notify_trigger_fn() returns trigger as $$
declare
begin
perform pg_notify(cast('restarts' as text), 'insert');
return new;
end;
$$ language 'plpgsql'`); err != nil {
		return errors.Wrap(err, "add stored procedure")
	}
	if _, err := tx.ExecContext(ctx, `create trigger notify_trigger `+
		`after insert on admin.restarts `+
		`for each row execute procedure admin.notify_trigger_fn()`); err != nil {
		return errors.Wrap(err, "add notify trigger to admin.restarts")
	}
	return nil
}
