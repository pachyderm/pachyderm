package chunk

import (
	"context"
	fmt "fmt"
	io "io"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
)

// Client allows manipulation of individual chunks, by maintaining consistency between
// a tracker and an obj.Client.
type Client struct {
	objc     obj.Client
	mdstore  MetadataStore
	tracker  tracker.Tracker
	chunkSet string
	ttl      time.Duration

	mu sync.Mutex
	n  int

	cancel context.CancelFunc
	err    error
	done   chan struct{}
}

func NewClient(objc obj.Client, mdstore MetadataStore, tracker tracker.Tracker, chunkSet string) *Client {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := &Client{
		objc:     objc,
		tracker:  tracker,
		mdstore:  mdstore,
		chunkSet: chunkSet,
		ttl:      defaultChunkTTL,
		done:     make(chan struct{}),
		cancel:   cancel,
	}
	if chunkSet != "" {
		go func() {
			c.err = c.runLoop(ctx)
			close(c.done)
		}()
	} else {
		close(c.done)
	}
	return c
}

func (c *Client) Create(ctx context.Context, md Metadata, r io.Reader) (ID, error) {
	chunkData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	chunkID := Hash(chunkData)
	var pointsTo []string
	for _, cid := range md.PointsTo {
		pointsTo = append(pointsTo, ChunkObjectID(cid))
	}
	// TODO: retry on ErrTombstone
	chunkOID := ChunkObjectID(chunkID)
	if err := c.tracker.CreateObject(ctx, chunkOID, pointsTo, c.ttl); err != nil {
		if err != tracker.ErrObjectExists {
			return nil, err
		}
	}
	// create an object whos sole purpose is to reference the chunk we created, and to have a structured name
	// which can be renewed in bulk by prefix
	n := c.getInt()
	if err := c.tracker.CreateObject(ctx, fmt.Sprintf("client-tmp/%s/%d", c.chunkSet, n), []string{chunkOID}, c.ttl); err != nil {
		return nil, err
	}
	// at this point no one will be trying to delete the chunk, because there is an object pointing to it.
	p := chunkPath(chunkID)
	if c.objc.Exists(ctx, p) {
		return chunkID, nil
	}
	if err := c.mdstore.Set(ctx, chunkID, md); err != nil {
		return nil, err
	}
	objW, err := c.objc.Writer(ctx, p)
	if err != nil {
		return nil, err
	}
	if _, err = objW.Write(chunkData); err != nil {
		return nil, err
	}
	if err := objW.Close(); err != nil {
		return nil, err
	}
	return chunkID, nil
}

func (c *Client) Get(ctx context.Context, chunkID ID, w io.Writer) error {
	p := chunkPath(chunkID)
	objR, err := c.objc.Reader(ctx, p, 0, 0)
	if err != nil {
		return err
	}
	defer objR.Close()
	_, err = io.Copy(w, objR)
	return err
}

func (c *Client) Close() error {
	c.cancel()
	<-c.done
	return c.err
}

func (c *Client) runLoop(ctx context.Context) error {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := c.tracker.SetTTLPrefix(ctx, fmt.Sprintf("client-tmp/%s/", c.chunkSet), c.ttl); err != nil {
				return err
			}
		}
	}
}

func (c *Client) getInt() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
	return c.n
}

func chunkPath(chunkID ID) string {
	if len(chunkID) == 0 {
		panic("chunkID cannot be empty")
	}
	return path.Join(prefix, chunkID.HexString())
}

func ChunkObjectID(chunkID ID) string {
	return "chunk/" + chunkID.HexString()
}

var _ tracker.Deleter = &deleter{}

type deleter struct {
	mdstore MetadataStore
	objc    obj.Client
}

func (d *deleter) Delete(ctx context.Context, id string) error {
	const prefix = "chunks/"
	if !strings.HasPrefix(id, prefix) {
		return errors.Errorf("cannot delete (%s)", id)
	}
	chunkID, err := ChunkIDFromHex(id[len(prefix):])
	if err != nil {
		return err
	}
	if err := d.objc.Delete(ctx, chunkPath(chunkID)); err != nil {
		return err
	}
	return d.mdstore.Delete(ctx, chunkID)
}
