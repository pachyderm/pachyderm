package storagedb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// SetupPostgresTrackerV0 sets up the table for the postgres tracker
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SchemaTrackerV0(ctx context.Context, tx *pachsql.Tx) error {
	const schema = `
	CREATE TABLE storage.tracker_objects (
		int_id BIGSERIAL PRIMARY KEY,
		str_id VARCHAR(4096) UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP
	);

	CREATE TABLE storage.tracker_refs (
		from_id INT8 NOT NULL,
		to_id INT8 NOT NULL,
		PRIMARY KEY (from_id, to_id)
	);

	CREATE INDEX ON storage.tracker_refs (
		to_id,
		from_id
	);
`
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
}

// SchemaChunkV0 sets up tables in db
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SchemaChunkV0(tx *pachsql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE storage.chunk_objects (
		chunk_id BYTEA NOT NULL,
		gen BIGSERIAL NOT NULL,
		uploaded BOOLEAN NOT NULL DEFAULT FALSE,
		tombstone BOOLEAN NOT NULL DEFAULT FALSE,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY(chunk_id, gen)
	);

	CREATE TABLE storage.keys (
		name VARCHAR(128) NOT NULL,
		data BYTEA NOT NULL,

		PRIMARY KEY(name)
	)
	`)
	return errors.EnsureStack(err)
}

// SetupPostgresStoreV0 sets up the tables for a Store
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SchemaFilesetV0(ctx context.Context, tx *pachsql.Tx) error {
	const schema = `
	CREATE TABLE storage.filesets (
		id UUID NOT NULL PRIMARY KEY,
		metadata_pb BYTEA NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
`
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
}

// SchemaFilesetCacheV1 creates the table for a cache.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SchemaFilesetCacheV1(ctx context.Context, tx *pachsql.Tx) error {
	const schema = `
	CREATE TABLE storage.cache (
		key text NOT NULL PRIMARY KEY,
		value_pb BYTEA NOT NULL,
		ids UUID[] NOT NULL,
		accessed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		tag text
	);
	CREATE INDEX ON storage.cache (accessed_at);
	CREATE INDEX ON storage.cache (tag);
`
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
}
