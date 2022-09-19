package fileset

import (
	"bytes"
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"golang.org/x/sync/semaphore"
)

// prefetcher prefetches chunks for the files emitted by the wrapped file set.
// The prefetching applies at chunk boundaries rather than file boundaries, so
// it handles the prefetching of chunks that contain the content of many small
// files as well as chunks that contain the content of a larger file. This is
// handled by buffering file metadata up to chunk boundaries, then kicking off
// concurrent tasks to prefetch the chunks and emit the files associated with
// the chunks. The prefetching of files with more chunks than the prefetch
// limit will not be handled by this abstraction. In this case, the prefetching
// will be handled when the content of the file is needed.
type prefetcher struct {
	FileSet
	storage *Storage
}

func NewPrefetcher(storage *Storage, fileSet FileSet) FileSet {
	return &prefetcher{
		FileSet: fileSet,
		storage: storage,
	}
}

func (p *prefetcher) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	taskChain := chunk.NewTaskChain(ctx, semaphore.NewWeighted(int64(p.storage.prefetchLimit)))
	fetchChunk := func(ref *chunk.DataRef, files []File) error {
		return taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
			if ref != nil {
				if err := p.storage.ChunkStorage().PrefetchData(ctx, ref); err != nil {
					return nil, err
				}
			}
			return func() error {
				for _, f := range files {
					if err := cb(f); err != nil {
						return err
					}
				}
				return nil
			}, nil
		})
	}
	waitFile := func(f File) error {
		done := make(chan struct{})
		if err := taskChain.CreateTask(func(_ context.Context) (func() error, error) {
			return func() error {
				defer close(done)
				return cb(f)
			}, nil
		}); err != nil {
			return err
		}
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return errors.EnsureStack(ctx.Err())
		}
	}
	var ref *chunk.DataRef
	var files []File
	maxFilesBuf := p.storage.shardCountThreshold / int64(p.storage.prefetchLimit)
	if err := p.FileSet.Iterate(ctx, func(f File) error {
		// Emit the files if a large number are buffered.
		if int64(len(files)) >= maxFilesBuf {
			if err := fetchChunk(ref, files); err != nil {
				return err
			}
			ref = nil
			files = nil
		}
		dataRefs := f.Index().File.DataRefs
		// Handle empty files.
		// It does not matter which chunk they get emitted with,
		// just that they are emitted in the correct order.
		if len(dataRefs) == 0 {
			files = append(files, f)
			return nil
		}
		// Handle files that are fully contained within one chunk.
		// Buffer the file if it is in the same chunk as the current chunk.
		// Otherwise, fetch the current chunk, then set up the new current chunk and buffered files.
		if len(dataRefs) == 1 {
			if ref != nil && bytes.Equal(dataRefs[0].Ref.Id, ref.Ref.Id) {
				files = append(files, f)
				return nil
			}
			if err := fetchChunk(ref, files); err != nil {
				return err
			}
			ref = chunk.FullRef(dataRefs[0])
			files = []File{f}
			return nil
		}
		// Handle files that span multiple chunks.
		// Fetch the current chunk first.
		if err := fetchChunk(ref, files); err != nil {
			return err
		}
		ref = nil
		files = nil
		// Handle files that span more chunks than the prefetch limit.
		// Don't attempt to prefetch these files and wait for the client
		// to be done with them to prevent thrashing the cache.
		if len(dataRefs) > p.storage.prefetchLimit {
			return waitFile(f)
		}
		// Fetch the full chunks that the file spans.
		for i := 0; i < len(dataRefs)-1; i++ {
			if err := fetchChunk(chunk.FullRef(dataRefs[i]), nil); err != nil {
				return err
			}
		}
		// Emit the file when the last chunk has been fetched.
		return fetchChunk(chunk.FullRef(dataRefs[len(dataRefs)-1]), []File{f})
	}, opts...); err != nil {
		return err
	}
	// Emit the last buffered files.
	if err := fetchChunk(ref, files); err != nil {
		return err
	}
	return taskChain.Wait()
}
