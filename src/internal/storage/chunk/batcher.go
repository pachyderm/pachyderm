package chunk

import (
	"context"

	"golang.org/x/sync/semaphore"

	"github.com/pachyderm/pachyderm/v2/src/internal/taskchain"
)

// ChunkFunc is a function that provides the metadata for the entries in a chunk and a data reference to the chunk.
type ChunkFunc = func([]interface{}, *DataRef) error

// EntryFunc is a function that provides the metadata for an entry and a data reference to the entry in a chunk.
// Size zero entries will have a nil data reference.
type EntryFunc = func(interface{}, *DataRef) error

// Batcher batches entries into chunks.
// Entries are buffered until they are past the configured threshold, then a chunk is created.
// Chunk creation is asynchronous with respect to the client, which is why the interface
// is callback based.
// Batcher provides one of two callback based interfaces defined by ChunkFunc and EntryFunc.
// Callbacks will be executed with respect to the order the entries are added (for the ChunkFunc
// interface, entries are ordered within as well as across calls).
type Batcher struct {
	client    Client
	entries   []*entry
	buf       []byte
	threshold int
	taskChain *taskchain.TaskChain
	chunkFunc ChunkFunc
	entryFunc EntryFunc
}

type entry struct {
	meta     interface{}
	size     int
	dataRef  *DataRef
	pointsTo []*DataRef
}

// TODO: Add config for number of entries.
func (s *Storage) NewBatcher(ctx context.Context, name string, threshold int, opts ...BatcherOption) *Batcher {
	client := NewClient(s.store, s.db, s.tracker, NewRenewer(ctx, s.tracker, name, defaultChunkTTL), s.pool)
	b := &Batcher{
		client:    client,
		threshold: threshold,
		taskChain: taskchain.New(ctx, semaphore.NewWeighted(taskParallelism)),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Batcher) Add(meta interface{}, data []byte, pointsTo []*DataRef) error {
	b.entries = append(b.entries, &entry{
		meta:     meta,
		size:     len(data),
		pointsTo: pointsTo,
	})
	b.buf = append(b.buf, data...)
	if len(b.buf) >= b.threshold {
		if err := b.createBatch(b.entries, b.buf); err != nil {
			return err
		}
		b.entries = nil
		b.buf = nil
	}
	return nil
}

func (b *Batcher) createBatch(entries []*entry, buf []byte) error {
	return b.taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
		pointsTo := getPointsTo(entries)
		dataRef, err := upload(ctx, b.client, buf, pointsTo, false)
		if err != nil {
			return nil, err
		}
		// Handle chunk callback.
		if b.chunkFunc != nil {
			var metas []interface{}
			for _, entry := range entries {
				metas = append(metas, entry.meta)
			}
			return func() error {
				return b.chunkFunc(metas, dataRef)
			}, nil
		}
		// Handle entry callback.
		if b.entryFunc != nil {
			offset := 0
			for _, entry := range entries {
				if entry.size == 0 {
					continue
				}
				entry.dataRef = NewDataRef(dataRef, buf, int64(offset), int64(entry.size))
				offset += entry.size
			}
			return func() error {
				for _, entry := range entries {
					if err := b.entryFunc(entry.meta, entry.dataRef); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}
		return nil, nil
	})
}

func getPointsTo(entries []*entry) []ID {
	var pointsTo []ID
	ids := make(map[string]struct{})
	for _, entry := range entries {
		for _, dr := range entry.pointsTo {
			id := dr.Ref.Id
			if _, exists := ids[string(id)]; !exists {
				pointsTo = append(pointsTo, id)
				ids[string(id)] = struct{}{}
			}
		}
	}
	return pointsTo
}

func (b *Batcher) Close() error {
	if len(b.entries) > 0 {
		if err := b.createBatch(b.entries, b.buf); err != nil {
			return err
		}
	}
	return b.taskChain.Wait()
}
