package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx      context.Context
	objC     obj.Client
	dataRefs []*DataRef
	curr     *DataReader
}

func newReader(ctx context.Context, objC obj.Client, dataRefs ...*DataRef) *Reader {
	return &Reader{
		ctx:      ctx,
		objC:     objC,
		dataRefs: dataRefs,
	}
}

// NextDataRefs sets the next data references for the reader.
func (r *Reader) NextDataRefs(dataRefs []*DataRef) {
	r.dataRefs = dataRefs
}

func (r *Reader) Peek() (*DataReader, error) {
	if r.curr == nil || r.curr.Done() {
		if len(r.dataRefs) == 0 {
			return nil, io.EOF
		}
		r.curr = newDataReader(r.ctx, r.objC, r.dataRefs[0], r.curr)
		r.dataRefs = r.dataRefs[1:]
	}
	return r.curr, nil
}

func (r *Reader) Iterate(f func(*DataReader) error) error {
	for {
		dr, err := r.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := f(dr); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reader) Get(w io.Writer) error {
	return r.Iterate(func(dr *DataReader) error {
		return dr.Get(w)
	})
}

type DataReader struct {
	ctx     context.Context
	objC    obj.Client
	dataRef *DataRef
	chunk   []byte
	offset  int64
	tags    []*Tag
	seed    *DataReader
	done    bool
}

func newDataReader(ctx context.Context, objC obj.Client, dataRef *DataRef, seed *DataReader) *DataReader {
	return &DataReader{
		ctx:     ctx,
		objC:    objC,
		dataRef: dataRef,
		offset:  dataRef.OffsetBytes,
		seed:    seed,
	}
}

func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

func (dr *DataReader) Len() int64 {
	var size int64
	for _, tag := range dr.tags {
		size += tag.SizeBytes
	}
	return size
}

func (dr *DataReader) Peek() (*Tag, error) {
	if dr.Done() {
		return nil, io.EOF
	}
	return dr.tags[0], nil
}

func (dr *DataReader) Done() bool {
	return dr.done
}

func (dr *DataReader) Iterate(f func(*Tag, io.Reader) error, tagBound ...string) error {
	if err := dr.setup(); err != nil {
		return err
	}
	for {
		if len(dr.tags) == 0 {
			dr.done = true
			return nil
		}
		tag := dr.tags[0]
		if !BeforeBound(tag.Id, tagBound...) {
			return nil
		}
		if err := f(tag, bytes.NewReader(dr.chunk[dr.offset:dr.offset+tag.SizeBytes])); err != nil {
			return err
		}
		dr.tags = dr.tags[1:]
		dr.offset += tag.SizeBytes
	}
}

func (dr *DataReader) setup() error {
	if dr.chunk != nil {
		return nil
	}
	// Use seed chunk if possible.
	if dr.seed != nil && dr.dataRef.Chunk.Hash == dr.seed.dataRef.Chunk.Hash {
		dr.chunk = dr.seed.chunk
		return nil
	}
	return dr.getChunk()
}

func (dr *DataReader) getChunk() error {
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

// BeforeBound checks if the passed in string is before the string bound (exclusive).
// The string bound is optional, so if no string bound is passed then it returns true.
func BeforeBound(str string, strBound ...string) bool {
	return len(strBound) == 0 || strings.Compare(str, strBound[0]) < 0
}

func (dr *DataReader) Get(w io.Writer) error {
	if err := dr.setup(); err != nil {
		return err
	}
	data := dr.chunk[dr.dataRef.OffsetBytes : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
	if _, err := w.Write(data); err != nil {
		return err
	}
	dr.done = true
	return nil
}

func (dr *DataReader) LimitReader(tagBound ...string) (*DataReader, error) {
	var tags []*Tag
	if err := dr.Iterate(func(tag *Tag, _ io.Reader) error {
		tags = append(tags, tag)
		return nil
	}, tagBound...); err != nil {
		return nil, err
	}
	return &DataReader{
		ctx:     dr.ctx,
		objC:    dr.objC,
		dataRef: dr.dataRef,
		offset:  dr.offset,
		tags:    tags,
		seed:    dr,
	}, nil
}
