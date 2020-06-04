package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
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

// PeekTag returns the next tag for the next data reader without progressing the reader.
func (r *Reader) PeekTag() (*Tag, error) {
	dr, err := r.Peek()
	if err != nil {
		return nil, err
	}
	return dr.PeekTag()
}

// Next returns the next data reader and progresses the reader.
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
// tagUpperBound is an optional parameter for specifiying the upper bound (exclusive) of the iteration.
func (r *Reader) Iterate(f func(*DataReader) error, tagUpperBound ...string) error {
	for {
		dr, err := r.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		tags, err := dr.PeekTags()
		if err != nil {
			return err
		}
		if !BeforeBound(tags[0].Id, tagUpperBound...) {
			return nil
		}
		if err := f(dr); err != nil {
			if err == errutil.ErrBreak {
				return nil
			}
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
func (r *Reader) Get(w io.Writer) error {
	return r.Iterate(func(dr *DataReader) error {
		return dr.Get(w)
	})
}

// NextTagReader sets up a tag reader for the next tagged data.
func (r *Reader) NextTagReader() *TagReader {
	return &TagReader{r: r}
}

// TagReader is an abstraction for reading tagged data. A tag
// may span multiple data readers, so this abstraction will
// connect the data readers and bound them appropriately to
// retrieve the content for a specific tag.
type TagReader struct {
	r   *Reader
	tag *Tag
}

// Iterate iterates over the data readers for the tagged data.
func (tr *TagReader) Iterate(f func(*DataReader) error) error {
	if err := tr.setup(); err != nil {
		return err
	}
	return tr.r.Iterate(func(dr *DataReader) error {
		tags, err := dr.PeekTags()
		if err != nil {
			return err
		}
		// Done if the current data reader does not have the tag.
		if tags[0].Id != tr.tag.Id {
			return errutil.ErrBreak
		}
		// Bound the current data reader if it has more than one tag.
		if len(tags) > 1 {
			dr = dr.BoundReader(tags[1].Id)
		}
		if err := f(dr); err != nil {
			return err
		}
		// Done if the current data reader has more than one tag.
		if len(tags) > 1 {
			return errutil.ErrBreak
		}
		return nil
	})
}

func (tr *TagReader) setup() error {
	if tr.tag == nil {
		var err error
		tr.tag, err = tr.r.PeekTag()
		return err
	}
	return nil
}

// Get writes the tagged data.
func (tr *TagReader) Get(w io.Writer) error {
	return tr.Iterate(func(dr *DataReader) error {
		return dr.Iterate(func(_ *Tag, r io.Reader) error {
			_, err := io.Copy(w, r)
			return err
		})
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

// DataRef returns the data reference associated with this data reader.
func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

// Len returns the length of the remaining data to be read.
func (dr *DataReader) Len() int64 {
	var size int64
	for _, tag := range dr.tags {
		size += tag.SizeBytes
	}
	return size
}

// PeekTag returns the next tag without progressing the reader.
func (dr *DataReader) PeekTag() (*Tag, error) {
	tags, err := dr.PeekTags()
	if err != nil {
		return nil, err
	}
	return tags[0], nil
}

// PeekTags returns the tags left in the reader without progressing the reader.
func (dr *DataReader) PeekTags() ([]*Tag, error) {
	if len(dr.tags) == 0 {
		return nil, io.EOF
	}
	return dr.tags, nil
}

// Iterate iterates over the tags in the data reference and passes the tag and a reader for getting
// the tagged data to the callback function.
// tagUpperBound is an optional parameter for specifiying the upper bound (exclusive) of the iteration.
func (dr *DataReader) Iterate(f func(*Tag, io.Reader) error, tagUpperBound ...string) error {
	if err := dr.getChunk(); err != nil {
		return err
	}
	for {
		tag, err := dr.PeekTag()
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
			if err == errutil.ErrBreak {
				return nil
			}
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

// Get writes the referenced data.
// This does not take into account the reading of tags.
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

// BoundReader creates a new data reader that reads a subset of the remaining tags in the current data reader.
// This tag subset is determined by the optional parameter tagUpperBound which specifies the upper bound (exclusive) of the new data reader.
// BoundReader will progress the current data reader past the tags in the tag subset.
// Data in the tag subset is fetched lazily by the new data reader.
func (dr *DataReader) BoundReader(tagUpperBound ...string) *DataReader {
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
