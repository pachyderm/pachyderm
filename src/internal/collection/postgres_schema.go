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

	// The payload format is [key, timestamp, operation, field, value, location, protobuf],
	// although the payload must be a string so these are joined with space
	// delimiters and base64-encoded where appropriate.
	notifyFunction := `
create or replace function collections.notify_trigger_fn() returns trigger as $$
declare
  row record;
	encoded_key text;
	payload text;
	payload_prefix text;
	payload_suffix text;
	base_channel text;
	field text;
	value text;
	old_value text;
	notify_channels text[];
	notify_payloads text[];
	max_len int;
	notification_id text;
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
	/* If the payload is too long, we have to write to a table to be read out separately by the client */
	payload_prefix := encoded_key || ' ' || date_part('epoch', now())::text || ' ';
	base_channel := 'pwc_' || tg_table_name;

	notify_channels := array[base_channel];
	notify_payloads := array[payload_prefix || tg_op || ' key ' || encoded_key || ' '];
	max_len := length(notify_payloads[1]);

	if tg_argv is not null then
		foreach field in array tg_argv loop
			execute 'select ($1).' || field || '::text;' into value using row;
			if value is not null then
				payload := payload_prefix || tg_op || ' ' || field || ' ' || encode(value::bytea, 'base64') || ' ';
				notify_payloads = notify_payloads || payload;
				notify_channels = notify_channels || (base_channel || '_' || md5(field || ' ' || value));
				if length(payload) > max_len then
					max_len = length(payload);
				end if;
			end if;

			/* If an update changes this field value, we need to delete it from any watchers of the old value */
			if tg_op = 'UPDATE' then
				execute 'select ($1).' || field || '::text;' into old_value using old;
				if old_value is not null and old_value is distinct from value then
					payload := payload_prefix || 'DELETE ' || field || ' ' || encode(old_value::bytea, 'base64') || ' ';
					notify_payloads = notify_payloads || payload;
					notify_channels = notify_channels || (base_channel || '_' || md5(field || ' ' || old_value));
					if length(payload) > max_len then
						max_len = length(payload);
					end if;
				end if;
			end if;
		end loop;
	end if;

	if max_len + 4 * ceil(length(row.proto)::float / 3) >= 7990 then
	  /* write payload to a table to be read out separately by the client */
		insert into collections.large_notifications ("proto") values (row.proto) returning id into notification_id;
		payload_suffix := 'stored' || ' ' || notification_id;
	else
		payload_suffix := 'inline' || ' ' || encode(row.proto, 'base64');
	end if;

	for i in 1..array_length(notify_channels, 1) loop
		perform pg_notify(notify_channels[i], notify_payloads[i] || payload_suffix);
	end loop;

	return row;
end;
$$ language 'plpgsql';
`
	if _, err := sqlTx.ExecContext(ctx, notifyFunction); err != nil {
		return errors.EnsureStack(err)
	}

	notificationTable := `
create table if not exists collections.large_notifications (
	id serial,
	createdat timestamp with time zone default current_timestamp,
	proto bytea
)
`

	if _, err := sqlTx.ExecContext(ctx, notificationTable); err != nil {
		return errors.EnsureStack(err)
	}

	notificationsIndex := "create index on collections.large_notifications (createdat);"
	if _, err := sqlTx.Exec(notificationsIndex); err != nil {
		return errors.EnsureStack(err)
	}

	// The sweep function makes sure that the size of the large_notifications
	// table is bounded by deleting notifications over a certain age when a new
	// notification is written.
	notificationsSweepFunction := `
create or replace function collections.notifications_sweep_trigger_fn() returns trigger as $$
begin
  delete from collections.large_notifications where createdat < now() - interval '1 hour';
  return new;
end;
$$ language 'plpgsql';
`
	if _, err := sqlTx.ExecContext(ctx, notificationsSweepFunction); err != nil {
		return errors.EnsureStack(err)
	}

	notificationsSweepTrigger := `
create trigger updatedat_trigger
	after insert on collections.large_notifications
	for each row execute procedure collections.notifications_sweep_trigger_fn();
`
	if _, err := sqlTx.ExecContext(ctx, notificationsSweepTrigger); err != nil {
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

		for _, idx := range col.indexes {
			createIndex := fmt.Sprintf("create index on collections.%s (%s);", col.table, indexFieldName(idx))
			if _, err := sqlTx.Exec(createIndex); err != nil {
				return errors.EnsureStack(err)
			}
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
