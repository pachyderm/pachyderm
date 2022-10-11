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
type ID [16]byte

func newID() ID {
	id := ID{}
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	return id
}

// ParseID parses a string into an ID or returns an error
func ParseID(x string) (*ID, error) {
	id, err := parseID([]byte(x))
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// HexString returns the ID encoded with the hex alphabet.
func (id ID) HexString() string {
	return hex.EncodeToString(id[:])
}

// TrackerID returns the tracker ID for the fileset.
func (id ID) TrackerID() string {
	return TrackerPrefix + id.HexString()
}

// Scan implements sql.Scanner
func (id *ID) Scan(src interface{}) error {
	var err error
	switch x := src.(type) {
	case []byte:
		*id, err = parseID(x)
	case string:
		*id, err = parseID([]byte(x))
	default:
		return errors.Errorf("scanning fileset.ID: can't turn %T into fileset.ID", src)
	}
	return err
}

// Value implements sql.Valuer
func (id ID) Value() (driver.Value, error) {
	return id.HexString(), nil
}

func parseID(x []byte) (ID, error) {
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

func HexStringsToIDs(xs []string) ([]ID, error) {
	ids := []ID{}
	for _, x := range xs {
		id, err := ParseID(x)
		if err != nil {
			return nil, err
		}
		ids = append(ids, *id)
	}
	return ids, nil
}

func IDsToHexStrings(ids []ID) []string {
	var xs []string
	for _, id := range ids {
		xs = append(xs, id.HexString())
	}
	return xs
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
		id, err := ParseID(c.Layers[i])
		if err != nil {
			return nil, err
		}
		ids[i] = *id
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
	Shards(ctx context.Context) ([]*index.PathRange, error)
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

func (efs emptyFileSet) Shards(_ context.Context) ([]*index.PathRange, error) {
	return nil, nil
}

func idsToHex(xs []ID) []string {
	ys := make([]string, len(xs))
	for i := range xs {
		ys[i] = xs[i].HexString()
	}
	return ys
}
