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
	ctx         context.Context
	client      Client
	memCache    kv.GetPut
	dataRefs    []*DataRef
	offsetBytes int64
}

type ReaderOption func(*Reader)

func WithOffsetBytes(offsetBytes int64) ReaderOption {
	return func(r *Reader) {
		r.offsetBytes = offsetBytes
	}
}

func newReader(ctx context.Context, client Client, memCache kv.GetPut, dataRefs []*DataRef, opts ...ReaderOption) *Reader {
	r := &Reader{
		ctx:      ctx,
		client:   client,
		memCache: memCache,
		dataRefs: dataRefs,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Iterate iterates over the data readers for the data references.
func (r *Reader) Iterate(cb func(*DataReader) error) error {
	offset := r.offsetBytes
	for _, dataRef := range r.dataRefs {
		if dataRef.SizeBytes <= offset {
			offset -= dataRef.SizeBytes
			continue
		}
		dr := newDataReader(r.ctx, r.client, r.memCache, dataRef, offset)
		offset = 0
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
	offset   int64
}

func newDataReader(ctx context.Context, client Client, memCache kv.GetPut, dataRef *DataRef, offset int64) *DataReader {
	return &DataReader{
		ctx:      ctx,
		client:   client,
		memCache: memCache,
		dataRef:  dataRef,
		offset:   offset,
	}
}

// DataRef returns the data reference associated with this data reader.
func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

// Get writes the data referenced by the data reference.
func (dr *DataReader) Get(w io.Writer) error {
	return Get(dr.ctx, dr.client, dr.memCache, dr.dataRef.Ref, func(chunk []byte) error {
		if dr.offset > dr.dataRef.SizeBytes {
			return errors.Errorf("DataReader.offset cannot be greater than the dataRef size. offset size: %v, dataRef size: %v.", dr.offset, dr.dataRef.SizeBytes)
		}
		data := chunk[dr.dataRef.OffsetBytes+dr.offset : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
		_, err := w.Write(data)
		return err
	})
}
