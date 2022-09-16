package fileset

import (
	"bytes"
	"context"

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
// the chunks.
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
		// Handle first chunk.
		if ref == nil {
			ref = chunk.FullRef(dataRefs[0])
		}
		// Handle files that are fully contained within one chunk.
		// Buffer the file if it is in the same chunk as the current chunk.
		// Otherwise, fetch the current chunk, then set up the new current chunk and buffered files.
		if len(dataRefs) == 1 {
			if bytes.Equal(dataRefs[0].Ref.Id, ref.Ref.Id) {
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
		// Fetch the curent chunk first.
		if err := fetchChunk(ref, files); err != nil {
			return err
		}
		// Fetch the first chunk if it is different from the current chunk.
		// It will be fetched by the above fetchChunk if it is the same.
		if !bytes.Equal(dataRefs[0].Ref.Id, ref.Ref.Id) {
			if err := fetchChunk(chunk.FullRef(dataRefs[0]), nil); err != nil {
				return err
			}
		}
		// Fetch the full chunks that the file spans.
		for i := 1; i < len(dataRefs)-1; i++ {
			if err := fetchChunk(chunk.FullRef(dataRefs[i]), nil); err != nil {
				return err
			}
		}
		// Set the current chunk to the last chunk and buffer the file.
		// The file will be emitted when the last chunk has been fetched.
		ref = chunk.FullRef(dataRefs[len(dataRefs)-1])
		files = []File{f}
		return nil
	}, opts...); err != nil {
		return err
	}
	// Emit the last buffered files.
	if err := fetchChunk(ref, files); err != nil {
		return err
	}
	return taskChain.Wait()
}
