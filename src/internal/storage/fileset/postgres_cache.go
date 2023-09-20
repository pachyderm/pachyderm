package fileset

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// CreatePostgresCacheV1 creates the table for a cache.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CreatePostgresCacheV1(ctx context.Context, tx *pachsql.Tx) error {
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

const CacheTrackerPrefix = "cache/"

func cacheTrackerKey(key string) string {
	return CacheTrackerPrefix + key
}

type Cache struct {
	db      *pachsql.DB
	tracker track.Tracker
	maxSize int
}

func NewCache(db *pachsql.DB, tracker track.Tracker, maxSize int) *Cache {
	return &Cache{
		db:      db,
		tracker: tracker,
		maxSize: maxSize,
	}
}

func (c *Cache) Put(ctx context.Context, key string, value *anypb.Any, ids []ID, tag string) error {
	data, err := proto.Marshal(value)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return dbutil.WithTx(ctx, c.db, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := c.put(tx, key, data, ids, tag); err != nil {
			return err
		}
		return c.applyEvictionPolicy(tx)
	})
}

func (c *Cache) put(tx *pachsql.Tx, key string, value []byte, ids []ID, tag string) error {
	if ids == nil {
		ids = []ID{}
	}
	_, err := tx.Exec(`
		INSERT INTO storage.cache (key, value_pb, ids, tag)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO NOTHING
	`, key, value, ids, tag)
	if err != nil {
		return errors.EnsureStack(err)
	}
	var pointsTo []string
	for _, id := range ids {
		pointsTo = append(pointsTo, id.TrackerID())
	}
	return errors.EnsureStack(c.tracker.CreateTx(tx, cacheTrackerKey(key), pointsTo, track.NoTTL))
}

func (c *Cache) applyEvictionPolicy(tx *pachsql.Tx) error {
	var size int
	if err := tx.Get(&size, `
		SELECT COUNT(key)
		FROM storage.cache
	`); err != nil {
		return errors.EnsureStack(err)
	}
	if size <= c.maxSize {
		return nil
	}
	var key string
	if err := tx.Get(&key, `
		DELETE FROM storage.cache
		WHERE key IN (
			SELECT key
			FROM storage.cache
			ORDER BY accessed_at
			LIMIT 1
		)
		RETURNING key
	`); err != nil {
		return errors.EnsureStack(err)
	}
	return c.tracker.DeleteTx(tx, cacheTrackerKey(key))
}

func (c *Cache) Get(ctx context.Context, key string) (*anypb.Any, error) {
	data := []byte{}
	if err := sqlx.GetContext(ctx, c.db, &data, `
		UPDATE storage.cache
		SET accessed_at = CURRENT_TIMESTAMP
		WHERE key = $1
		RETURNING value_pb
	`, key); err != nil {
		return nil, errors.EnsureStack(err)
	}
	value := &anypb.Any{}
	if err := proto.Unmarshal(data, value); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return value, nil
}

func (c *Cache) Clear(ctx context.Context, tagPrefix string) error {
	var keys []string
	if err := c.db.SelectContext(ctx, &keys, `
		SELECT key
		FROM storage.cache
		WHERE tag LIKE $1 || '%'
	`, tagPrefix); err != nil {
		return errors.EnsureStack(err)
	}
	for _, key := range keys {
		if err := dbutil.WithTx(ctx, c.db, func(ctx context.Context, tx *pachsql.Tx) error {
			if _, err := tx.Exec(`
				DELETE FROM storage.cache
				WHERE key = $1
			`, key); err != nil {
				return errors.EnsureStack(err)
			}
			return c.tracker.DeleteTx(tx, cacheTrackerKey(key))
		}); err != nil {
			return err
		}
	}
	return nil
}
