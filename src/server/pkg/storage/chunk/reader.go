package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math"
	"path"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"modernc.org/mathutil"
)

// ReaderFunc is a callback that returns the next set of data references
// to a reader.
type ReaderFunc func() ([]*DataRef, error)

// Reader reads a set of DataRefs from chunk storage.
type Reader struct {
	ctx              context.Context
	objC             obj.Client
	dataRefs         []*DataRef
	curr             *DataRef
	buf              *bytes.Buffer
	r                *bytes.Reader
	len              int64
	f                ReaderFunc
	onSplit          func()
	bytesBeforeSplit int64
}

func newReader(ctx context.Context, objC obj.Client, f ...ReaderFunc) *Reader {
	r := &Reader{
		ctx:  ctx,
		objC: objC,
		buf:  &bytes.Buffer{},
	}
	if len(f) > 0 {
		r.f = f[0]
	}
	return r
}

// NextRange sets the next range for the reader.
func (r *Reader) NextRange(dataRefs []*DataRef) {
	r.dataRefs = dataRefs
	r.r = bytes.NewReader([]byte{})
	r.len = 0
	for _, dataRef := range dataRefs {
		r.len += dataRef.SizeBytes
	}
}

// Len returns the number of bytes left.
func (r *Reader) Len() int64 {
	return r.len
}

// Read reads from the byte stream produced by the set of DataRefs.
func (r *Reader) Read(data []byte) (int, error) {
	var totalRead int
	defer func() {
		r.len -= int64(totalRead)
		r.bytesBeforeSplit += int64(totalRead)
	}()
	for len(data) > 0 {
		n, err := r.r.Read(data)
		data = data[n:]
		totalRead += n
		if err != nil {
			if err := r.nextDataRef(); err != nil {
				return totalRead, err
			}
		}
	}
	return totalRead, nil
}

func (r *Reader) nextDataRef() error {
	// If all DataRefs have been read, then io.EOF.
	if len(r.dataRefs) == 0 {
		if r.f != nil {
			dataRefs, err := r.f()
			if err != nil {
				return err
			}
			r.NextRange(dataRefs)
		} else {
			return io.EOF
		}
	}
	// Get next chunk if necessary.
	if r.curr == nil || r.curr.Chunk.Hash != r.dataRefs[0].Chunk.Hash {
		r.executeOnSplitFunc()
		if err := r.readChunk(r.dataRefs[0].Chunk); err != nil {
			return err
		}
	}
	r.curr = r.dataRefs[0]
	r.dataRefs = r.dataRefs[1:]
	r.r = bytes.NewReader(r.buf.Bytes()[r.curr.OffsetBytes : r.curr.OffsetBytes+r.curr.SizeBytes])
	return nil
}

func (r *Reader) readChunk(chunk *Chunk) error {
	objR, err := r.objC.Reader(r.ctx, path.Join(prefix, chunk.Hash), 0, 0)
	if err != nil {
		return err
	}
	defer objR.Close()
	gzipR, err := gzip.NewReader(objR)
	if err != nil {
		return err
	}
	defer gzipR.Close()
	r.buf.Reset()
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	if _, err := io.CopyBuffer(r.buf, gzipR, buf); err != nil {
		return err
	}
	return nil
}

// OnSplit registers a callback for when a chunk split point is encountered.
// The callback is only executed at a split point found after reading WindowSize bytes.
// The reason for this is to guarantee that the same split point will appear in the writer
// the data is being written to.
func (r *Reader) OnSplit(f func()) {
	r.bytesBeforeSplit = 0
	r.onSplit = f
}

func (r *Reader) executeOnSplitFunc() {
	if r.onSplit != nil && r.bytesBeforeSplit > WindowSize {
		r.onSplit()
		r.onSplit = nil
	}
}

// Copy is the basic data structure to represent a copy of data from
// a reader to a writer.
// before/after are the raw bytes that precede/follow full chunks
// in the set of bytes represented by the copy.
type Copy struct {
	before, after *bytes.Buffer
	chunkRefs     []*DataRef
}

// ReadCopy reads copy data from the reader.
func (r *Reader) ReadCopy(n ...int64) (*Copy, error) {
	totalLeft := int64(math.MaxInt64)
	if len(n) > 0 {
		totalLeft = n[0]
		if r.Len() < totalLeft {
			return nil, fmt.Errorf("reader length (%v) less than copy length (%v)", r.Len(), totalLeft)
		}
	}
	rawCopy := func(w io.Writer, n int64) error {
		if _, err := io.CopyN(w, r, n); err != nil {
			return err
		}
		totalLeft -= n
		return nil
	}
	c := &Copy{
		before: &bytes.Buffer{},
		after:  &bytes.Buffer{},
	}
	// Copy the first WindowSize bytes raw to be sure that
	// the chunks we will copy will exist in the writer that
	// this data is being copied to.
	if err := rawCopy(c.before, mathutil.MinInt64(WindowSize, totalLeft)); err != nil {
		return nil, err
	}
	// Copy the bytes left in the current chunk (if any)
	if r.r.Len() > 0 {
		if err := rawCopy(c.before, mathutil.MinInt64(int64(r.r.Len()), totalLeft)); err != nil {
			return nil, err
		}
	}
	// Copy the in between chunk references.
	// (bryce) is there an edge case with a size zero chunk?
	for len(r.dataRefs) > 0 && r.dataRefs[0].Hash == "" && totalLeft >= r.dataRefs[0].SizeBytes {
		r.executeOnSplitFunc()
		c.chunkRefs = append(c.chunkRefs, r.dataRefs[0])
		totalLeft -= r.dataRefs[0].SizeBytes
		r.len -= r.dataRefs[0].SizeBytes
		r.dataRefs = r.dataRefs[1:]
	}
	// Copy the rest of the bytes raw.
	if err := rawCopy(c.after, totalLeft); err != nil {
		return nil, err
	}
	return c, nil
}

// Close closes the reader.
// Currently a no-op, but will be used when streaming is implemented.
func (r *Reader) Close() error {
	return nil
}
