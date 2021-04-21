package collection

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

func CreatePostgresSchema(ctx context.Context, sqlTx *sqlx.Tx) error {
	_, err := sqlTx.ExecContext(ctx, `CREATE SCHEMA collections`)
	return err
}

func SetupPostgresV0(ctx context.Context, sqlTx *sqlx.Tx) error {
	updatedAtFunction := `
create or replace function collections.updatedat_trigger_fn() returns trigger as $$
begin
  new.updatedat = now();
	return new;
end;
$$ language 'plpgsql';
`
	if _, err := sqlTx.ExecContext(ctx, updatedAtFunction); err != nil {
		return errors.EnsureStack(err)
	}

	// This trigger function will publish events on various channels that may be
	// interested in changes to this object. There is a table-wide channel that
	// will receive notifications for every change to the table, as well as
	// deterministically named channels for listening to changes on a specific row
	// or index value.

	// For indexed notifications, we publish to a channel based on a hash of the
	// field and value being indexed on. There may be collisions where a channel
	// receives notifications for unrelated events (either from the same change or
	// different changes), so the payload includes enough information to filter
	// those events out.

	// The payload format is [field, value, key, operation, timestamp, protobuf],
	// although the payload must be a string so these are joined with space
	// delimiters and base64-encoded where appropriate.
	notifyFunction := `
create or replace function collections.notify_trigger_fn() returns trigger as $$
declare
  row record;
	encoded_key text;
	base_payload text;
	payload_end text;
	payload text;
	base_channel text;
	channel text;
	field text;
	value text;
	old_value text;
begin
  case tg_op
	when 'INSERT', 'UPDATE' then
	  row := new;
	when 'DELETE' then
	  row := old;
	else
	  raise exception 'Unrecognized tg_op: "%"', tg_op;
	end case;

	encoded_key := encode(row.key::bytea, 'base64');
	payload_end := date_part('epoch', now())::text || ' ' || encode(row.proto, 'base64');
	base_payload := encoded_key || ' ' || tg_op || ' ' || payload_end;
	base_channel := 'pwc_' || tg_table_name;

	if tg_argv is not null then
		foreach field in array tg_argv loop
			execute 'select ($1).' || field || '::text;' into value using row;
			if value is not null then
				payload := field || ' ' || encode(value::bytea, 'base64') || ' ' || base_payload;
				channel := base_channel || '_' || md5(field || ' ' || value);
				perform pg_notify(channel, payload);

				/* If an update changes this field value, we need to delete it from any watchers of the old value */
				if tg_op = 'UPDATE' then
					execute 'select ($1).' || field || '::text;' into old_value using old;
					if old_value is not null and old_value is distinct from value then
						payload := field || ' ' || encode(old_value::bytea, 'base64') || ' ' || encoded_key || ' DELETE ' || payload_end;
						channel := base_channel || '_' || md5(field || ' ' || old_value);
						perform pg_notify(channel, payload);
					end if;
				end if;
			end if;
		end loop;
	end if;

	payload := 'key ' || encoded_key || ' ' || base_payload;
	perform pg_notify(base_channel, payload);

	return row;
end;
$$ language 'plpgsql';
`
	if _, err := sqlTx.ExecContext(ctx, notifyFunction); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

func SetupPostgresCollections(ctx context.Context, sqlTx *sqlx.Tx, collections ...PostgresCollection) error {
	for _, pgc := range collections {
		col := pgc.(*postgresCollection)
		columns := []string{
			"createdat timestamp with time zone default current_timestamp",
			"updatedat timestamp with time zone default current_timestamp",
			"proto bytea",
			"version text",
			"key text primary key",
		}

		indexFields := []string{"'key'"}
		for _, idx := range col.indexes {
			name := indexFieldName(idx)
			columns = append(columns, name+" text")
			indexFields = append(indexFields, "'"+name+"'")
		}

		// TODO: create indexes on table
		createTable := fmt.Sprintf("create table if not exists collections.%s (%s);", col.table, strings.Join(columns, ", "))
		if _, err := sqlTx.Exec(createTable); err != nil {
			return errors.EnsureStack(err)
		}

		updatedatTrigger := fmt.Sprintf(`
	create trigger updatedat_trigger
		before insert or update or delete on collections.%s
		for each row execute procedure collections.updatedat_trigger_fn();
	`, col.table)
		if _, err := sqlTx.ExecContext(ctx, updatedatTrigger); err != nil {
			return errors.EnsureStack(err)
		}

		notifyTrigger := fmt.Sprintf(`
	create trigger notify_trigger
		after insert or update or delete on collections.%s
		for each row execute procedure collections.notify_trigger_fn(%s);
	`, col.table, strings.Join(indexFields, ", "))
		if _, err := sqlTx.ExecContext(ctx, notifyTrigger); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}
