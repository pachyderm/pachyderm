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
