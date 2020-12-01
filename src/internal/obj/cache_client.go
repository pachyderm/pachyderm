package obj

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	log "github.com/sirupsen/logrus"
)

var _ Client = &cacheClient{}

type cacheClient struct {
	slow, fast Client

	mu           sync.Mutex
	cache        *simplelru.LRU
	populateOnce sync.Once
}

// NewCacheClient returns slow wrapped in an LRU write-through cache using fast for storing
// cached data.
func NewCacheClient(slow, fast Client, size int) Client {
	if size == 0 {
		return slow
	}
	client := &cacheClient{
		slow: slow,
		fast: fast,
	}
	cache, err := simplelru.NewLRU(size, client.onEvicted)
	if err != nil {
		// lru.NewWithEvict only errors for size < 1
		panic(err)
	}
	client.cache = cache
	return client
}

func (c *cacheClient) Reader(ctx context.Context, p string, offset, size uint64) (io.ReadCloser, error) {
	c.doPopulateOnce(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.cache.Get(p); exists {
		return c.fast.Reader(ctx, p, offset, size)
	}
	if err := Copy(ctx, c.slow, c.fast, p, p); err != nil {
		return nil, err
	}
	c.cache.Add(p, struct{}{})
	return c.fast.Reader(ctx, p, offset, size)
}

func (c *cacheClient) Writer(ctx context.Context, p string) (io.WriteCloser, error) {
	return c.slow.Writer(ctx, p)
}

func (c *cacheClient) Delete(ctx context.Context, p string) error {
	if err := c.slow.Delete(ctx, p); err != nil {
		return err
	}
	return c.deleteFromCache(ctx, p)
}

func (c *cacheClient) Exists(ctx context.Context, p string) bool {
	return c.slow.Exists(ctx, p)
}

func (c *cacheClient) Walk(ctx context.Context, p string, cb func(p string) error) error {
	return c.slow.Walk(ctx, p, cb)
}

func (c *cacheClient) IsIgnorable(err error) bool {
	return c.fast.IsIgnorable(err) || c.slow.IsIgnorable(err)
}

func (c *cacheClient) IsNotExist(err error) bool {
	return c.fast.IsNotExist(err) || c.slow.IsNotExist(err)
}

func (c *cacheClient) IsRetryable(err error) bool {
	return c.fast.IsRetryable(err) || c.slow.IsRetryable(err)
}

func (c *cacheClient) deleteFromCache(ctx context.Context, p string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Remove(p)
	return nil
}

// onEvicted is called by the cache implementation.
// the gorountine executing onEvicted will have c.mu
func (c *cacheClient) onEvicted(key, value interface{}) {
	p := key.(string)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.fast.Delete(ctx, p); err != nil && !c.fast.IsNotExist(err) {
		log.Errorf("could not delete from cache's fast store: %v", err)
	}
}

func (c *cacheClient) populate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fast.Walk(ctx, "", func(p string) error {
		c.cache.Add(p, struct{}{})
		return nil
	})
}

func (c *cacheClient) doPopulateOnce(ctx context.Context) {
	c.populateOnce.Do(func() {
		if err := c.populate(ctx); err != nil {
			log.Warnf("could not populate cache: %v", err)
		}
	})
}

type voidClient struct{}

// NewVoid create a client which retains no objects. It can be used with the cache client
func NewVoid() Client {
	return &memClient{}
}

func (c *voidClient) Writer(ctx context.Context, p string) (io.WriteCloser, error) {
	return voidWriter{}, nil
}

func (c *voidClient) Reader(ctx context.Context, p string, offset, size uint64) (io.ReadCloser, error) {
	return nil, os.ErrNotExist
}

func (c *voidClient) Exists(ctx context.Context, p string) bool {
	return false
}

func (c *voidClient) Delete(ctx context.Context, p string) error {
	return nil
}

func (c *voidClient) IsIgnorable(error) bool {
	return false
}

func (c *voidClient) IsNotExist(err error) bool {
	return true
}

func (c *voidClient) IsRetryable(error) bool {
	return false
}

func (c *voidClient) Walk(ctx context.Context, prefix string, cb func(p string) error) error {
	return nil
}

type voidWriter struct{}

func (voidWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
func (voidWriter) Close() error {
	return nil
}
