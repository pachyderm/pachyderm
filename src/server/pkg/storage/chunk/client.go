package chunk

import (
	"context"
	io "io"
	"io/ioutil"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/track"
)

// Client allows manipulation of individual chunks, by maintaining consistency between
// a tracker and an obj.Client.
type Client struct {
	objc    obj.Client
	mdstore MetadataStore
	tracker track.Tracker
	renewer *track.Renewer
	ttl     time.Duration
}

// NewClient returns a client which will write to objc, mdstore, and tracker.  Name is used
// for the set of temporary objects
func NewClient(objc obj.Client, mdstore MetadataStore, tr track.Tracker, name string) *Client {
	var renewer *track.Renewer
	if name != "" {
		renewer = track.NewRenewer(tr, name, defaultChunkTTL)
	}
	c := &Client{
		objc:    objc,
		tracker: tr,
		mdstore: mdstore,
		renewer: renewer,
		ttl:     defaultChunkTTL,
	}
	return c
}

// Create creates a chunk with data from r and Metadata md
func (c *Client) Create(ctx context.Context, md Metadata, r io.Reader) (_ ID, retErr error) {
	chunkData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
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
	// at this point no one will be trying to delete the chunk, because there is an object pointing to it.
	p := chunkPath(chunkID)
	if c.objc.Exists(ctx, p) {
		return chunkID, nil
	}
	if err := c.mdstore.Set(ctx, chunkID, md); err != nil && err != ErrMetadataExists {
		return nil, err
	}
	objW, err := c.objc.Writer(ctx, p)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := objW.Close(); retErr == nil {
			retErr = err
		}
	}()
	if _, err = objW.Write(chunkData); err != nil {
		return nil, err
	}
	return chunkID, nil
}

// Get writes data for a chunk with ID chunkID to w.
func (c *Client) Get(ctx context.Context, chunkID ID, w io.Writer) (retErr error) {
	p := chunkPath(chunkID)
	objR, err := c.objc.Reader(ctx, p, 0, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err := objR.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(w, objR)
	return err
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

// ObjectID returns an object ID for use with a tracker
func ObjectID(chunkID ID) string {
	return prefix + "/" + chunkID.HexString()
}

var _ track.Deleter = &deleter{}

type deleter struct {
	mdstore MetadataStore
	objc    obj.Client
}

func (d *deleter) Delete(ctx context.Context, id string) error {
	if !strings.HasPrefix(id, prefix+"/") {
		return errors.Errorf("cannot delete (%s)", id)
	}
	chunkID, err := IDFromHex(id[len(prefix):])
	if err != nil {
		return err
	}
	if err := d.objc.Delete(ctx, chunkPath(chunkID)); err != nil {
		return err
	}
	return d.mdstore.Delete(ctx, chunkID)
}
