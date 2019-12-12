package hashtree

import (
	"fmt"
	"io"
	"reflect"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pachyderm/pachyderm/src/server/pkg/localcache"
	"github.com/sirupsen/logrus"
)

// Cache is an LRU cache for hashtrees.
type Cache struct {
	*lru.Cache
	syncEvict *bool
}

func evict(value interface{}) {
	tree, ok := value.(*dbHashTree)
	if !ok {
		logrus.Infof("non hashtree slice value of type: %v", reflect.TypeOf(value))
		return
	}
	if err := tree.Destroy(); err != nil {
		logrus.Infof("failed to destroy hashtree: %v", err)
	}
}

// NewCache creates a new cache.
func NewCache(size int) (*Cache, error) {
	syncEvict := false
	c, err := lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		if syncEvict {
			evict(value)
		} else {
			go evict(value)
		}
	})
	if err != nil {
		return nil, err
	}
	return &Cache{Cache: c, syncEvict: &syncEvict}, nil
}

// Close will synchronously evict all hashtrees from the cache, cleaning up any on-disk data.
func (c *Cache) Close() {
	*c.syncEvict = true
	c.Purge()
}

// MergeCache is an unbounded hashtree cache that can merge the hashtrees in the cache.
type MergeCache struct {
	*localcache.Cache
}

// NewMergeCache creates a new cache.
func NewMergeCache(root string) *MergeCache {
	return &MergeCache{
		Cache: localcache.NewCache(root),
	}
}

// Put puts an id/hashtree pair in the cache and reads the hashtree from the passed in io.Reader.
func (c *MergeCache) Put(id int64, tree io.Reader) (retErr error) {
	return c.Cache.Put(fmt.Sprint(id), tree)
}

// Get does a filtered write of id's hashtree to the passed in io.Writer.
func (c *MergeCache) Get(id int64, w io.Writer, filter Filter) (retErr error) {
	r, err := c.Cache.Get(fmt.Sprint(id))
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
func (c *MergeCache) Delete(id int64) error {
	return c.Cache.Delete(fmt.Sprint(id))
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
