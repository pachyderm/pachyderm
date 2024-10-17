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
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"google.golang.org/protobuf/proto"
)

type Token [16]byte

func newToken() Token {
	token := Token{}
	if _, err := rand.Read(token[:]); err != nil {
		panic(err)
	}
	return token
}

// HexString returns the token encoded with the hex alphabet.
func (t Token) HexString() string {
	return hex.EncodeToString(t[:])
}

// TrackerID returns the tracker ID for the fileset.
func (t Token) TrackerID() string {
	return TrackerPrefix + t.HexString()
}

// Scan implements sql.Scanner
func (t *Token) Scan(src interface{}) error {
	var err error
	switch tokenData := src.(type) {
	case []byte:
		*t, err = parseToken(tokenData)
	case string:
		*t, err = parseToken([]byte(tokenData))
	default:
		return errors.Errorf("scanning fileset.Token: can't turn %T into fileset.Token", src)
	}
	return err
}

// Value implements sql.Valuer
func (t Token) Value() (driver.Value, error) {
	return t.HexString(), nil
}

func parseToken(tokenBytes []byte) (Token, error) {
	tokenBytes = bytes.Replace(tokenBytes, []byte{'-'}, []byte{}, -1)
	token := Token{}
	if len(tokenBytes) < 32 {
		return Token{}, errors.Errorf("parsing fileset.Token: too short to be Token len=%d", len(tokenBytes))
	}
	if len(tokenBytes) > 32 {
		return Token{}, errors.Errorf("parsing fileset.Token: too long to be Token len=%d", len(tokenBytes))
	}
	_, err := hex.Decode(token[:], tokenBytes)
	if err != nil {
		return Token{}, errors.EnsureStack(err)
	}
	return token, nil
}

type ID pachhash.Output

// HexString returns the id encoded with the hex alphabet.
func (i ID) HexString() string {
	return hex.EncodeToString(i[:])
}

func computeId(tx *pachsql.Tx, store MetadataStore, md *Metadata) (ID, error) {
	switch v := md.Value.(type) {
	case *Metadata_Primitive:
		data, err := proto.Marshal(v.Primitive)
		if err != nil {
			return ID{}, errors.EnsureStack(err)
		}
		return pachhash.Sum(data), nil
	case *Metadata_Composite:
		hasher := pachhash.New()
		// TODO: Use fileset table entries as a cache when it is migrated.
		for _, layer := range v.Composite.Layers {
			token, err := parseToken([]byte(layer))
			if err != nil {
				return ID{}, err
			}
			md, err := store.GetTx(tx, token)
			if err != nil {
				return ID{}, err
			}
			id, err := computeId(tx, store, md)
			if err != nil {
				return ID{}, err
			}
			hasher.Write(id[:])
		}
		id := ID{}
		copy(id[:], hasher.Sum(nil))
		return id, nil
	default:
		return ID{}, errors.Errorf("cannot compute id of type %T", md.Value)
	}
}

type Handle struct {
	token Token
	// Stable fileset ID (optional)
	id ID
}

// NewHandle creates a fileset handle with the provided token.
// TODO: Remove NewHandle and Token when Pin is implemented.
func NewHandle(token Token) *Handle {
	return &Handle{token: token}
}
func (h *Handle) Token() Token {
	return h.token
}

// HexString returns the handle encoded with the hex alphabet.
func (h *Handle) HexString() string {
	hexStr := h.token.HexString()
	defaultId := ID{}
	if h.id != defaultId {
		hexStr += "." + h.id.HexString()
	}
	return hexStr
}

// ParseHandle parses the string encoding of a fileset handle.
// The string encoding of a fileset handle is token[.id].
func ParseHandle(handle string) (*Handle, error) {
	tokenStr, _, _ := strings.Cut(handle, ".")
	token, err := parseToken([]byte(tokenStr))
	if err != nil {
		return nil, err
	}
	return &Handle{token: token}, nil
}

func HexStringsToHandles(hexStrs []string) ([]*Handle, error) {
	handles := []*Handle{}
	for _, hexStr := range hexStrs {
		handle, err := ParseHandle(hexStr)
		if err != nil {
			return nil, err
		}
		handles = append(handles, handle)
	}
	return handles, nil
}

func HandlesToHexStrings(handles []*Handle) []string {
	var hexStrs []string
	for _, handle := range handles {
		hexStrs = append(hexStrs, handle.HexString())
	}
	return hexStrs
}

func handlesToTokenHexStrings(handles []*Handle) []string {
	var hexStrs []string
	for _, handle := range handles {
		hexStrs = append(hexStrs, handle.token.HexString())
	}
	return hexStrs
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
func (c *Composite) PointsTo() ([]*Handle, error) {
	handles := make([]*Handle, len(c.Layers))
	for i := range c.Layers {
		handle, err := ParseHandle(c.Layers[i])
		if err != nil {
			return nil, err
		}
		handles[i] = handle
	}
	return handles, nil
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
