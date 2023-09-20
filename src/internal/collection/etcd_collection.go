package collection

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// defaultLimit was experimentally determined to be the highest value that could work
// (It gets scaled down for specific collections if they trip the max-message size.)
const (
	defaultLimit    int64  = 262144
	DefaultPrefix   string = "pachyderm/1.7.0"
	indexIdentifier string = "__index_"
)

var (
	// ErrNotClaimed is an error used to indicate that a different requester beat
	// the current requester to a key claim.
	ErrNotClaimed = errors.New("NOT_CLAIMED")
	DefaultTTL    = int64(30)
)

type etcdCollection struct {
	etcdClient *etcd.Client
	prefix     string
	indexes    []*Index
	// The limit used when listing the collection. This gets automatically
	// tuned when requests fail so it's stored per collection.
	limit int64
	// We need this to figure out the concrete type of the objects
	// that this collection is storing. Not sure what else we can
	// do since types in Go are not first-class objects.
	// To be clear, this is only necessary because of `Delete`, where we
	// need to know the type in order to properly remove secondary indexes.
	template proto.Message
	// keyCheck is a function that checks if a key is valid.  Invalid keys
	// cannot be created.
	keyCheck func(string) error

	// valCheck is a function that checks if a value is valid.
	valCheck func(proto.Message) error
}

// NewEtcdCollection creates a new collection backed by etcd.
func NewEtcdCollection(etcdClient *etcd.Client, prefix string, indexes []*Index, template proto.Message, keyCheck func(string) error, valCheck func(proto.Message) error) EtcdCollection {
	// We want to ensure that the prefix always ends with a trailing
	// slash.  Otherwise, when you list the items under a collection
	// such as `foo`, you might end up listing items under `foobar`
	// as well.
	if len(prefix) > 0 && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}

	return &etcdCollection{
		prefix:     prefix,
		etcdClient: etcdClient,
		indexes:    indexes,
		limit:      defaultLimit,
		template:   template,
		keyCheck:   keyCheck,
		valCheck:   valCheck,
	}
}

func (c *etcdCollection) ReadWrite(stm STM) EtcdReadWriteCollection {
	return &etcdReadWriteCollection{
		etcdCollection: c,
		stm:            stm,
	}
}

func (c *etcdCollection) ReadOnly(ctx context.Context) EtcdReadOnlyCollection {
	return &etcdReadOnlyCollection{
		etcdCollection: c,
		ctx:            ctx,
	}
}

func (c *etcdCollection) Claim(ctx context.Context, key string, val proto.Message, cb func(context.Context) error) error {
	var claimed bool
	if _, err := NewSTM(ctx, c.etcdClient, func(stm STM) error {
		readWriteC := c.ReadWrite(stm)
		if err := readWriteC.Get(key, val); err != nil {
			if !IsErrNotFound(err) {
				return errors.EnsureStack(err)
			}
			claimed = true
			return errors.EnsureStack(readWriteC.PutTTL(key, val, DefaultTTL))
		}
		claimed = false
		return nil
	}); err != nil {
		return err
	}
	if !claimed {
		return ErrNotClaimed
	}
	return c.WithRenewer(ctx, func(ctx context.Context, renewer *Renewer) error {
		if err := renewer.Put(ctx, key, val); err != nil {
			return err
		}
		return cb(ctx)
	})
}

type Renewer struct {
	client     *etcd.Client
	collection *etcdCollection
	lease      etcd.LeaseID
}

func newRenewer(client *etcd.Client, collection *etcdCollection, lease etcd.LeaseID) *Renewer {
	return &Renewer{
		client:     client,
		collection: collection,
		lease:      lease,
	}
}

func (r *Renewer) Put(ctx context.Context, key string, val proto.Message) error {
	_, err := NewSTM(ctx, r.client, func(stm STM) error {
		// TODO: This is a bit messy, but I don't want put lease to be a part of the exported interface.
		col := &etcdReadWriteCollection{
			etcdCollection: r.collection,
			stm:            stm,
		}
		return col.putLease(key, val, r.lease)
	})
	return err
}

func (c *etcdCollection) WithRenewer(ctx context.Context, cb func(context.Context, *Renewer) error) error {
	resp, err := c.etcdClient.Grant(ctx, DefaultTTL)
	if err != nil {
		return errors.EnsureStack(err)
	}
	var cancel context.CancelFunc
	ctx, cancel = pctx.WithCancel(pctx.Child(ctx, "WithRenewer"))
	defer cancel()
	keepAliveChan, err := c.etcdClient.KeepAlive(ctx, resp.ID)
	if err != nil {
		return errors.EnsureStack(err)
	}
	go func() {
		for {
			_, more := <-keepAliveChan
			if !more {
				if ctx.Err() == nil {
					log.Error(ctx, "failed to renew etcd lease", zap.Stack("stacktrace"))
					cancel()
				}
				return
			}
		}
	}()
	renewer := newRenewer(c.etcdClient, c, resp.ID)
	return cb(ctx, renewer)
}

// path returns the full path of a key in the etcd namespace
func (c *etcdCollection) path(key string) string {
	return path.Join(c.prefix, key)
}

func (c *etcdCollection) indexRoot(index *Index) string {
	// remove trailing slash from c.prefix
	return fmt.Sprintf("%s%s%s/",
		strings.TrimRight(c.prefix, "/"), indexIdentifier, index.Name)
}

// See the documentation for `Index` for details.
func (c *etcdCollection) indexDir(index *Index, indexVal string) string {
	return path.Join(c.indexRoot(index), indexVal)
}

// See the documentation for `Index` for details.
func (c *etcdCollection) indexPath(index *Index, indexVal string, key string) string {
	return path.Join(c.indexDir(index, indexVal), key)
}

type etcdReadWriteCollection struct {
	*etcdCollection
	stm STM
}

func (c *etcdReadWriteCollection) Get(maybeKey interface{}, val proto.Message) (retErr error) {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	span, _ := tracing.AddSpanToAnyExisting(c.stm.Context(), "/etcd.RW/Get",
		"col", c.prefix, "key", strings.TrimPrefix(key, c.prefix))
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	valStr, err := c.stm.Get(c.path(key))
	if err != nil {
		if IsErrNotFound(err) {
			return ErrNotFound{Type: c.prefix, Key: key}
		}
		return errors.EnsureStack(err)
	}
	c.stm.SetSafePutCheck(c.path(key), reflect.ValueOf(val).Pointer())
	return errors.EnsureStack(proto.Unmarshal([]byte(valStr), val))
}

func cloneProtoMsg(original proto.Message) proto.Message {
	val := reflect.ValueOf(original)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	return reflect.New(val.Type()).Interface().(proto.Message)
}

// Giving a value, an index, and the key of the item, return the path
// under which the new index item should be stored.
func (c *etcdReadWriteCollection) getIndexPath(val proto.Message, index *Index, key string) string {
	return c.indexPath(index, index.Extract(val), key)
}

func (c *etcdReadWriteCollection) Put(maybeKey interface{}, val proto.Message) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	return c.PutTTL(key, val, 0)
}

func (c *etcdReadWriteCollection) TTL(key string) (int64, error) {
	ttl, err := c.stm.TTL(c.path(key))
	if IsErrNotFound(err) {
		return ttl, ErrNotFound{Type: c.prefix, Key: key}
	}
	return ttl, errors.EnsureStack(err)
}

func (c *etcdReadWriteCollection) PutTTL(key string, val proto.Message, ttl int64) error {
	return c.put(key, val, func(key string, val string, ptr uintptr) error {
		return errors.EnsureStack(c.stm.Put(key, val, ttl, ptr))
	})
}

func (c *etcdReadWriteCollection) put(key string, val proto.Message, putFunc func(string, string, uintptr) error) error {
	if strings.Contains(key, indexIdentifier) {
		return errors.Errorf("cannot put key %q which contains reserved string %q", key, indexIdentifier)
	}
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}

	if c.keyCheck != nil {
		if err := c.keyCheck(key); err != nil {
			return err
		}
	}

	if c.valCheck != nil {
		if err := c.valCheck(val); err != nil {
			return err
		}
	}

	ptr := reflect.ValueOf(val).Pointer()
	if !c.stm.IsSafePut(c.path(key), ptr) {
		return errors.Errorf("unsafe put for key %v (passed ptr did not receive updated value)", key)
	}

	if c.indexes != nil {
		clone := cloneProtoMsg(val)

		// Put the appropriate record in any of c's secondary indexes
		for _, index := range c.indexes {
			indexPath := c.getIndexPath(val, index, key)
			// If we can get the original value, we remove the original indexes
			if err := c.Get(key, clone); err == nil {
				originalIndexPath := c.getIndexPath(clone, index, key)
				if originalIndexPath != indexPath {
					c.stm.Del(originalIndexPath)
				}
			}
			// Put the index even if it already exists, so that watchers may be
			// notified that the value has been updated.
			if err := putFunc(indexPath, key, 0); err != nil {
				return err
			}
		}
	}
	bytes, err := proto.Marshal(val)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return putFunc(c.path(key), string(bytes), ptr)
}

func (c *etcdReadWriteCollection) putLease(key string, val proto.Message, lease etcd.LeaseID) error {
	return c.put(key, val, func(key string, val string, ptr uintptr) error {
		return errors.EnsureStack(c.stm.PutLease(key, val, lease, ptr))
	})
}

func (c *etcdReadWriteCollection) putIgnoreLease(key string, val proto.Message) error {
	return c.put(key, val, func(key string, val string, ptr uintptr) error {
		return errors.EnsureStack(c.stm.PutIgnoreLease(key, val, ptr))
	})
}

// Update reads the current value associated with 'key', calls 'f' to update
// the value, and writes the new value back to the collection. 'key' must be
// present in the collection, or a 'Not Found' error is returned
func (c *etcdReadWriteCollection) Update(maybeKey interface{}, val proto.Message, f func() error) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	if err := c.Get(key, val); err != nil {
		if IsErrNotFound(err) {
			return ErrNotFound{Type: c.prefix, Key: key}
		}
		return err
	}
	if err := f(); err != nil {
		return err
	}
	return c.putIgnoreLease(key, val)
}

// Upsert is like Update but 'key' is not required to be present
func (c *etcdReadWriteCollection) Upsert(maybeKey interface{}, val proto.Message, f func() error) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	if err := c.Get(key, val); err != nil && !IsErrNotFound(err) {
		return err
	}
	if err := f(); err != nil {
		return err
	}
	return c.Put(key, val)
}

func (c *etcdReadWriteCollection) Create(maybeKey interface{}, val proto.Message) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	fullKey := c.path(key)
	_, err := c.stm.Get(fullKey)
	if err != nil && !IsErrNotFound(err) {
		return errors.EnsureStack(err)
	}
	if err == nil {
		return ErrExists{Type: c.prefix, Key: key}
	}
	return c.Put(key, val)
}

func (c *etcdReadWriteCollection) Delete(maybeKey interface{}) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	fullKey := c.path(key)
	if _, err := c.stm.Get(fullKey); err != nil {
		return errors.EnsureStack(err)
	}
	if c.indexes != nil && c.template != nil {
		val := proto.Clone(c.template)
		for _, index := range c.indexes {
			if err := c.Get(key, val); err == nil {
				indexPath := c.getIndexPath(val, index, key)
				c.stm.Del(indexPath)
			}
		}
	}
	c.stm.Del(fullKey)
	return nil
}

func (c *etcdReadWriteCollection) DeleteAll() error {
	// Delete indexes
	for _, index := range c.indexes {
		c.stm.DelAll(c.indexRoot(index))
	}
	c.stm.DelAll(c.prefix)
	return nil
}

func (c *etcdReadWriteCollection) DeleteAllPrefix(prefix string) error {
	c.stm.DelAll(path.Join(c.prefix, prefix) + "/")
	return nil
}

type etcdReadOnlyCollection struct {
	*etcdCollection
	ctx context.Context
}

// get is an internal wrapper around etcdClient.Get that wraps the call in a
// trace
func (c *etcdReadOnlyCollection) get(key string, opts ...etcd.OpOption) (resp *etcd.GetResponse, retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(c.ctx, "/etcd.RO/Get",
		"col", c.prefix, "key", strings.TrimPrefix(key, c.prefix))
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	resp, err := c.etcdClient.Get(ctx, key, opts...)
	return resp, errors.EnsureStack(err)
}

func (c *etcdReadOnlyCollection) Get(maybeKey interface{}, val proto.Message) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	resp, err := c.get(c.path(key))
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound{Type: c.prefix, Key: key}
	}

	return errors.EnsureStack(proto.Unmarshal(resp.Kvs[0].Value, val))
}

func (c *etcdReadOnlyCollection) GetByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(key string) error) error {
	span, _ := tracing.AddSpanToAnyExisting(c.ctx, "/etcd.RO/GetByIndex", "col", c.prefix, "index", index, "indexVal", indexVal)
	defer tracing.FinishAnySpan(span)
	if atomic.LoadInt64(&index.limit) == 0 {
		atomic.CompareAndSwapInt64(&index.limit, 0, defaultLimit)
	}
	return c.list(c.indexDir(index, indexVal), &index.limit, opts, func(kv *mvccpb.KeyValue) error {
		key := path.Base(string(kv.Key))
		if err := c.Get(key, val); err != nil {
			if IsErrNotFound(err) {
				// In cases where we changed how certain objects are
				// indexed, we could end up in a situation where the
				// object is deleted but the old indexes still exist.
				return nil
			}
			return err
		}
		return f(key)
	})
}

func (c *etcdReadOnlyCollection) TTL(key string) (int64, error) {
	resp, err := c.get(c.path(key))
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, ErrNotFound{Type: c.prefix, Key: key}
	}
	leaseID := etcd.LeaseID(resp.Kvs[0].Lease)

	span, ctx := tracing.AddSpanToAnyExisting(c.ctx, "/etcd.RO/TimeToLive")
	defer tracing.FinishAnySpan(span)
	leaseTTLResp, err := c.etcdClient.TimeToLive(ctx, leaseID)
	if err != nil {
		return 0, errors.Wrapf(err, "could not fetch lease TTL")
	}
	return leaseTTLResp.TTL, nil
}

// List returns objects sorted based on the options passed in. f will be called
// for each item, val will contain the corresponding value. Val is not an
// argument to f because that would require f to perform a cast before it could
// be used.
// You can break out of iteration by returning errutil.ErrBreak.
func (c *etcdReadOnlyCollection) List(val proto.Message, opts *Options, f func(key string) error) error {
	span, _ := tracing.AddSpanToAnyExisting(c.ctx, "/etcd.RO/List", "col", c.prefix)
	defer tracing.FinishAnySpan(span)
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	return c.list(c.prefix, &c.limit, opts, func(kv *mvccpb.KeyValue) error {
		if err := proto.Unmarshal(kv.Value, val); err != nil {
			return errors.EnsureStack(err)
		}
		return f(strings.TrimPrefix(string(kv.Key), c.prefix))
	})
}

// ListRev returns objects sorted based on the options passed in. f will be called
// with each key and the create-revision of the key, val will contain the
// corresponding value. Val is not an argument to f because that would require
// f to perform a cast before it could be used.  You can break out of iteration
// by returning errutil.ErrBreak.
func (c *etcdReadOnlyCollection) ListRev(val proto.Message, opts *Options, f func(key string, createRev int64) error) error {
	span, _ := tracing.AddSpanToAnyExisting(c.ctx, "/etcd.RO/List", "col", c.prefix)
	defer tracing.FinishAnySpan(span)
	if err := watch.CheckType(c.template, val); err != nil {
		return err
	}
	return c.list(c.prefix, &c.limit, opts, func(kv *mvccpb.KeyValue) error {
		if err := proto.Unmarshal(kv.Value, val); err != nil {
			return errors.EnsureStack(err)
		}
		return f(strings.TrimPrefix(string(kv.Key), c.prefix), kv.CreateRevision)
	})
}

func (c *etcdReadOnlyCollection) list(prefix string, limitPtr *int64, opts *Options, f func(*mvccpb.KeyValue) error) error {
	return listRevision(c, prefix, limitPtr, opts, f)
}

var (
	countOpts = []etcd.OpOption{etcd.WithPrefix(), etcd.WithCountOnly()}
)

func (c *etcdReadOnlyCollection) Count() (int64, error) {
	resp, err := c.get(c.prefix, countOpts...)
	if err != nil {
		return 0, err
	}
	return resp.Count, err
}

func (c *etcdReadOnlyCollection) CountRev(rev int64) (int64, int64, error) {
	resp, err := c.get(c.prefix, append(countOpts, etcd.WithRev(rev))...)
	if err != nil {
		return 0, 0, err
	}
	return resp.Count, resp.Header.Revision, err
}

// Watch a collection, returning the current content of the collection as
// well as any future additions.
func (c *etcdReadOnlyCollection) Watch(opts ...watch.Option) (watch.Watcher, error) {
	return watch.NewEtcdWatcher(c.ctx, c.etcdClient, c.prefix, c.prefix, c.template, opts...)
}

// WatchF watches a collection and executes a callback function each time an event occurs.
func (c *etcdReadOnlyCollection) WatchF(f func(e *watch.Event) error, opts ...watch.Option) error {
	watcher, err := c.Watch(opts...)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchF(c.ctx, watcher, f)
}

func watchF(ctx context.Context, watcher watch.Watcher, f func(e *watch.Event) error) error {
	for {
		select {
		case e, ok := <-watcher.Watch():
			if !ok {
				return nil
			}
			if e.Type == watch.EventError {
				return e.Err
			}
			if err := f(e); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		}
	}
}

// WatchByIndex watches items in a collection that match a particular index
func (c *etcdReadOnlyCollection) WatchByIndex(index *Index, val string, opts ...watch.Option) (watch.Watcher, error) {
	eventCh := make(chan *watch.Event)
	done := make(chan struct{})
	watcher, err := watch.NewEtcdWatcher(c.ctx, c.etcdClient, c.prefix, c.indexDir(index, val), c.template, opts...)
	if err != nil {
		return nil, err
	}
	go func() {
		watchForEvents := func() error {
			for {
				var ev *watch.Event
				var ok bool
				select {
				case ev, ok = <-watcher.Watch():
				case <-done:
					watcher.Close()
					return nil
				}
				if !ok {
					watcher.Close()
					return nil
				}

				var directEv *watch.Event
				switch ev.Type {
				case watch.EventError:
					// pass along the error
					return ev.Err
				case watch.EventPut:
					resp, err := c.get(c.path(path.Base(string(ev.Key))))
					if err != nil {
						return err
					}
					if len(resp.Kvs) == 0 {
						// this happens only if the item was deleted shortly after
						// we receive this event.
						continue
					}
					directEv = &watch.Event{
						Key:      []byte(path.Base(string(ev.Key))),
						Value:    resp.Kvs[0].Value,
						Type:     ev.Type,
						Template: c.template,
					}
				case watch.EventDelete:
					directEv = &watch.Event{
						Key:      []byte(path.Base(string(ev.Key))),
						Type:     ev.Type,
						Template: c.template,
					}
				}
				eventCh <- directEv
			}
		}
		defer close(eventCh)
		if err := watchForEvents(); err != nil {
			eventCh <- &watch.Event{Type: watch.EventError, Err: err}
			watcher.Close()
		}
	}()
	return watch.MakeEtcdWatcher(eventCh, done), nil
}

func (c *etcdReadOnlyCollection) WatchByIndexF(index *Index, indexVal string, f func(e *watch.Event) error, opts ...watch.Option) error {
	watcher, err := c.WatchByIndex(index, indexVal, opts...)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchF(c.ctx, watcher, f)
}

// WatchOne watches a given item.  The first value returned from the watch
// will be the current value of the item.
func (c *etcdReadOnlyCollection) WatchOne(maybeKey interface{}, opts ...watch.Option) (watch.Watcher, error) {
	key, ok := maybeKey.(string)
	if !ok {
		return nil, errors.New("key must be a string")
	}
	return watch.NewEtcdWatcher(c.ctx, c.etcdClient, c.prefix, c.path(key), c.template, opts...)
}

// WatchOneF watches a given item and executes a callback function each time an event occurs.
// The first value returned from the watch will be the current value of the item.
func (c *etcdReadOnlyCollection) WatchOneF(maybeKey interface{}, f func(e *watch.Event) error, opts ...watch.Option) error {
	key, ok := maybeKey.(string)
	if !ok {
		return errors.New("key must be a string")
	}
	watcher, err := watch.NewEtcdWatcher(c.ctx, c.etcdClient, c.prefix, c.path(key), c.template, opts...)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchF(c.ctx, watcher, f)
}
