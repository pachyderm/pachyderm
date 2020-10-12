package chunk

import (
	"context"
	io "io"
	"path"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Client allows manipulation of individual chunks, by maintaining consistency between
// a tracker and an obj.Client.
// Client is not safe for concurrent use.
type Client struct {
	objc     obj.Client
	tracker  Tracker
	chunkSet string
	ttl      time.Duration

	mu   sync.Mutex // prevent mode switching during a call to SetTTL
	mode *bool

	cancel context.CancelFunc
	err    error
	done   chan struct{}
}

func NewClient(objc obj.Client, tracker Tracker, chunkSet string) *Client {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := &Client{
		objc:     objc,
		tracker:  tracker,
		chunkSet: chunkSet,
		ttl:      defaultChunkTTL,
		done:     make(chan struct{}),
		cancel:   cancel,
	}
	go func() {
		c.err = c.runLoop(ctx)
		close(c.done)
	}()
	return c
}

/*
	Create gets a lock on the chunk, does an existence check and then adds the metadata
	to the tracker and uploads the chunk data to object storage.  This ensures the data in
	the tracker is a superset of the data in object storage.
	Chunks locked to an upload set do not conflict, so multiple uploaders (writers) can get a lock
	in the event of a concurrent upload, there will be a race to write the same data to postgres (last write wins not an issue)
	and only the first writer will upload the object, the others will get exist errors which they can ignore.
	This will ensure that if a writer fails another writer will not skip the upload assuming someone has succeeded.

	No one will be deleting the chunk during the existence check, metadata put, or data upload.
*/

// Create adds a chunk with data provided by r and metadata md.
// TODO: it would be nice if this returned (ChunkID, error), and we hashed the data here.
func (c *Client) Create(ctx context.Context, chunkID ChunkID, md ChunkMetadata, r io.Reader) error {
	if err := c.ensureMode(ctx, false); err != nil {
		return err
	}
	// TODO: retry until done
	if err := c.tracker.LockChunk(ctx, c.chunkSet, chunkID); err != nil {
		return err
	}
	defer c.tracker.UnlockChunk(ctx, c.chunkSet, chunkID)
	// at this point no one will be trying to delete the chunk
	p := chunkPath(chunkID)
	if c.objc.Exists(ctx, p) {
		return nil
	}
	if err := c.tracker.SetChunkInfo(ctx, chunkID, md); err != nil {
		return err
	}
	objW, err := c.objc.Writer(ctx, p)
	if err != nil {
		return err
	}
	defer objW.Close()
	if _, err = io.Copy(objW, r); err != nil {
		return err
	}
	return objW.Close()
}

/*
	If there is a chunk in object storage, it is good to use, no lock required.
	Get should not be used to implement an existence check because conditionally
	uploading based on a check done without a lock is not okay.
*/

// Get writes the chunk data for a chunk with id chunkID to w.
func (c *Client) Get(ctx context.Context, chunkID ChunkID, w io.Writer) error {
	p := chunkPath(chunkID)
	objR, err := c.objc.Reader(ctx, p, 0, 0)
	if err != nil {
		return err
	}
	defer objR.Close()
	_, err = io.Copy(w, objR)
	return err
}

/*
	Delete ensures a delete mode chunk set exists for this client, then locks the chunk.
	It deletes the chunk from object storage first, then the metadata.
	This ensures the tracker has a superset of what is in object storage.
	After the chunk is gone the lock can be released, there is no point in
	maintaining a lock on a resource that doesn't exist.  This allows a creation to happen before the client is closed.

	If another client is trying to delete, or is uploading the chunk, delete will fail with ErrChunkLocked
*/

// Delete removes a chunk and metadata with ID chunkID
func (c *Client) Delete(ctx context.Context, chunkID ChunkID) error {
	if err := c.ensureMode(ctx, true); err != nil {
		return err
	}
	if err := c.tracker.LockChunk(ctx, c.chunkSet, chunkID); err != nil {
		return err
	}
	defer c.tracker.UnlockChunk(ctx, c.chunkSet, chunkID)
	p := chunkPath(chunkID)
	if err := c.objc.Delete(ctx, p); err != nil {
		return err
	}
	return c.tracker.DeleteChunkInfo(ctx, chunkID)
}

func (c *Client) Close() error {
	c.cancel()
	<-c.done
	return c.err
}

func (c *Client) ensureMode(ctx context.Context, delete bool) error {
	const defaultTTL = 60 * time.Second
	if c.mode != nil && *c.mode == delete {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.tracker.DropChunkSet(ctx, c.chunkSet); err != nil {
		return err
	}
	if err := c.tracker.NewChunkSet(ctx, c.chunkSet, delete, defaultTTL); err != nil {
		return err
	}
	c.mode = &delete
	return nil
}

func (c *Client) runLoop(ctx context.Context) error {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			c.mu.Lock()
			if _, err := c.tracker.SetTTL(ctx, c.chunkSet, c.ttl); err != nil {
				c.mu.Unlock()
				return err
			}
			c.mu.Unlock()
		}
	}
}

func chunkPath(chunkID ChunkID) string {
	return path.Join(prefix, chunkID.HexString())
}
