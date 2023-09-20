package kv

import (
	"context"
	"math"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"golang.org/x/sync/semaphore"
)

type Semaphored struct {
	inner            Store
	upload, download *semaphore.Weighted
}

func NewSemaphored(inner Store, download, upload int) *Semaphored {
	if download <= 0 {
		download = math.MaxInt
	}
	if upload <= 0 {
		upload = math.MaxInt
	}
	return &Semaphored{
		download: semaphore.NewWeighted(int64(download)),
		upload:   semaphore.NewWeighted(int64(upload)),
		inner:    inner,
	}
}

func (s *Semaphored) Put(ctx context.Context, key, value []byte) error {
	if err := s.upload.Acquire(ctx, 1); err != nil {
		return errors.EnsureStack(err)
	}
	defer s.upload.Release(1)
	return s.inner.Put(ctx, key, value)
}

func (s *Semaphored) Get(ctx context.Context, key, buf []byte) (int, error) {
	if err := s.download.Acquire(ctx, 1); err != nil {
		return 0, errors.EnsureStack(err)
	}
	defer s.download.Release(1)
	return s.inner.Get(ctx, key, buf)
}

func (s *Semaphored) Delete(ctx context.Context, key []byte) error {
	return s.inner.Delete(ctx, key)
}

func (s *Semaphored) Exists(ctx context.Context, key []byte) (bool, error) {
	return s.inner.Exists(ctx, key)
}

func (sem *Semaphored) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	return sem.inner.NewKeyIterator(span)
}
