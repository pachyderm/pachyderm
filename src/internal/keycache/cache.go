package keycache

import (
	"fmt"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

// Cache watches a key in etcd and caches the value in an atomic value
// This is useful for frequently read but infrequently updated values
type Cache struct {
	ctx          context.Context
	c            col.Collection
	defaultValue proto.Message
	key          string
	value        *atomic.Value
}

// NewCache returns a cache for the given key in the etcd collection
func NewCache(ctx context.Context, c col.Collection, key string, defaultValue proto.Message) *Cache {
	value := &atomic.Value{}
	value.Store(defaultValue)
	return &Cache{
		ctx:          ctx,
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
			select {
			case <-c.ctx.Done():
				return errors.New("done")
			case ev, ok := <-watcher.Watch():
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
		}
	}, backoff.NewInfiniteBackOff(), backoff.NotifyCtx(c.ctx, fmt.Sprintf("watcher for %v", c.key)))
}

// Load retrieves the current cached value
func (c *Cache) Load() interface{} {
	return c.value.Load()
}
