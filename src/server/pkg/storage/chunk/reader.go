package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"math"
	"path"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

type ReaderFunc func() ([]*DataRef, error)

// Reader reads a set of DataRefs from chunk storage.
type Reader struct {
	ctx      context.Context
	objC     obj.Client
	dataRefs []*DataRef
	curr     *DataRef
	buf      *bytes.Buffer
	r        *bytes.Reader
	len      int64
	f        ReaderFunc
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
			var err error
			r.dataRefs, err = r.f()
			if err != nil {
				return err
			}
		} else {
			return io.EOF
		}
	}
	// Get next chunk if necessary.
	if r.curr == nil || r.curr.Chunk.Hash != r.dataRefs[0].Chunk.Hash {
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

// WriteToN writes n bytes from the reader to the passed in writer. These writes are
// data reference copies when full chunks are being written to the writer.
func (r *Reader) WriteToN(w *Writer, n int64) error {
	for {
		// Read from the current data reference first.
		if r.r.Len() > 0 {
			nCopied, err := io.CopyN(w, r, int64(math.Min(float64(n), float64(r.r.Len()))))
			n -= nCopied
			if err != nil {
				return err
			}
		}
		// Done when there are no bytes left to write.
		if n == 0 {
			return nil
		}
		// A data reference can be cheaply copied when:
		// - The writer is at a split point.
		// - The data reference is a full chunk reference.
		// - The size of the chunk is less than or equal to the number of bytes left.
		if w.atSplit() {
			for r.dataRefs[0].Hash == "" && r.dataRefs[0].SizeBytes <= n {
				if err := w.writeChunk(r.dataRefs[0]); err != nil {
					return err
				}
				n -= r.dataRefs[0].SizeBytes
				r.dataRefs = r.dataRefs[1:]
			}
		}
		// Setup next data reference for reading.
		if err := r.nextDataRef(); err != nil {
			return err
		}
	}
}

// Close closes the reader.
// Currently a no-op, but will be used when streaming is implemented.
func (r *Reader) Close() error {
	return nil
}
