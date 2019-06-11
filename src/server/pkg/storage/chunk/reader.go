package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Reader reads a set of DataRefs from chunk storage.
type Reader struct {
	ctx      context.Context
	objC     obj.Client
	dataRefs []*DataRef
	curr     *DataRef
	buf      *bytes.Buffer
	r        *bytes.Reader
	len      int64
}

func newReader(ctx context.Context, objC obj.Client, dataRefs ...*DataRef) *Reader {
	r := &Reader{
		ctx:  ctx,
		objC: objC,
		buf:  &bytes.Buffer{},
	}
	r.NextRange(dataRefs)
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
			// If all DataRefs have been read, then io.EOF.
			if len(r.dataRefs) == 0 {
				return totalRead, io.EOF
			}
			if err := r.nextDataRef(); err != nil {
				return totalRead, err
			}
		}
	}
	return totalRead, nil

}

func (r *Reader) nextDataRef() error {
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

// Close closes the reader.
// Currently a no-op, but will be used when streaming is implemented.
func (r *Reader) Close() error {
	return nil
}
