package obj

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
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

func (c *cacheClient) Get(ctx context.Context, p string, w io.Writer) error {
	c.doPopulateOnce(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.cache.Get(p); exists {
		cacheHitMetric.Inc()
		return c.fast.Get(ctx, p, w)
	}
	cacheMissMetric.Inc()
	if err := Copy(ctx, c.slow, c.fast, p, p); err != nil {
		return err
	}
	c.cache.Add(p, struct{}{})
	return c.fast.Get(ctx, p, w)
}

func (c *cacheClient) Put(ctx context.Context, p string, r io.Reader) error {
	return c.slow.Put(ctx, p, r)
}

func (c *cacheClient) Delete(ctx context.Context, p string) error {
	if err := c.slow.Delete(ctx, p); err != nil {
		return err
	}
	return c.deleteFromCache(ctx, p)
}

func (c *cacheClient) Exists(ctx context.Context, p string) (bool, error) {
	return c.slow.Exists(ctx, p)
}

func (c *cacheClient) Walk(ctx context.Context, p string, cb func(p string) error) error {
	return c.slow.Walk(ctx, p, cb)
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
	if err := c.fast.Delete(ctx, p); err != nil && !pacherr.IsNotExist(err) {
		log.Errorf("could not delete from cache's fast store: %v", err)
	}
	cacheEvictionMetric.Inc()
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
