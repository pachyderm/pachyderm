package chunk

import (
	"context"
	io "io"
	"math"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"golang.org/x/sync/semaphore"
)

var _ obj.Client = limitedObjClient{}

type limitedObjClient struct {
	obj.Client
	writersSem *semaphore.Weighted
	readersSem *semaphore.Weighted
}

func newLimitedObjClient(client obj.Client, maxReaders, maxWriters int) limitedObjClient {
	if maxReaders < 1 {
		maxReaders = int(math.MaxInt64)
	}
	if maxWriters < 1 {
		maxWriters = int(math.MaxInt64)
	}
	return limitedObjClient{
		Client:     client,
		writersSem: semaphore.NewWeighted(int64(maxWriters)),
		readersSem: semaphore.NewWeighted(int64(maxReaders)),
	}
}

func (loc limitedObjClient) Writer(ctx context.Context, name string) (io.WriteCloser, error) {
	if err := loc.writersSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	w, err := loc.Client.Writer(ctx, name)
	if err != nil {
		return nil, err
	}
	return releaseWriteCloser{w, loc}, nil
}

func (loc limitedObjClient) Reader(ctx context.Context, name string, offset, size uint64) (io.ReadCloser, error) {
	if err := loc.readersSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	r, err := loc.Client.Reader(ctx, name, offset, size)
	if err != nil {
		return nil, err
	}
	return releaseReadCloser{r, loc}, nil
}

type releaseWriteCloser struct {
	io.WriteCloser
	loc limitedObjClient
}

func (rwc releaseWriteCloser) Close() error {
	rwc.loc.writersSem.Release(1)
	return rwc.WriteCloser.Close()
}

type releaseReadCloser struct {
	io.ReadCloser
	loc limitedObjClient
}

func (rrc releaseReadCloser) Close() error {
	rrc.loc.readersSem.Release(1)
	return rrc.ReadCloser.Close()
}
