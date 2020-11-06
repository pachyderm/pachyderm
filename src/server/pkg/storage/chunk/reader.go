package chunk

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx      context.Context
	client   *Client
	dataRefs []*DataRef
}

func newReader(ctx context.Context, client *Client, dataRefs []*DataRef) *Reader {
	return &Reader{
		ctx:      ctx,
		client:   client,
		dataRefs: dataRefs,
	}
}

// Iterate iterates over the data readers for the data references.
func (r *Reader) Iterate(cb func(*DataReader) error) error {
	var seed *DataReader
	for _, dataRef := range r.dataRefs {
		dr := newDataReader(r.ctx, r.client, dataRef, seed)
		if err := cb(dr); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
		seed = dr
	}
	return nil
}

// Get writes the concatenation of the data referenced by the data references.
func (r *Reader) Get(w io.Writer) error {
	return r.Iterate(func(dr *DataReader) error {
		return dr.Get(w)
	})
}

// DataReader is an abstraction that lazily reads data referenced by a data reference.
// The seed is set to avoid re-downloading a chunk that is shared between this data reference
// and the prior in a chain of data references.
// TODO: Probably don't need seed with caching.
type DataReader struct {
	ctx        context.Context
	client     *Client
	dataRef    *DataRef
	seed       *DataReader
	getChunkMu sync.Mutex
	chunk      []byte
}

func newDataReader(ctx context.Context, client *Client, dataRef *DataRef, seed *DataReader) *DataReader {
	return &DataReader{
		ctx:     ctx,
		client:  client,
		dataRef: dataRef,
		seed:    seed,
	}
}

// DataRef returns the data reference associated with this data reader.
func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

// Get writes the data referenced by the data reference.
func (dr *DataReader) Get(w io.Writer) error {
	if err := dr.getChunk(); err != nil {
		return err
	}
	data := dr.chunk[dr.dataRef.OffsetBytes : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
	_, err := w.Write(data)
	return err
}

func (dr *DataReader) getChunk() error {
	dr.getChunkMu.Lock()
	defer dr.getChunkMu.Unlock()
	if dr.chunk != nil {
		return nil
	}
	// Use seed chunk if possible.
	if dr.seed != nil && bytes.Equal(dr.dataRef.Ref.Id, dr.seed.dataRef.Ref.Id) {
		if err := dr.seed.getChunk(); err != nil {
			return err
		}
		dr.chunk = dr.seed.chunk
		return nil
	}
	// Get chunk from object storage.
	buf := &bytes.Buffer{}
	chunkID := dr.dataRef.Ref.Id
	if err := dr.client.Get(dr.ctx, chunkID, buf); err != nil {
		return err
	}
	dr.chunk = buf.Bytes()
	return nil
}
