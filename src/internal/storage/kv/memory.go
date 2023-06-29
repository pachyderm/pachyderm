package kv

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/google/btree"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// MemStore implements an in-memory store
type MemStore struct {
	mu   sync.RWMutex
	tree *btree.BTreeG[memEntry]
}

func NewMemStore() *MemStore {
	return &MemStore{
		// TODO: I have no idea what the degree of this tree should be.
		tree: btree.NewG(2, func(a, b memEntry) bool {
			return bytes.Compare(a.Key, b.Key) < 0
		}),
	}
}

func (ms *MemStore) Get(ctx context.Context, key []byte, buf []byte) (int, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	// this entry should not be retained, so we don't need to copy the key.
	ent, exists := ms.tree.Get(memEntry{Key: key})
	if !exists {
		return 0, pacherr.NewNotExist("kv.MemStore", string(key))
	}
	if len(ent.Value) > len(buf) {
		return 0, io.ErrShortBuffer
	}
	return copy(buf, ent.Value), nil
}

func (ms *MemStore) Put(ctx context.Context, key, value []byte) error {
	ent := newMemEntry(key, value)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.tree.ReplaceOrInsert(ent)
	return nil
}

func (ms *MemStore) Exists(ctx context.Context, key []byte) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.tree.Has(memEntry{Key: key}), nil
}

func (ms *MemStore) Delete(ctx context.Context, key []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.tree.Delete(memEntry{Key: key})
	return nil
}

func (ms *MemStore) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	panic("not implemented")
}

type memEntry struct {
	Key   []byte
	Value []byte
}

func (me memEntry) Less(other memEntry) bool {
	return bytes.Compare(me.Key, other.Key) < 0
}

func newMemEntry(key, value []byte) memEntry {
	// make a copy
	return memEntry{
		Key:   append([]byte{}, key...),
		Value: append([]byte{}, value...),
	}
}
