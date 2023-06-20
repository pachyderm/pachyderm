package chunk

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/taskchain"

	"golang.org/x/sync/semaphore"
)

// Reader reads data from chunk storage.
type Reader struct {
	ctx           context.Context
	storage       *Storage
	client        Client
	dataRefs      []*DataRef
	offsetBytes   int64
	prefetchLimit int
}

func newReader(ctx context.Context, s *Storage, client Client, dataRefs []*DataRef, opts ...ReaderOption) *Reader {
	r := &Reader{
		ctx:      ctx,
		client:   client,
		storage:  s,
		dataRefs: dataRefs,
	}
	for _, opt := range opts {
		opt(r)
	}
	for len(r.dataRefs) > 0 {
		if r.dataRefs[0].SizeBytes > r.offsetBytes {
			break
		}
		r.offsetBytes -= r.dataRefs[0].SizeBytes
		r.dataRefs = r.dataRefs[1:]
	}
	return r
}

// Get writes the concatenation of the data referenced by the data references.
func (r *Reader) Get(w io.Writer) (retErr error) {
	if len(r.dataRefs) == 0 {
		return nil
	}
	if len(r.dataRefs) == 1 {
		_, err := io.Copy(w, newDataReader(r.ctx, r.storage, r.client, r.dataRefs[0], r.offsetBytes))
		return errors.EnsureStack(err)
	}
	ctx, cancel := pctx.WithCancel(r.ctx)
	defer cancel()
	taskChain := taskchain.New(ctx, semaphore.NewWeighted(int64(r.prefetchLimit)))
	defer func() {
		if retErr != nil {
			cancel()
		}
		if err := taskChain.Wait(); retErr == nil {
			retErr = err
		}
	}()
	for i, dataRef := range r.dataRefs {
		var offset int64
		if i == 0 {
			offset = r.offsetBytes
		}
		dr := newDataReader(r.ctx, r.storage, r.client, dataRef, offset)
		if err := taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
			if err := dr.fetchData(); err != nil {
				return nil, err
			}
			return func() error {
				_, err := io.Copy(w, dr)
				return errors.EnsureStack(err)
			}, nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// DataReader is an abstraction that lazily reads data referenced by a data reference.
type DataReader struct {
	ctx      context.Context
	client   Client
	memCache *memoryCache
	pool     *kv.Pool
	deduper  *miscutil.WorkDeduper[pachhash.Output]
	dataRef  *DataRef
	offset   int64
	r        io.Reader
}

func newDataReader(ctx context.Context, s *Storage, client Client, dataRef *DataRef, offset int64) *DataReader {
	return &DataReader{
		ctx:      ctx,
		client:   client,
		memCache: s.memCache,
		pool:     s.pool,
		deduper:  s.deduper,
		dataRef:  dataRef,
		offset:   offset,
	}
}

func (dr *DataReader) Read(data []byte) (int, error) {
	if err := dr.fetchData(); err != nil {
		return 0, err
	}
	n, err := dr.r.Read(data)
	return n, errors.EnsureStack(err)
}

func (dr *DataReader) fetchData() error {
	if dr.r != nil {
		return nil
	}
	ref := dr.dataRef.Ref
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Millisecond
	var data []byte
	if err := backoff.RetryUntilCancel(dr.ctx, func() error {
		chunkData, err := getFromCache(dr.memCache, ref)
		if err != nil {
			return err
		}
		data = chunkData[dr.dataRef.OffsetBytes+dr.offset : dr.dataRef.OffsetBytes+dr.dataRef.SizeBytes]
		return nil
	}, b, func(err error, _ time.Duration) error {
		if !pacherr.IsNotExist(err) {
			return err
		}
		return dr.deduper.Do(dr.ctx, ref.Key(), func() error {
			data, err := Get(dr.ctx, dr.client, ref)
			if err != nil {
				return err
			}
			putInCache(dr.memCache, ref, data)
			return nil
		})
	}); err != nil {
		return err
	}
	dr.r = bytes.NewReader(data)
	return nil
}
