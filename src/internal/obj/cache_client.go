package obj

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	cacheHitMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage_cache",
		Name:      "hits_total",
		Help:      "Number of object storage gets served from cache",
	})
	cacheMissMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage_cache",
		Name:      "misses_total",
		Help:      "Number of object storage gets that were not served from cache",
	})
	cacheEvictionMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage_cache",
		Name:      "evictions_total",
		Help:      "Number of objects evicted from LRU cache",
	})
)

var _ Client = &cacheClient{}

type cacheClient struct {
	slow, fast   Client
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

func (c *cacheClient) Get(ctx context.Context, p string, w io.Writer) error {
	c.doPopulateOnce(ctx) // always call before acquiring locks
	if hit, err := c.getFast(ctx, p, w); err != nil {
		return err
	} else if hit {
		cacheHitMetric.Inc()
		return nil
	}
	cacheMissMetric.Inc()
	return c.getSlow(ctx, p, w)
}

// getFast retrieves an object from c.fast if it exists, writes it to w and returns true.
// if it does not exist nothing is written to w and false is returned.
func (c *cacheClient) getFast(ctx context.Context, p string, w io.Writer) (bool, error) {
	c.mu.Lock()
	_, exists := c.cache.Get(p)
	c.mu.Unlock()
	if !exists {
		return false, nil
	}
	if err := c.fast.Get(ctx, p, w); err != nil && pacherr.IsNotExist(err) {
		// could have been deleted since we released the lock, but that's fine.
		return false, nil
	} else if err != nil {
		return false, errors.EnsureStack(err)
	}
	return true, nil
}

// getSlow retrieves an object from c.slow and concurrently writes it to w and to c.fast.
// c.mu is not acquired until after a succesful put to prevent holding the lock while doing IO,
// and to prevent a slow w from holding up cache access for other callers.
func (c *cacheClient) getSlow(ctx context.Context, p string, w io.Writer) error {
	return miscutil.WithPipe(func(w2 io.Writer) error {
		mw := io.MultiWriter(w, w2)
		err := c.slow.Get(ctx, p, mw)
		return errors.EnsureStack(err)
	}, func(r io.Reader) error {
		if err := c.fast.Put(ctx, p, r); pacherr.IsNotExist(err) {
			// if the object could not be found, the first read will return a NotExist error
			return errors.EnsureStack(err)
		} else if err != nil {
			log.Error(ctx, "obj.cacheClient: error writing to cache", zap.Error(err))
			return nil
		}
		c.mu.Lock()
		c.cache.Add(p, struct{}{})
		c.mu.Unlock()
		return nil
	})
}

func (c *cacheClient) Put(ctx context.Context, p string, r io.Reader) error {
	return errors.EnsureStack(c.slow.Put(ctx, p, r))
}

func (c *cacheClient) Delete(ctx context.Context, p string) error {
	if err := c.slow.Delete(ctx, p); err != nil {
		return errors.EnsureStack(err)
	}
	return c.deleteFromCache(ctx, p)
}

func (c *cacheClient) Exists(ctx context.Context, p string) (bool, error) {
	res, err := c.slow.Exists(ctx, p)
	return res, errors.EnsureStack(err)
}

func (c *cacheClient) Walk(ctx context.Context, p string, cb func(p string) error) error {
	return errors.EnsureStack(c.slow.Walk(ctx, p, cb))
}

func (c *cacheClient) BucketURL() ObjectStoreURL {
	return c.slow.BucketURL()
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.fast.Delete(ctx, p); err != nil && !pacherr.IsNotExist(err) {
		log.Error(ctx, "could not delete from cache's fast store", zap.Error(err))
	}
	cacheEvictionMetric.Inc()
}

func (c *cacheClient) populate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return errors.EnsureStack(c.fast.Walk(ctx, "", func(p string) error {
		c.cache.Add(p, struct{}{})
		return nil
	}))
}

func (c *cacheClient) doPopulateOnce(ctx context.Context) {
	c.populateOnce.Do(func() {
		if err := c.populate(ctx); err != nil {
			log.Error(ctx, "could not populate cache", zap.Error(err))
		}
	})
}
