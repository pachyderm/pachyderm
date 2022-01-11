package chunk

import (
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx           context.Context
	client        Client
	memCache      kv.GetPut
	deduper       *miscutil.WorkDeduper
	dataRefs      []*DataRef
	offsetBytes   int64
	prefetchLimit int
}

type ReaderOption func(*Reader)

func WithOffsetBytes(offsetBytes int64) ReaderOption {
	return func(r *Reader) {
		r.offsetBytes = offsetBytes
	}
}

func newReader(ctx context.Context, client Client, memCache kv.GetPut, deduper *miscutil.WorkDeduper, prefetchLimit int, dataRefs []*DataRef, opts ...ReaderOption) *Reader {
	r := &Reader{
		ctx:           ctx,
		client:        client,
		memCache:      memCache,
		deduper:       deduper,
		prefetchLimit: prefetchLimit,
		dataRefs:      dataRefs,
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
		dr := newDataReader(r.ctx, r.client, r.memCache, r.deduper, dataRef, offset)
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
func (r *Reader) Get(w io.Writer) (retErr error) {
	if len(r.dataRefs) <= 1 {
		return r.Iterate(func(dr *DataReader) error {
			return dr.Get(w)
		})
	}
	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()
	taskChain := NewTaskChain(ctx, int64(r.prefetchLimit))
	defer func() {
		if retErr != nil {
			cancel()
		}
		if err := taskChain.Wait(); retErr == nil {
			retErr = err
		}
	}()
	return r.Iterate(func(dr *DataReader) error {
		return taskChain.CreateTask(func(ctx context.Context, serial func(func() error) error) error {
			if err := dr.Prefetch(); err != nil {
				return err
			}
			return serial(func() error {
				return dr.Get(w)
			})
		})
	})
}

// DataReader is an abstraction that lazily reads data referenced by a data reference.
type DataReader struct {
	ctx      context.Context
	client   Client
	memCache kv.GetPut
	deduper  *miscutil.WorkDeduper
	dataRef  *DataRef
	offset   int64
}

func newDataReader(ctx context.Context, client Client, memCache kv.GetPut, deduper *miscutil.WorkDeduper, dataRef *DataRef, offset int64) *DataReader {
	return &DataReader{
		ctx:      ctx,
		client:   client,
		memCache: memCache,
		deduper:  deduper,
		dataRef:  dataRef,
		offset:   offset,
	}
}

// DataRef returns the data reference associated with this data reader.
func (dr *DataReader) DataRef() *DataRef {
	return dr.dataRef
}

func (dr *DataReader) Prefetch() error {
	return dr.Get(io.Discard)
}

// Get writes the data referenced by the data reference.
func (dr *DataReader) Get(w io.Writer) error {
	if dr.offset > dr.dataRef.SizeBytes {
		return errors.Errorf("DataReader.offset cannot be greater than the dataRef size. offset size: %v, dataRef size: %v.", dr.offset, dr.dataRef.SizeBytes)
	}
	ref := dr.dataRef.Ref
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Millisecond
	return backoff.RetryUntilCancel(dr.ctx, func() error {
		return getFromCache(dr.ctx, dr.memCache, ref, func(chunk []byte) error {
			data := chunk[dr.dataRef.OffsetBytes+dr.offset : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
			_, err := w.Write(data)
			return errors.EnsureStack(err)
		})
	}, b, func(err error, _ time.Duration) error {
		if !pacherr.IsNotExist(err) {
			return err
		}
		return dr.deduper.Do(dr.ctx, ref.Key(), func() error {
			return Get(dr.ctx, dr.client, ref, func(rawData []byte) error {
				return putInCache(dr.ctx, dr.memCache, ref, rawData)
			})
		})
	})
}

func getFromCache(ctx context.Context, cache kv.GetPut, ref *Ref, cb kv.ValueCallback) error {
	key := ref.Key()
	return errors.EnsureStack(cache.Get(ctx, key[:], cb))
}

func putInCache(ctx context.Context, cache kv.GetPut, ref *Ref, data []byte) error {
	key := ref.Key()
	return errors.EnsureStack(cache.Put(ctx, key[:], data))
}
