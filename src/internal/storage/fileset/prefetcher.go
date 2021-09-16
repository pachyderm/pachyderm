package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func (s *Storage) NewPrefetcher(fs FileSet) FileSet {
	return newPrefetcher(s.chunks, fs, s.prefetchLimit)
}

var _ FileSet = &prefetcher{}

// TODO: Refactor the limiting into the task chain abstraction.
type prefetcher struct {
	chunks *chunk.Storage
	fs     FileSet
	limit  int
}

func newPrefetcher(chunks *chunk.Storage, fs FileSet, limit int) FileSet {
	return &prefetcher{
		chunks: chunks,
		fs:     fs,
		limit:  limit,
	}
}

// TODO: Consider deletive.
func (p *prefetcher) Iterate(ctx context.Context, cb func(File) error, _ ...bool) error {
	fileChan := make(chan []File, p.limit)
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	eg.Go(func() error {
		for {
			select {
			case files, more := <-fileChan:
				if !more {
					return nil
				}
				for _, f := range files {
					if err := cb(f); err != nil {
						return err
					}
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	eg.Go(func() error {
		defer close(fileChan)
		return p.iterateFiles(ctx, fileChan)
	})
	return eg.Wait()
}

func (p *prefetcher) iterateFiles(ctx context.Context, fileChan chan []File) (retErr error) {
	sem := semaphore.NewWeighted(int64(p.limit))
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)
	defer func() {
		if retErr != nil {
			cancel()
		}
		if err := eg.Wait(); retErr == nil {
			retErr = err
		}
	}()
	var files []File
	var prevChunk string
	if err := p.fs.Iterate(ctx, func(f File) error {
		idx := f.Index()
		files = append(files, f)
		if idx.File.DataRefs == nil {
			select {
			case fileChan <- files:
			case <-ctx.Done():
				return ctx.Err()
			}
			files = nil
			return nil
		}
		// Buffer files that only reference the previous chunk.
		firstChunk := chunk.ID(idx.File.DataRefs[0].Ref.Id).String()
		if len(idx.File.DataRefs) == 1 && firstChunk == prevChunk {
			return nil
		}
		prevChunk = firstChunk
		// Set up goroutines for fetching the chunks referenced by the file.
		for _, dataRef := range idx.File.DataRefs {
			ref := dataRef.Ref
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			eg.Go(func() error {
				defer sem.Release(1)
				return p.chunks.Prefetch(ctx, ref)
			})
		}
		select {
		case fileChan <- files:
		case <-ctx.Done():
			return ctx.Err()
		}
		files = nil
		return nil
	}); err != nil {
		return err
	}
	select {
	case fileChan <- files:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
