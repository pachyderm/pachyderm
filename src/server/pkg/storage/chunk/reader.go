package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx      context.Context
	objC     obj.Client
	dataRefs []*DataRef
	peek     *DataReader
	prev     *DataReader
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

// Peek returns the next data reader without progressing the reader.
func (r *Reader) Peek() (*DataReader, error) {
	if r.peek == nil {
		var err error
		r.peek, err = r.Next()
		if err != nil {
			return nil, err
		}
	}
	return r.peek, nil
}

// Next progresses the reader to the next data reader.
func (r *Reader) Next() (*DataReader, error) {
	if r.peek != nil {
		dr := r.peek
		r.peek = nil
		return dr, nil
	}
	if len(r.dataRefs) == 0 {
		return nil, io.EOF
	}
	dr := newDataReader(r.ctx, r.objC, r.dataRefs[0], r.prev)
	r.dataRefs = r.dataRefs[1:]
	r.prev = dr
	return dr, nil
}

// Iterate iterates over the data readers for the current data references
// set in the reader.
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
		if _, err := r.Next(); err != nil {
			return err
		}
	}
	return nil
}

// Get writes the concatenation of the data represented by the data references
// set in the reader.
// (bryce) probably should make a decision on whether this should be blocked for
// a reader that already has been partially iterated.
func (r *Reader) Get(w io.Writer) error {
	return r.Iterate(func(dr *DataReader) error {
		return dr.Get(w)
	})
}

// DataReader is an abstraction that lazily reads data referenced by a data reference.
// The seed is set to avoid re-downloading a chunk that is shared between this data reference
// and the prior in a chain of data references.
type DataReader struct {
	ctx        context.Context
	objC       obj.Client
	dataRef    *DataRef
	getChunkMu sync.Mutex
	chunk      []byte
	offset     int64
	tags       []*Tag
	seed       *DataReader
}

func newDataReader(ctx context.Context, objC obj.Client, dataRef *DataRef, seed *DataReader) *DataReader {
	return &DataReader{
		ctx:     ctx,
		objC:    objC,
		dataRef: dataRef,
		offset:  dataRef.OffsetBytes,
		tags:    dataRef.Tags,
		seed:    seed,
	}
}

// DataRef is the data reference associated with this data reader.
func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

// Len is the length of the remaining data to be read.
func (dr *DataReader) Len() int64 {
	var size int64
	for _, tag := range dr.tags {
		size += tag.SizeBytes
	}
	return size
}

// Peek peeks ahead in the tags.
func (dr *DataReader) Peek() (*Tag, error) {
	if len(dr.tags) == 0 {
		return nil, io.EOF
	}
	return dr.tags[0], nil
}

// Iterate iterates over the tags in the data reference and passes the tag and a reader for getting
// the tagged content to the callback function.
// tagUpperBound is an optional parameter for specifiying the upper bound (exclusive) of the iteration.
func (dr *DataReader) Iterate(f func(*Tag, io.Reader) error, tagUpperBound ...string) error {
	if err := dr.getChunk(); err != nil {
		return err
	}
	for {
		tag, err := dr.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !BeforeBound(tag.Id, tagUpperBound...) {
			return nil
		}
		if err := f(tag, bytes.NewReader(dr.chunk[dr.offset:dr.offset+tag.SizeBytes])); err != nil {
			return err
		}
		dr.tags = dr.tags[1:]
		dr.offset += tag.SizeBytes
	}
}

func (dr *DataReader) getChunk() error {
	dr.getChunkMu.Lock()
	defer dr.getChunkMu.Unlock()
	if dr.chunk != nil {
		return nil
	}
	// Use seed chunk if possible.
	if dr.seed != nil && dr.dataRef.ChunkInfo.Chunk.Hash == dr.seed.dataRef.ChunkInfo.Chunk.Hash {
		if err := dr.seed.getChunk(); err != nil {
			return err
		}
		dr.chunk = dr.seed.chunk
		return nil
	}
	// Get chunk from object storage.
	objR, err := dr.objC.Reader(dr.ctx, path.Join(prefix, dr.dataRef.ChunkInfo.Chunk.Hash), 0, 0)
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
	return len(strBound) == 0 || str < strBound[0]
}

// Get writes the data referenced by the data reference.
func (dr *DataReader) Get(w io.Writer) error {
	if err := dr.getChunk(); err != nil {
		return err
	}
	data := dr.chunk[dr.dataRef.OffsetBytes : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

// LimitReader creates a new data reader that reads a subset of the remaining tags in the current data reader.
// This tag subset is determined by the optional parameter tagUpperBound which specifies the upper bound (exclusive) of the new data reader.
// LimitReader will progress the current data reader past the tags in the tag subset.
// Data in the tag subset is fetched lazily by the new data reader.
func (dr *DataReader) LimitReader(tagUpperBound ...string) *DataReader {
	offset := dr.offset
	var tags []*Tag
	for {
		if len(dr.tags) == 0 {
			break
		}
		tag := dr.tags[0]
		if !BeforeBound(tag.Id, tagUpperBound...) {
			break
		}
		tags = append(tags, tag)
		dr.offset += tag.SizeBytes
		dr.tags = dr.tags[1:]
	}
	return &DataReader{
		ctx:     dr.ctx,
		objC:    dr.objC,
		dataRef: dr.dataRef,
		offset:  offset,
		tags:    tags,
		seed:    dr,
	}
}
