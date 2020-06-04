package obj

import (
	"context"
	io "io"
	"math"

	"golang.org/x/sync/semaphore"
)

var _ Client = &limitedClient{}

type limitedClient struct {
	Client
	writersSem *semaphore.Weighted
	readersSem *semaphore.Weighted
}

func NewLimitedClient(client Client, maxReaders, maxWriters int) *limitedClient {
	if maxReaders < 1 {
		maxReaders = int(math.MaxInt64)
	}
	if maxWriters < 1 {
		maxWriters = int(math.MaxInt64)
	}
	return &limitedClient{
		Client:     client,
		writersSem: semaphore.NewWeighted(int64(maxWriters)),
		readersSem: semaphore.NewWeighted(int64(maxReaders)),
	}
}

func (loc *limitedClient) Writer(ctx context.Context, name string) (io.WriteCloser, error) {
	if err := loc.writersSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	w, err := loc.Client.Writer(ctx, name)
	if err != nil {
		return nil, err
	}
	return releaseWriteCloser{w, loc.writersSem}, nil
}

func (loc *limitedClient) Reader(ctx context.Context, name string, offset, size uint64) (io.ReadCloser, error) {
	if err := loc.readersSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	r, err := loc.Client.Reader(ctx, name, offset, size)
	if err != nil {
		return nil, err
	}
	return releaseReadCloser{r, loc.readersSem}, nil
}

type releaseWriteCloser struct {
	io.WriteCloser
	sem *semaphore.Weighted
}

func (rwc releaseWriteCloser) Close() error {
	rwc.sem.Release(1)
	return rwc.WriteCloser.Close()
}

type releaseReadCloser struct {
	io.ReadCloser
	sem *semaphore.Weighted
}

func (rrc releaseReadCloser) Close() error {
	rrc.sem.Release(1)
	return rrc.ReadCloser.Close()
}
