package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx      context.Context
	objC     obj.Client
	dataRefs []*DataRef
	curr     *DataReader
}

func newReader(ctx context.Context, objC obj.Client) *Reader {
	return &Reader{
		ctx:  ctx,
		objC: objC,
	}
}

// NextDataRefs sets the next data references for the reader.
func (r *Reader) NextDataRefs(dataRefs []*DataRef) {
	r.dataRefs = dataRefs
}

func (r *Reader) Iterate(f func(*DataReader) error) error {
	for {
		if len(r.dataRefs) == 0 {
			return nil
		}
		r.curr = newDataReader(r.ctx, r.objC, r.dataRefs[0], r.curr)
		if err := f(r.curr); err != nil {
			return err
		}
		r.dataRefs = r.dataRefs[1:]
	}
	return nil
}

type DataReader struct {
	ctx     context.Context
	objC    obj.Client
	dataRef *DataRef
	chunk   []byte
	seed    *DataReader
}

func newDataReader(ctx context.Context, objC obj.Client, dataRef *DataRef, seed *DataReader) *DataReader {
	return &DataReader{
		ctx:     ctx,
		objC:    objC,
		dataRef: dataRef,
		seed:    seed,
	}
}

func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

func (dr *DataReader) Get(w io.Writer) error {
	if err := dr.getChunk(); err != nil {
		return err
	}
	data := dr.chunk[dr.dataRef.OffsetBytes : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
	_, err := w.Write(data)
	return err
}

func (dr *DataReader) getChunk() error {
	// Use seed chunk if possible.
	if dr.seed != nil && dr.dataRef.Chunk.Hash == dr.seed.dataRef.Chunk.Hash {
		dr.chunk = dr.seed.chunk
		return nil
	}
	objR, err := dr.objC.Reader(dr.ctx, path.Join(prefix, dr.dataRef.Chunk.Hash), 0, 0)
	if err != nil {
		return err
	}
	defer objR.Close()
	gzipR, err := gzip.NewReader(objR)
	if err != nil {
		return err
	}
	defer gzipR.Close()
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, gzipR); err != nil {
		return err
	}
	dr.chunk = buf.Bytes()
	return nil
}

func (r *Reader) Get(w io.Writer) error {
	return r.Iterate(func(dr *DataReader) error {
		return dr.Get(w)
	})
}

//// Copy is the basic data structure to represent a copy of data from
//// a reader to a writer.
//// before/after are the raw bytes that precede/follow full chunks
//// in the set of bytes represented by the copy.
//type Copy struct {
//	before, after *bytes.Buffer
//	chunkRefs     []*DataRef
//}
//
//// ReadCopy reads copy data from the reader.
//func (r *Reader) ReadCopy(n ...int64) (*Copy, error) {
//	totalLeft := int64(math.MaxInt64)
//	if len(n) > 0 {
//		totalLeft = n[0]
//		if r.Len() < totalLeft {
//			return nil, fmt.Errorf("reader length (%v) less than copy length (%v)", r.Len(), totalLeft)
//		}
//	}
//	rawCopy := func(w io.Writer, n int64) error {
//		if _, err := io.CopyN(w, r, n); err != nil {
//			return err
//		}
//		totalLeft -= n
//		return nil
//	}
//	c := &Copy{
//		before: &bytes.Buffer{},
//		after:  &bytes.Buffer{},
//	}
//	// Copy the first WindowSize bytes raw to be sure that
//	// the chunks we will copy will exist in the writer that
//	// this data is being copied to.
//	if err := rawCopy(c.before, mathutil.MinInt64(WindowSize, totalLeft)); err != nil {
//		return nil, err
//	}
//	// Copy the bytes left in the current chunk (if any)
//	if r.r.Len() > 0 {
//		if err := rawCopy(c.before, mathutil.MinInt64(int64(r.r.Len()), totalLeft)); err != nil {
//			return nil, err
//		}
//	}
//	// Copy the in between chunk references.
//	// (bryce) is there an edge case with a size zero chunk?
//	for len(r.dataRefs) > 0 && r.dataRefs[0].Hash == "" && totalLeft >= r.dataRefs[0].SizeBytes {
//		r.executeOnSplitFunc()
//		c.chunkRefs = append(c.chunkRefs, r.dataRefs[0])
//		totalLeft -= r.dataRefs[0].SizeBytes
//		r.len -= r.dataRefs[0].SizeBytes
//		r.dataRefs = r.dataRefs[1:]
//	}
//	// Copy the rest of the bytes raw.
//	if err := rawCopy(c.after, totalLeft); err != nil {
//		return nil, err
//	}
//	return c, nil
//}
