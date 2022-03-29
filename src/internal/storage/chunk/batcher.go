package chunk

import (
	"context"

	"golang.org/x/sync/semaphore"
)

type ChunkFunc = func([]interface{}, *DataRef) error
type EntryFunc = func(interface{}, *DataRef) error

type Batcher struct {
	client    Client
	entries   []*entry
	buf       []byte
	threshold int
	taskChain *TaskChain
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
// TODO: Size zero entries.
func (s *Storage) NewBatcher(ctx context.Context, name string, threshold int, opts ...BatcherOption) *Batcher {
	client := NewClient(s.store, s.db, s.tracker, NewRenewer(ctx, s.tracker, name, defaultChunkTTL))
	b := &Batcher{
		client:    client,
		threshold: threshold,
		taskChain: NewTaskChain(ctx, semaphore.NewWeighted(chunkParallelism)),
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
