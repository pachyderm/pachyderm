/*
Package fileset provides access to files through file sets.

A file set is the basic unit of storage for files. File sets come in two types:
primitive and composite. A primitive file set stores a list of files added and
a list of files deleted. A composite file set composes multiple composite and /
or primitive file sets into layers which can be merged to produce a new file
set.  Composite file sets allow file operations to be applied lazily, which can
be very beneficial for improving write / storage costs.

File sets are ephemeral by default, so unreferenced file sets will eventually
be garbage collected. Processing that involves creating and using file sets
should be done inside a renewer to keep the file sets alive while the
processing is occurring.

File sets can be created with a file set writer and files can be retrieved by
opening a file set with optional indexing and filtering, then iterating through
the files. Files must be added to a file set in lexicographical order and files
are emitted by a file set in lexicographical order. Various file set wrappers
are available that can perform computations and provide filtering based on the
files emitted by a file set.

Backwards compatibility:

The algorithms for file and index chunking have changed throughout 2.x, and we
must support previously written data. Here are some examples of conditions in
past data which current code will not generate:
  - a file that is split across multiple chunks may share some of them with
    other files (in the current code, a file split across chunks will be the only
    file in those chunks).
  - related, even small files may be split across multiple chunks
  - an index range data reference may start part way through a chunk
*/
package fileset

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// ID is the unique identifier for a fileset
// ID's should never be sent over the wire. Do not add serialization or HexString methods to this type.
// IDs are used by PFS and PJS to pin Filesets, and relate them to other objects in the system.
//
// TODO: change this to a pachhash
type ID [16]byte

func newID() ID {
	id := ID{}
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	return id
}

func parseID(x string) (ID, error) {
	return parse16byte([]byte(x))
}

// TrackerID returns the tracker ID for the fileset.
func (id ID) TrackerID() string {
	return TrackerPrefix + id.hexString()
}

func (id ID) hexString() string {
	return hex.EncodeToString(id[:])
}

// Scan implements sql.Scanner
func (id *ID) Scan(src interface{}) error {
	var err error
	switch x := src.(type) {
	case []byte:
		*id, err = parse16byte(x)
	case string:
		*id, err = parse16byte([]byte(x))
	default:
		return errors.Errorf("scanning fileset.ID: can't turn %T into fileset.ID", src)
	}
	return err
}

// Value implements sql.Valuer
func (id ID) Value() (driver.Value, error) {
	return id.hexString(), nil
}

// Handles are given out through the API to access Filesets.
type Handle struct {
	id ID
}

// ParseID parses a string into an ID or returns an error
func ParseHandle(x string) (*Handle, error) {
	id, err := parse16byte([]byte(x))
	if err != nil {
		return nil, err
	}
	return &Handle{id: id}, nil
}

func (h Handle) HexString() string {
	return hex.EncodeToString(h.id[:])
}

func (h Handle) String() string {
	return h.HexString()
}

// ParseID parses a string into an ID or returns an error
func ParseHandles(xs []string) (ret []Handle, _ error) {
	for _, x := range xs {
		h, err := ParseHandle(x)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *h)
	}
	return ret, nil
}

func HandlessToHexStrings(hs []Handle) []string {
	var xs []string
	for _, h := range hs {
		xs = append(xs, h.String())
	}
	return xs
}

func parse16byte(x []byte) ([16]byte, error) {
	x = bytes.Replace(x, []byte{'-'}, []byte{}, -1)
	id := ID{}
	if len(x) < 32 {
		return ID{}, errors.Errorf("parsing fileset.ID: too short to be ID len=%d", len(x))
	}
	if len(x) > 32 {
		return ID{}, errors.Errorf("parsing fileset.ID: too long to be ID len=%d", len(x))
	}
	_, err := hex.Decode(id[:], x)
	if err != nil {
		return ID{}, errors.EnsureStack(err)
	}
	return id, nil
}

// PointsTo returns a slice of the chunk.IDs which this fileset immediately points to.
// Transitively reachable chunks are not included in the slice.
func (p *Primitive) PointsTo() []chunk.ID {
	var ids []chunk.ID
	ids = append(ids, index.PointsTo(p.Additive)...)
	ids = append(ids, index.PointsTo(p.Deletive)...)
	return ids
}

// PointsTo returns the IDs of the filesets which this composite fileset points to
func (c *Composite) PointsTo() ([]ID, error) {
	ids := make([]ID, len(c.Layers))
	for i := range c.Layers {
		id, err := parse16byte([]byte(c.Layers[i]))
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}

// File represents a file.
type File interface {
	// Index returns the index for the file.
	Index() *index.Index
	// Content writes the content of the file.
	Content(ctx context.Context, w io.Writer, opts ...chunk.ReaderOption) error
	// Hash returns the hash of the file.
	Hash(ctx context.Context) ([]byte, error)
}

var _ File = &MergeFileReader{}
var _ File = &FileReader{}

// FileSet represents a set of files.
type FileSet interface {
	// Iterate iterates over the files in the file set.
	Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error
	// IterateDeletes iterates over the deleted files in the file set.
	IterateDeletes(ctx context.Context, cb func(File) error, opts ...index.Option) error
	// Shards returns a list of shards for the file set.
	Shards(ctx context.Context, opts ...index.Option) ([]*index.PathRange, error)
}

var _ FileSet = &MergeReader{}
var _ FileSet = &Reader{}

type emptyFileSet struct{}

func (efs emptyFileSet) Iterate(_ context.Context, _ func(File) error, _ ...index.Option) error {
	return nil
}

func (efs emptyFileSet) IterateDeletes(_ context.Context, _ func(File) error, _ ...index.Option) error {
	return nil
}

func (efs emptyFileSet) Shards(_ context.Context, _ ...index.Option) ([]*index.PathRange, error) {
	return []*index.PathRange{{}}, nil
}

func idsFromHandles(hs []Handle) []ID {
	return mapSlice(hs, func(h Handle) ID {
		return h.id
	})
}

func mapSlice[X, Y any](xs []X, fn func(X) Y) []Y {
	ys := make([]Y, len(xs))
	for i := range xs {
		ys[i] = fn(xs[i])
	}
	return ys
}
