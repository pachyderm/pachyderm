package localcache

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Cache is a simple unbounded disk cache and is safe for concurrency.
type Cache struct {
	root string
	keys map[string]struct{}
	mu   sync.Mutex
}

// NewCache creates a new cache.
func NewCache(root string) *Cache {
	return &Cache{
		root: root,
		keys: make(map[string]struct{}),
	}
}

// Put puts a key/value pair in the cache and reads the value from an io.Reader.
func (c *Cache) Put(key string, value io.Reader) (retErr error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, err := os.Create(filepath.Join(c.root, key))
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if _, err := io.Copy(f, value); err != nil {
		return err
	}
	c.keys[key] = struct{}{}
	return nil
}

// Get gets a key's value by returning an io.ReadCloser that should be closed when done.
func (c *Cache) Get(key string) (io.ReadCloser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.keys[key]; !ok {
		return nil, fmt.Errorf("key %v not found in cache", key)
	}
	f, err := os.Open(filepath.Join(c.root, key))
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Keys returns the keys in sorted order.
func (c *Cache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var keys []string
	for key := range c.keys {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Delete deletes a key/value pair.
func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.keys[key]; !ok {
		return nil
	}
	delete(c.keys, key)
	return os.Remove(filepath.Join(c.root, key))
}

// Clear clears the cache.
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() {
		c.keys = make(map[string]struct{})
	}()
	for key := range c.keys {
		if err := os.Remove(filepath.Join(c.root, key)); err != nil {
			return err
		}
	}
	return nil
}
