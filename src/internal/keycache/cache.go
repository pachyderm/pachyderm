package keycache

import (
	"context"
	"fmt"
	"sync/atomic"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

// Cache watches a key in etcd and caches the value in an atomic value
// This is useful for frequently read but infrequently updated values
type Cache struct {
	ctx          context.Context
	readOnly     col.ReadOnlyCollection
	defaultValue proto.Message
	key          string
	value        *atomic.Value
}

// NewCache returns a cache for the given key in the etcd collection
func NewCache(ctx context.Context, readOnly col.ReadOnlyCollection, key string, defaultValue proto.Message) *Cache {
	value := &atomic.Value{}
	value.Store(defaultValue)
	return &Cache{
		readOnly:     readOnly,
		ctx:          ctx,
		value:        value,
		key:          key,
		defaultValue: defaultValue,
	}
}

// Watch should be called in a goroutine to start the watcher
func (c *Cache) Watch() {
	backoff.RetryNotify(func() error { //nolint:errcheck
		err := c.readOnly.WatchOneF(c.key, func(ev *watch.Event) error {
			switch ev.Type {
			case watch.EventPut:
				val := proto.Clone(c.defaultValue)
				if err := proto.Unmarshal(ev.Value, val); err != nil {
					return errors.EnsureStack(err)
				}
				c.value.Store(val)
			case watch.EventDelete:
				c.value.Store(c.defaultValue)
			}
			return nil
		})
		return errors.EnsureStack(err)
	}, backoff.NewInfiniteBackOff(), backoff.NotifyCtx(c.ctx, fmt.Sprintf("watcher for %v", c.key)))
}

// Load retrieves the current cached value
func (c *Cache) Load() interface{} {
	return proto.Clone(c.value.Load().(proto.Message))
}
