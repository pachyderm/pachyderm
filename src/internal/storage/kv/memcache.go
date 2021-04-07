package kv

import (
	"context"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
)

type memoryCache struct {
	mu    sync.Mutex
	cache *simplelru.LRU
}

// NewMemCache returns a new in memory cache
func NewMemCache(size int) GetPut {
	mc := &memoryCache{}
	var err error
	mc.cache, err = simplelru.NewLRU(size, nil)
	if err != nil {
		panic(err)
	}
	return mc
}

func (mc *memoryCache) Put(ctx context.Context, key, value []byte) error {
	k := string(key)
	v := append([]byte{}, value...)
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cache.Add(k, v)
	return nil
}

func (mc *memoryCache) Get(ctx context.Context, key []byte, cb ValueCallback) error {
	v := mc.get(key)
	if v == nil {
		return ErrKeyNotFound
	}
	return cb(v)
}

func (mc *memoryCache) Delete(ctx context.Context, key []byte) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cache.Remove(key)
	return nil
}

func (mc *memoryCache) Exists(ctx context.Context, key []byte) (bool, error) {
	v := mc.get(key)
	return v == nil, nil
}

func (mc *memoryCache) get(key []byte) []byte {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	v, ok := mc.cache.Get(string(key))
	if !ok {
		return nil
	}
	return v.([]byte)
}
