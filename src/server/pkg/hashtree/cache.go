package hashtree

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/pachyderm/pachyderm/src/server/pkg/localcache"
	"github.com/sirupsen/logrus"
)

type seedling struct {
	cond *sync.Cond
	err  error
}

// Cache is an LRU cache for hashtrees.
type Cache struct {
	lruCache  simplelru.LRUCache
	lock      sync.Mutex
	syncEvict *bool
	seedlings map[interface{}]*seedling
}

func castValue(value interface{}) (HashTree, error) {
	tree, ok := value.(HashTree)
	if !ok {
		return nil, fmt.Errorf("corrupted cache: expected HashTree, found: %v", reflect.TypeOf(value))
	}
	return tree, nil
}

func evict(value interface{}) {
	if tree, err := castValue(value); err != nil {
		logrus.Infof(err.Error())
	} else {
		if err := tree.Destroy(); err != nil {
			logrus.Infof("failed to destroy hashtree: %v", err)
		}
	}
}

// NewCache creates a new cache.
func NewCache(size int) (*Cache, error) {
	syncEvict := false
	lruCache, err := simplelru.NewLRU(size, func(key interface{}, value interface{}) {
		if syncEvict {
			evict(value)
		} else {
			go evict(value)
		}
	})
	if err != nil {
		return nil, err
	}
	return &Cache{
		lruCache:  lruCache,
		syncEvict: &syncEvict,
		seedlings: make(map[interface{}]*seedling),
	}, nil
}

// Close will synchronously evict all hashtrees from the cache, cleaning up any on-disk data.
func (c *Cache) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	*c.syncEvict = true
	c.lruCache.Purge()
}

// GetOrAdd will atomically attempt to get the given HashTree from the cache,
// and if it does not exist, it will generate the HashTree using the provided
// function and store it in the cache before returning it. This is implemented
// because the underlying LRU library does not provide a way to atomically
// perform this operation without preconstructing the HashTree, which is a
// relatively expensive operation.
// If the generator panics, the cache will be left in an unreliable state.
func (c *Cache) GetOrAdd(key interface{}, generator func() (HashTree, error)) (HashTree, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// First try to get a tree from the cache
	if value, ok := c.lruCache.Get(key); ok {
		return castValue(value)
	} else if s, ok := c.seedlings[key]; ok {
		// There is a pending hashtree being generated, wait for it
		s.cond.Wait()

		if s.err != nil {
			return nil, s.err
		} else if value, ok := c.lruCache.Get(key); ok {
			return castValue(value)
		}

		// If we get here, that means the hashtree was evicted between when the
		// seedling completed and when this goroutine woke up. Fallthrough so
		// that we can run the generator to reconstruct it.
	}

	s := &seedling{cond: sync.NewCond(&c.lock)}
	c.seedlings[key] = s
	c.lock.Unlock()

	// We do the generator outside of the lock so we don't bottleneck on this,
	// as it may be a long-running operation.
	newValue, err := generator()
	if err != nil {
		c.lock.Lock()
		delete(c.seedlings, key)
		s.err = err
		s.cond.Broadcast()

		return nil, err
	}

	c.lock.Lock()
	c.lruCache.Add(key, newValue)
	delete(c.seedlings, key)
	s.cond.Broadcast()

	return newValue, nil
}

// MergeCache is an unbounded hashtree cache that can merge the hashtrees in the cache.
type MergeCache struct {
	*localcache.Cache
}

// NewMergeCache creates a new cache.
func NewMergeCache(root string) (*MergeCache, error) {
	cache, err := localcache.NewCache(root)
	if err != nil {
		return nil, err
	}

	return &MergeCache{Cache: cache}, nil
}

// Put puts an id/hashtree pair in the cache and reads the hashtree from the passed in io.Reader.
func (c *MergeCache) Put(id string, tree io.Reader) (retErr error) {
	return c.Cache.Put(id, tree)
}

// Get does a filtered write of id's hashtree to the passed in io.Writer.
func (c *MergeCache) Get(id string, w io.Writer, filter Filter) (retErr error) {
	r, err := c.Cache.Get(id)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return NewWriter(w).Copy(NewReader(r, filter))
}

// Delete deletes a hashtree from the cache.
func (c *MergeCache) Delete(id string) error {
	return c.Cache.Delete(id)
}

// Has returns true if the given id is present in the cache, false otherwise.
func (c *MergeCache) Has(id string) bool {
	return c.Cache.Has(id)
}

// Merge does a filtered merge of the hashtrees in the cache.
// The results are written to the passed in *Writer.
// The base field is used as the base hashtree if it is non-nil
func (c *MergeCache) Merge(w *Writer, base io.Reader, filter Filter) (retErr error) {
	var trees []*Reader
	if base != nil {
		trees = append(trees, NewReader(base, filter))
	}
	for _, key := range c.Keys() {
		r, err := c.Cache.Get(key)
		if err != nil {
			return err
		}
		defer func() {
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		trees = append(trees, NewReader(r, filter))
	}
	return Merge(w, trees)
}
