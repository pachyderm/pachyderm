package chunk

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx      context.Context
	client   Client
	memCache kv.GetPut
	dataRefs []*DataRef
}

func newReader(ctx context.Context, client Client, memCache kv.GetPut, dataRefs []*DataRef) *Reader {
	return &Reader{
		ctx:      ctx,
		client:   client,
		memCache: memCache,
		dataRefs: dataRefs,
	}
}

// Iterate iterates over the data readers for the data references.
func (r *Reader) Iterate(cb func(*DataReader) error) error {
	for _, dataRef := range r.dataRefs {
		dr := newDataReader(r.ctx, r.client, r.memCache, dataRef)
		if err := cb(dr); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
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
type DataReader struct {
	ctx      context.Context
	client   Client
	memCache kv.GetPut
	dataRef  *DataRef
}

func newDataReader(ctx context.Context, client Client, memCache kv.GetPut, dataRef *DataRef) *DataReader {
	return &DataReader{
		ctx:      ctx,
		client:   client,
		memCache: memCache,
		dataRef:  dataRef,
	}
}

// DataRef returns the data reference associated with this data reader.
func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

// Get writes the data referenced by the data reference.
func (dr *DataReader) Get(w io.Writer) error {
	return Get(dr.ctx, dr.client, dr.memCache, dr.dataRef.Ref, func(chunk []byte) error {
		data := chunk[dr.dataRef.OffsetBytes : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
		_, err := w.Write(data)
		return err
	})
}
