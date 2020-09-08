package obj

import (
	"context"
	"io"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

const numEvictors = 10

var _ Client = &cacheClient{}

type cacheClient struct {
	slow, fast Client

	cache         *lru.Cache
	pathLocks     sync.Map
	evictionQueue chan string
	populateOnce  sync.Once
}

// NewCacheClient returns slow wrapped in an LRU write-through cache using fast for storing
// cached data.
func NewCacheClient(slow, fast Client, size int) Client {
	if size == 0 {
		return slow
	}
	client := &cacheClient{
		slow:          slow,
		fast:          fast,
		evictionQueue: make(chan string, evictionQueueSize(size)),
	}
	cache, err := lru.NewWithEvict(size, client.onEvicted)
	if err != nil {
		// lru.NewWithEvict only errors for size < 1
		panic(err)
	}
	client.cache = cache
	for i := 0; i < numEvictors; i++ {
		go client.evictor()
	}
	return client
}

func (c *cacheClient) Reader(ctx context.Context, p string, offset, size uint64) (io.ReadCloser, error) {
	c.doPopulateOnce(ctx)
	if offset > 0 || size > 0 {
		return c.slow.Reader(ctx, p, offset, size)
	}
	c.lockPath(p)
	if _, exists := c.cache.Get(p); !exists {
		if err := Copy(ctx, c.slow, c.fast, p, p); err != nil {
			c.unlockPath(p)
			return nil, err
		}
		c.cache.Add(p, struct{}{})
	}
	rc, err := c.fast.Reader(ctx, p, 0, 0)
	if err != nil {
		c.unlockPath(p)
		return nil, err
	}
	// cacheReader.Close() will unlock the path
	return &cacheReader{path: p, ReadCloser: rc, client: c}, nil
}

func (c *cacheClient) Writer(ctx context.Context, p string) (io.WriteCloser, error) {
	return c.slow.Writer(ctx, p)
}

func (c *cacheClient) Delete(ctx context.Context, p string) error {
	c.lockPath(p)
	defer c.unlockPath(p)
	if err := c.fast.Delete(ctx, p); err != nil {
		return err
	}
	if err := c.slow.Delete(ctx, p); err != nil {
		return err
	}
	return nil
}

func (c *cacheClient) Exists(ctx context.Context, p string) bool {
	c.doPopulateOnce(ctx)
	return c.fast.Exists(ctx, p) || c.slow.Exists(ctx, p)
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

func (c *cacheClient) Close() error {
	close(c.evictionQueue)
	return nil
}

func (c *cacheClient) lockPath(p string) {
	v, _ := c.pathLocks.LoadOrStore(p, &sync.Mutex{})
	v.(*sync.Mutex).Lock()
}

func (c *cacheClient) unlockPath(p string) {
	v, exists := c.pathLocks.LoadAndDelete(p)
	if !exists {
		panic("unlock path that was never locked: " + p)
	}
	v.(*sync.Mutex).Unlock()
}

func (c *cacheClient) onEvicted(key, value interface{}) {
	p := key.(string)
	select {
	case c.evictionQueue <- p:
	default:
		// this is preferable to blocking all operations, but after this we are no longer healthy.
		log.Errorf("cache eviction queue is full.  Not deleting %s", p)
	}
}

func (c *cacheClient) evictor() {
	ctx := context.Background()
	for p := range c.evictionQueue {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		c.lockPath(p)
		if err := c.fast.Delete(ctx, p); err != nil {
			log.Errorf("object cache could not delete from fast layer: %v", err)
		}
		c.unlockPath(p)
		cancel()
	}
}

func (c *cacheClient) populate(ctx context.Context) error {
	return c.fast.Walk(ctx, "", func(p string) error {
		c.cache.Add(p, struct{}{})
		return nil
	})
}

func (c *cacheClient) doPopulateOnce(ctx context.Context) {
	c.populateOnce.Do(func() {
		err := c.populate(ctx)
		log.Warnf("could not populate cache: %v", err)
	})
}

type cacheReader struct {
	io.ReadCloser
	path   string
	client *cacheClient
}

func (cr *cacheReader) Close() error {
	cr.client.unlockPath(cr.path)
	return cr.ReadCloser.Close()
}

func evictionQueueSize(cacheSize int) int {
	y := cacheSize * 1 / 10
	if y == 0 {
		y = 1
	}
	return y
}
