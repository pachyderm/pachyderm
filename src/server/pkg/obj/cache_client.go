package obj

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	log "github.com/sirupsen/logrus"
)

var _ Client = &cacheClient{}

type cacheClient struct {
	slow, fast Client

	mu    sync.Mutex
	cache *simplelru.LRU
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
	c.mu.Lock()
	defer c.mu.Unlock()
	cacheP := cachePath(p, offset, size)
	if _, exists := c.cache.Get(cacheP); exists {
		return c.fast.Reader(ctx, cacheP, 0, 0)
	}
	if err := copyToCache(ctx, c.slow, c.fast, p, offset, size); err != nil {
		return nil, err
	}
	c.cache.Add(cacheP, struct{}{})
	return c.fast.Reader(ctx, cacheP, 0, 0)
}

func cachePath(p string, offset, size uint64) string {
	return fmt.Sprintf("%v-%v-%v", p, offset, size)
}

func copyToCache(ctx context.Context, src, dst Client, p string, offset, size uint64) (retErr error) {
	rc, err := src.Reader(ctx, p, offset, size)
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); retErr == nil {
			retErr = err
		}
	}()
	wc, err := dst.Writer(ctx, cachePath(p, offset, size))
	if err != nil {
		return err
	}
	defer func() {
		if err := wc.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(wc, rc)
	return err
}

func (c *cacheClient) Writer(ctx context.Context, p string) (io.WriteCloser, error) {
	return c.slow.Writer(ctx, p)
}

func (c *cacheClient) Delete(ctx context.Context, p string) error {
	return c.slow.Delete(ctx, p)
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
