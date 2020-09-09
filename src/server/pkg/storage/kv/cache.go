package kv

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	log "github.com/sirupsen/logrus"
)

var _ Store = &WriteThroughCache{}

type WriteThroughCache struct {
	slow, fast Store

	mu    sync.Mutex
	cache *simplelru.LRU
}

func NewWriteThroughCache(slow, fast Store, size int) *WriteThroughCache {
	c := &WriteThroughCache{
		slow: slow,
		fast: fast,
	}
	var err error
	c.cache, err = simplelru.NewLRU(size, c.onEvict)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *WriteThroughCache) GetF(ctx context.Context, p string, cb func([]byte) error) error {
	c.mu.Lock()
	if _, exists := c.cache.Get(p); exists {
		c.mu.Unlock()
		if err := c.fast.GetF(ctx, p, cb); err != nil {
			// if it was removed from the fast layer since we gave up the lock, get it from the slow layer.
			if c.fast.IsNotExist(err) {
				return c.slow.GetF(ctx, p, cb)
			}
			return err
		}
		return nil
	}
	return c.slow.GetF(ctx, p, func(data []byte) error {
		if err := c.fast.Put(ctx, p, data); err != nil {
			c.mu.Unlock()
			return err
		}
		c.cache.Add(p, struct{}{})
		c.mu.Unlock()
		return cb(data)
	})
}

func (c *WriteThroughCache) Put(ctx context.Context, key string, data []byte) error {
	if err := c.slow.Put(ctx, key, data); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.deleteFromCache(ctx, key)
}

func (c *WriteThroughCache) Exists(ctx context.Context, key string) bool {
	return c.fast.Exists(ctx, key) || c.slow.Exists(ctx, key)
}

func (c *WriteThroughCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	if err := c.deleteFromCache(ctx, key); err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()
	return c.slow.Delete(ctx, key)
}

func (c *WriteThroughCache) Walk(ctx context.Context, prefix string, cb func(key string) error) error {
	return c.slow.Walk(ctx, prefix, cb)
}

func (c *WriteThroughCache) IsNotExist(err error) bool {
	return c.fast.IsNotExist(err) || c.slow.IsNotExist(err)
}

func (c *WriteThroughCache) deleteFromCache(ctx context.Context, key string) error {
	c.cache.Remove(key)
	if err := c.fast.Delete(ctx, key); err != nil {
		log.Errorf("error deleting key (%s) from cache's fast layer: %v", key, err)
		return err
	}
	return nil
}

func (c *WriteThroughCache) onEvict(key, value interface{}) {
	// we have c.mu when this is called
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	c.deleteFromCache(ctx, key.(string))
}
