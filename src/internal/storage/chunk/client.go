package chunk

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	fmt "fmt"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// Client mediates access to a content-addressed store
type Client interface {
	Create(ctx context.Context, md Metadata, chunkData []byte) (ID, error)
	Get(ctx context.Context, chunkID ID, cb kv.ValueCallback) error
	Close() error
}

// trackedClient allows manipulation of individual chunks, by maintaining consistency between
// a tracker and an kv.Store
type trackedClient struct {
	store        kv.Store
	pool         *kv.Pool
	db           *pachsql.DB
	tracker      track.Tracker
	maxChunkSize int
	renewer      *Renewer
	ttl          time.Duration
}

// NewClient returns a client which will write to objc, mdstore, and tracker.  Name is used
// for the set of temporary objects
func NewClient(store kv.Store, db *pachsql.DB, tr track.Tracker, renewer *Renewer, pool *kv.Pool) Client {
	return &trackedClient{
		store:        store,
		pool:         pool,
		db:           db,
		tracker:      tr,
		maxChunkSize: DefaultMaxChunkSize,
		renewer:      renewer,
		ttl:          defaultChunkTTL,
	}
}

// Create creates a new chunk from metadata and chunkData.
// It returns the ID for the chunk
func (c *trackedClient) Create(ctx context.Context, md Metadata, chunkData []byte) (_ ID, retErr error) {
	ctx = pctx.Child(ctx, "trackedClient")
	if c.renewer == nil {
		panic("client must have a renewer to create chunks")
	}
	if len(chunkData) > c.maxChunkSize {
		return nil, errors.Errorf("data len=%d exceeds max chunk size %d", len(chunkData), c.maxChunkSize)
	}
	chunkID := Hash(chunkData)
	needUpload, gen, err := c.beforeUpload(ctx, chunkID, md)
	if err != nil {
		return nil, err
	}
	if err := c.renewer.Add(ctx, chunkID); err != nil {
		return nil, err
	}
	key := chunkKey(chunkID, gen)
	if !needUpload {
		if err := c.errIfNotExists(ctx, key); err != nil {
			return nil, err
		}
		return chunkID, nil
	}
	if err := c.store.Put(ctx, key, chunkData); err != nil {
		return nil, errors.EnsureStack(err)
	}
	if err := c.errIfNotExists(ctx, key); err != nil {
		return nil, err
	}
	if err := c.afterUpload(ctx, chunkID, gen); err != nil {
		return nil, err
	}
	return chunkID, nil
}

func (c *trackedClient) errIfNotExists(ctx context.Context, key []byte) error {
	ok, err := c.store.Exists(ctx, key)
	if err != nil {
		return errors.Wrap(err, "checking chunk existence")
	}
	if !ok {
		return errors.Errorf("chunk %s does not exist after attempting to upload it to the backend.", string(key))
	}
	return nil
}

// beforeUpload checks the table in postgres to see if a chunk with chunkID already exists.
func (c *trackedClient) beforeUpload(ctx context.Context, chunkID ID, md Metadata) (needUpload bool, gen uint64, _ error) {
	var pointsTo []string
	for _, cid := range md.PointsTo {
		pointsTo = append(pointsTo, cid.TrackerID())
	}
	chunkTID := chunkID.TrackerID()
	var ents []Entry
	if err := dbutil.WithTx(ctx, c.db, func(cbCtx context.Context, tx *pachsql.Tx) (retErr error) {
		needUpload, gen = false, 0
		if err := c.tracker.CreateTx(tx, chunkTID, pointsTo, c.ttl); err != nil {
			return errors.EnsureStack(err)
		}
		if err := tx.SelectContext(cbCtx, &ents, `
		SELECT chunk_id, gen
		FROM storage.chunk_objects
		WHERE uploaded = TRUE AND tombstone = FALSE AND chunk_id = $1`, chunkID); err != nil {
			return errors.EnsureStack(err)
		}
		if len(ents) > 0 {
			needUpload = false
			return nil
		}
		if err := tx.GetContext(cbCtx, &gen, `
		INSERT INTO storage.chunk_objects (chunk_id, size)
		VALUES ($1, $2)
		RETURNING gen
		`, chunkID, md.Size); err != nil {
			return errors.EnsureStack(err)
		}
		needUpload = true
		return nil
	}); err != nil {
		return false, 0, err
	}
	if len(ents) > 0 {
		return ents[0].Uploaded, ents[0].Gen, nil
	}
	return needUpload, gen, nil
}

// afterUpload marks the (chunkID, gen) pair as a successfully uploaded object in postgres.
func (c *trackedClient) afterUpload(ctx context.Context, chunkID ID, gen uint64) error {
	res, err := c.db.ExecContext(ctx, `
	UPDATE storage.chunk_objects
	SET uploaded = TRUE
	WHERE chunk_id = $1 AND gen = $2
	`, chunkID, gen)
	if err != nil {
		return errors.EnsureStack(err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if affected < 1 {
		return errors.Errorf("no chunk entry for object post upload: chunk=%v gen=%v", chunkID, gen)
	}
	if affected > 1 {
		panic("(chunk_id, gen) is not unique")
	}
	return nil
}

// Get writes data for a chunk with ID chunkID to w.
func (c *trackedClient) Get(ctx context.Context, chunkID ID, cb kv.ValueCallback) error {
	var gen uint64
	err := c.db.Get(&gen, `
	SELECT gen
	FROM storage.chunk_objects
	WHERE uploaded = TRUE AND tombstone = FALSE AND chunk_id = $1
	LIMIT 1
	`, chunkID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.Errorf("no objects for chunk %v", chunkID)
		}
		return err
	}
	key := chunkKey(chunkID, gen)
	return errors.EnsureStack(c.pool.GetF(ctx, c.store, key, cb))
}

// Close closes the client, stopping the background renewal of created objects
func (c *trackedClient) Close() error {
	if c.renewer != nil {
		return c.renewer.Close()
	}
	return nil
}

// CheckEntries runs an integrity check on the objects in object storage.
// It lists through chunks with IDs >= first, lexicographically.
func (c *trackedClient) CheckEntries(ctx context.Context, first []byte, limit int, readChunks bool) (n int, last ID, _ error) {
	if first == nil {
		first = []byte{} // in SQL: nothing is comparable to nil
	}
	var ents []Entry
	if err := c.db.SelectContext(ctx, &ents,
		`SELECT chunk_id, gen, uploaded, tombstone FROM storage.chunk_objects
		WHERE chunk_id >= $1 AND uploaded = true AND tombstone = false
		ORDER BY chunk_id
		LIMIT $2
	`, first, limit); err != nil {
		return 0, nil, errors.EnsureStack(err)
	}
	for _, ent := range ents {
		if readChunks {
			if err := c.pool.GetF(ctx, c.store, chunkKey(ent.ChunkID, ent.Gen), func(data []byte) error {
				return verifyData(ent.ChunkID, data)
			}); err != nil {
				if pacherr.IsNotExist(err) {
					if exists, err := c.entryExists(ctx, ent.ChunkID, ent.Gen); err != nil {
						return n, nil, err
					} else if exists {
						return n, nil, newErrMissingObject(ent)
					}
				}
			}
		} else {
			exists, err := c.store.Exists(ctx, chunkKey(ent.ChunkID, ent.Gen))
			if err != nil {
				return n, nil, errors.EnsureStack(err)
			}
			if !exists {
				if exists2, err := c.entryExists(ctx, ent.ChunkID, ent.Gen); err != nil {
					return n, nil, err
				} else if exists2 {
					return n, nil, newErrMissingObject(ent)
				}
			}
		}
		last = ent.ChunkID
		n++
	}
	if n == limit && bytes.Equal(first, last) {
		return n, nil, errors.Errorf("limit too small to check all chunk entries limit=%d", limit)
	}
	return n, last, nil
}

func (c *trackedClient) entryExists(ctx context.Context, chunkID ID, gen uint64) (bool, error) {
	var x int
	err := c.db.GetContext(ctx, &x, `SELECT FROM storage.chunk_objects WHERE chunk_id = $1 AND gen = $2`, chunkID, gen)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, errors.EnsureStack(err)
	}
	return true, nil
}

func chunkKey(chunkID ID, gen uint64) (ret []byte) {
	if len(chunkID) == 0 {
		panic("chunkID cannot be empty")
	}
	return fmt.Appendf(ret, "%s.%016x", chunkID.HexString(), gen)
}

func parseKey(key []byte) (ID, uint64, error) {
	parts := bytes.SplitN(key, []byte("."), 2)
	if len(parts) < 2 {
		return nil, 0, errors.Errorf("invalid chunk key %q", key)
	}
	chunkID := make([]byte, hex.DecodedLen(len(parts[0])))
	n, err := hex.Decode(chunkID, parts[0])
	if err != nil {
		return nil, 0, errors.EnsureStack(err)
	}
	chunkID = chunkID[:n]
	gen, err := strconv.ParseUint(string(parts[1]), 16, 64)
	if err != nil {
		return nil, 0, errors.EnsureStack(err)
	}
	return chunkID, gen, nil
}

func newErrMissingObject(ent Entry) error {
	return errors.Errorf("missing object for chunk entry: chunkID=%v gen=%v uploaded=%v tombstone=%v", ent.ChunkID, ent.Gen, ent.Uploaded, ent.Tombstone)
}

var _ track.Deleter = &deleter{}

type deleter struct{}

func (d *deleter) DeleteTx(tx *pachsql.Tx, id string) error {
	chunkID, err := ParseTrackerID(id)
	if err != nil {
		return errors.Wrapf(err, "deleting chunk")
	}
	_, err = tx.Exec(`
		UPDATE storage.chunk_objects
		SET tombstone = TRUE
		WHERE chunk_id = $1
	`, chunkID)
	return errors.EnsureStack(err)
}
