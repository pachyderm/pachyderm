package obj

import (
	"context"
	"io"
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
	pathLocks    sync.Map
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
	if offset > 0 || size > 0 {
		return c.slow.Reader(ctx, p, offset, size)
	}
	c.mu.Lock()
	if _, exists := c.cache.Get(p); !exists {
		c.lockPath(p)
		c.mu.Unlock()
		if err := Copy(ctx, c.slow, c.fast, p, p); err != nil {
			c.unlockPath(p)
			log.Errorf("untracked object (%s) in cache fast layer. error: %v", p, err)
			return nil, err
		}
		c.mu.Lock()
		c.cache.Add(p, struct{}{})
		c.unlockPath(p)
		c.mu.Unlock()
	} else {
		c.mu.Unlock()
	}

	c.rlockPath(p)
	rc, err := c.fast.Reader(ctx, p, 0, 0)
	if err != nil {
		c.runlockPath(p)
		// at this point it should be in the fast cache. If it is not, return a transient error
		if c.fast.IsNotExist(err) {
			return nil, ErrCacheRace{}
		}
		return nil, err
	}
	// cacheReader.Close() will unlock the path
	return &cacheReader{path: p, ReadCloser: rc, client: c}, nil
}

func (c *cacheClient) Writer(ctx context.Context, p string) (io.WriteCloser, error) {
	return c.slow.Writer(ctx, p)
}

func (c *cacheClient) Delete(ctx context.Context, p string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lockPath(p)
	defer c.unlockPath(p)
	c.cache.Remove(p)
	if err := c.fast.Delete(ctx, p); err != nil {
		log.Errorf("error deleting object, object is no longer indexed by cache")
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
	switch err := err.(type) {
	case ErrCacheRace:
		return true
	default:
		return c.fast.IsRetryable(err) || c.slow.IsRetryable(err)
	}
}

func (c *cacheClient) lockPath(p string) {
	v, _ := c.pathLocks.LoadOrStore(p, &sync.Mutex{})
	v.(*sync.RWMutex).Lock()
}

func (c *cacheClient) rlockPath(p string) {
	v, _ := c.pathLocks.LoadOrStore(p, &sync.Mutex{})
	v.(*sync.RWMutex).RLock()
}

func (c *cacheClient) unlockPath(p string) {
	v, exists := c.pathLocks.Load(p)
	if !exists {
		panic("unlock path that was never locked: " + p)
	}
	v.(*sync.RWMutex).Unlock()
}

func (c *cacheClient) runlockPath(p string) {
	v, exists := c.pathLocks.Load(p)
	if !exists {
		panic("unlock path that was never locked: " + p)
	}
	v.(*sync.RWMutex).RUnlock()
}

func (c *cacheClient) onEvicted(key, value interface{}) {
	p := key.(string)
	c.lockPath(p)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := c.fast.Delete(ctx, p); err != nil {
		log.Error("could not delete from cache's fast store")
	}
	cancel()
	c.unlockPath(p)
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

type ErrCacheRace struct {
}

func (e ErrCacheRace) Error() string {
	return "cache race. retry"
}
