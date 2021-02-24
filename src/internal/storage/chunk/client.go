package chunk

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

// Client allows manipulation of individual chunks, by maintaining consistency between
// a tracker and an kv.Store
type Client struct {
	store   kv.Store
	mdstore MetadataStore
	tracker track.Tracker
	renewer *track.Renewer
	ttl     time.Duration
}

// NewClient returns a client which will write to objc, mdstore, and tracker.  Name is used
// for the set of temporary objects
func NewClient(store kv.Store, mdstore MetadataStore, tr track.Tracker, name string) *Client {
	var renewer *track.Renewer
	if name != "" {
		renewer = track.NewRenewer(tr, name, defaultChunkTTL)
	}
	c := &Client{
		store:   store,
		tracker: tr,
		mdstore: mdstore,
		renewer: renewer,
		ttl:     defaultChunkTTL,
	}
	return c
}

// Create creates a new chunk from metadata and chunkData.
// It returns the ID for the chunk
func (c *Client) Create(ctx context.Context, md Metadata, chunkData []byte) (_ ID, retErr error) {
	chunkID := Hash(chunkData)
	var pointsTo []string
	for _, cid := range md.PointsTo {
		pointsTo = append(pointsTo, ObjectID(cid))
	}
	// TODO: retry on ErrTombstone
	chunkOID := ObjectID(chunkID)
	if err := c.tracker.CreateObject(ctx, chunkOID, pointsTo, c.ttl); err != nil {
		if err != track.ErrObjectExists {
			return nil, err
		}
	}
	if err := c.renewer.Add(ctx, chunkOID); err != nil {
		return nil, err
	}
	if err := c.mdstore.Set(ctx, chunkID, md); err != nil && err != ErrMetadataExists {
		return nil, err
	}
	key := chunkKey(chunkID)
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
func (c *Client) Get(ctx context.Context, chunkID ID, cb kv.ValueCallback) (retErr error) {
	key := chunkKey(chunkID)
	return c.store.Get(ctx, key, cb)
}

// Close closes the client, stopping the background renewal of created objects
func (c *Client) Close() error {
	if c.renewer != nil {
		return c.renewer.Close()
	}
	return nil
}

func chunkPath(chunkID ID) string {
	if len(chunkID) == 0 {
		panic("chunkID cannot be empty")
	}
	return path.Join(prefix, chunkID.HexString())
}

func chunkKey(chunkID ID) []byte {
	return []byte(chunkPath(chunkID))
}

// ObjectID returns an object ID for use with a tracker
func ObjectID(chunkID ID) string {
	return prefix + "/" + chunkID.HexString()
}

var _ track.Deleter = &deleter{}

type deleter struct {
	mdstore MetadataStore
	store   kv.Store
}

func (d *deleter) Delete(ctx context.Context, id string) error {
	if !strings.HasPrefix(id, prefix+"/") {
		return errors.Errorf("cannot delete (%s)", id)
	}
	chunkID, err := IDFromHex(id[len(TrackerPrefix):])
	if err != nil {
		return err
	}
	if err := d.store.Delete(ctx, chunkKey(chunkID)); err != nil {
		return err
	}
	return d.mdstore.Delete(ctx, chunkID)
}
