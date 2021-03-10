package chunk

import (
	"context"
	fmt "fmt"
	"path"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	store   kv.Store
	db      *sqlx.DB
	tracker track.Tracker
	renewer *track.Renewer
	ttl     time.Duration
}

// NewClient returns a client which will write to objc, mdstore, and tracker.  Name is used
// for the set of temporary objects
func NewClient(store kv.Store, db *sqlx.DB, tr track.Tracker, name string) Client {
	var renewer *track.Renewer
	if name != "" {
		renewer = track.NewRenewer(tr, name, defaultChunkTTL)
	}
	c := &trackedClient{
		store:   store,
		tracker: tr,
		renewer: renewer,
		ttl:     defaultChunkTTL,
	}
	return c
}

// Create creates a new chunk from metadata and chunkData.
// It returns the ID for the chunk
func (c *trackedClient) Create(ctx context.Context, md Metadata, chunkData []byte) (_ ID, retErr error) {
	chunkID := Hash(chunkData)
	var pointsTo []string
	for _, cid := range md.PointsTo {
		pointsTo = append(pointsTo, cid.TrackerID())
	}
	chunkOID := chunkID.TrackerID()
	var needUpload bool
	var gen uint64
	if err := dbutil.WithTx(ctx, c.db, func(tx *sqlx.Tx) error {
		if err := c.tracker.CreateTx(tx, chunkOID, pointsTo, c.ttl); err != nil {
			return err
		}
		var ents []Entry
		if err := tx.Select(&ents, `
		SELECT hash_id, gen
		FROM storage.chunk_objects
		WHERE uploaded = TRUE AND tombstone = FALSE AND chunk_id = $1`, chunkID); err != nil {
			return err
		}
		if len(ents) > 0 {
			needUpload = false
			return nil
		}
		if err := tx.Get(&gen, `
		INSERT INTO storage.chunk_objects (chunk_id, size)
		VALUES ($1, $2)
		RETURNING gen
		`, chunkID, md.Size); err != nil {
			return err
		}
		needUpload = true
		return nil
	}); err != nil {
		return nil, err
	}
	if err := c.renewer.Add(ctx, chunkOID); err != nil {
		return nil, err
	}
	if !needUpload {
		return chunkID, nil
	}
	// TODO: need to check for existence of specific objects.
	key := chunkKey(chunkID, gen)
	if exists, err := c.store.Exists(ctx, key); err != nil {
		return nil, err
	} else if exists {
		return chunkID, nil
	}
	if err := c.store.Put(ctx, key, chunkData); err != nil {
		return nil, err
	}
	return chunkID, nil
}

// Get writes data for a chunk with ID chunkID to w.
func (c *trackedClient) Get(ctx context.Context, chunkID ID, cb kv.ValueCallback) (retErr error) {
	var gen uint64
	err := c.db.Get(&gen, `
	SELECT gen
	FROM storage.chunk_objects
	WHERE chunk_id = %1
	LIMIT 1
	`, chunkID)
	if err != nil {
		err = errors.Errorf("no objects for chunk %v", chunkID)
		return err
	}
	key := chunkKey(chunkID, gen)
	return c.store.Get(ctx, key, cb)
}

// Close closes the client, stopping the background renewal of created objects
func (c *trackedClient) Close() error {
	if c.renewer != nil {
		return c.renewer.Close()
	}
	return nil
}

func chunkPath(chunkID ID, gen uint64) string {
	if len(chunkID) == 0 {
		panic("chunkID cannot be empty")
	}
	return path.Join(prefix, fmt.Sprintf("%s.%016x", chunkID, gen))
}

func chunkKey(chunkID ID, gen uint64) []byte {
	return []byte(chunkPath(chunkID, gen))
}

var _ track.Deleter = &deleter{}

type deleter struct {
}

func (d *deleter) DeleteTx(tx *sqlx.Tx, id string) error {
	if !strings.HasPrefix(id, prefix+"/") {
		return errors.Errorf("cannot delete (%s)", id)
	}
	chunkID, err := IDFromHex(id[len(TrackerPrefix):])
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		UPDATE storage.chunk_objects
		SET tombstone = TRUE
		WHERE chunk_id = $1
	`, chunkID)
	return err
}
