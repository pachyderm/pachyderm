package keycache

import (
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/src/internal/collection"
	"github.com/pachyderm/pachyderm/src/internal/errors"
	"github.com/pachyderm/pachyderm/src/internal/watch"
)

// Cache watches a key in etcd and caches the value in an atomic value
// This is useful for frequently read but infrequently updated values
type Cache struct {
	c            col.Collection
	defaultValue proto.Message
	key          string
	value        *atomic.Value
}

// NewCache returns a cache for the given key in the etcd collection
func NewCache(c col.Collection, key string, defaultValue proto.Message) *Cache {
	value := &atomic.Value{}
	value.Store(defaultValue)
	return &Cache{
		c:            c,
		value:        value,
		key:          key,
		defaultValue: defaultValue,
	}
}

// Watch should be called in a goroutine to start the watcher
func (c *Cache) Watch() {
	backoff.RetryNotify(func() error {
		watcher, err := c.c.ReadOnly(context.Background()).Watch()
		if err != nil {
			return err
		}
		defer watcher.Close()
		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return errors.New("watch closed unexpectedly")
			}

			if ev.Type == watch.EventError {
				return ev.Err
			}

			if string(ev.Key) == c.key {
				switch ev.Type {
				case watch.EventPut:
					val := proto.Clone(c.defaultValue)
					if err := proto.Unmarshal(ev.Value, val); err != nil {
						return err
					}
					c.value.Store(val)
				case watch.EventDelete:
					c.value.Store(c.defaultValue)
				}
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logrus.Printf("error from watcher for %v: %v; retrying in %v", c.key, err, d)
		return nil
	})
}

// Load retrieves the current cached value
func (c *Cache) Load() interface{} {
	return c.value.Load()
}
