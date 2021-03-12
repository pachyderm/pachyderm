package chunk

import (
	"bytes"
	"context"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/sirupsen/logrus"
)

// GarbageCollector removes unused chunks from object storage
type GarbageCollector struct {
	s *Storage
}

// NewGC returns a new garbage collector operating on s
func NewGC(s *Storage) *GarbageCollector {
	return &GarbageCollector{s: s}
}

// RunForever calls RunOnce until the context is cancelled, logging any errors.
func (gc *GarbageCollector) RunForever(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := gc.RunOnce(ctx); err != nil {
			select {
			case <-ctx.Done():
				return err
			default:
			}
			logrus.Errorf("during chunk GC: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// RunOnce runs 1 cycle of garbage collection.
func (gc *GarbageCollector) RunOnce(ctx context.Context) (retErr error) {
	rows, err := gc.s.db.QueryContext(ctx, `
	SELECT chunk_id, gen, uploaded FROM storage.chunk_objects
	WHERE tombstone = true
	`)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); retErr == nil {
			retErr = err
		}
	}()
	for rows.Next() {
		var ent Entry
		if err := sqlx.StructScan(rows, &ent); err != nil {
			return err
		}
		if !ent.Uploaded {
			logrus.Warnf("possibility for untracked chunk %s", chunkPath(ent.ChunkID, ent.Gen))
		}
		if err := gc.deleteOne(ctx, ent); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (gc *GarbageCollector) deleteOne(ctx context.Context, ent Entry) error {
	if err := gc.deleteObject(ctx, ent.ChunkID, ent.Gen); err != nil {
		return err
	}
	return gc.deleteEntry(ctx, ent.ChunkID, ent.Gen)
}

func (gc *GarbageCollector) deleteObject(ctx context.Context, chunkID ID, gen uint64) error {
	return gc.s.store.Delete(ctx, chunkKey(chunkID, gen))
}

func (gc *GarbageCollector) deleteEntry(ctx context.Context, chunkID ID, gen uint64) error {
	_, err := gc.s.db.ExecContext(ctx, `
	DELETE FROM storage.chunk_objects
	WHERE chunk_id = $1 AND gen = $2 AND tombstone = TRUE
	LIMIT 1
	`, chunkID, gen)
	return err
}

// ReduceObjectsPerChunk eliminates redundant objects
func ReduceObjectsPerChunk(ctx context.Context, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, `
	WITH keep as (
		SELECT chunk_id, max(gen) as gen
		FROM storage.chunk_objects
		GROUP BY chunk_id
	)
	UPDATE storage.chunk_objects
	SET tombstone = TRUE
	WHERE (chunk_id, gen) NOT IN (SELECT * FROM keep)
	`)
	return err
}

// GCObjects walks store to delete untracked objects.
func GCObjects(ctx context.Context, db *sqlx.DB, store kv.Store) error {
	return store.Walk(ctx, []byte("chunk/"), func(key []byte) error {

		chunkID, gen, err := parseKey(key)
		if err != nil {
			logrus.Error(err)
			return nil
		}
		var count int
		if err := db.GetContext(ctx, &count, `
		SELECT count(1) FROM storage.chunk_objects
		WHERE chunkID = $1 AND gen = $2 AND tombstone = FALSE
		`, chunkID, gen); err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		return store.Delete(ctx, key)
	})
}

func parseKey(x []byte) (ID, uint64, error) {
	parts := bytes.SplitN(x, []byte("."), 2)
	if len(parts) < 2 {
		return nil, 0, errors.Errorf("invalid key")
	}
	id := make([]byte, hex.DecodedLen(len(parts[0])))
	_, err := hex.Decode(id, parts[0])
	if err != nil {
		return nil, 0, err
	}
	gen, err := strconv.ParseUint(string(parts[1]), 16, 64)
	if err != nil {
		return nil, 0, err
	}
	return id, gen, nil
}
