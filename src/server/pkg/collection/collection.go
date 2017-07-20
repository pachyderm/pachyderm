package collection

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

type collection struct {
	etcdClient *etcd.Client
	prefix     string
	indexes    []Index
	// We need this to figure out the concrete type of the objects
	// that this collection is storing. It's pretty retarded, but
	// not sure what else we can do since types in Go are not first-class
	// objects.
	// To be clear, this is only necessary because of `Delete`, where we
	// need to know the type in order to properly remove secondary indexes.
	template proto.Message
	// keyCheck is a function that checks if a key is valid.  Invalid keys
	// cannot be created.
	keyCheck func(string) error
}

// NewCollection creates a new collection.
func NewCollection(etcdClient *etcd.Client, prefix string, indexes []Index, template proto.Message, keyCheck func(string) error) Collection {
	// We want to ensure that the prefix always ends with a trailing
	// slash.  Otherwise, when you list the items under a collection
	// such as `foo`, you might end up listing items under `foobar`
	// as well.
	if len(prefix) > 0 && prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}

	return &collection{
		prefix:     prefix,
		etcdClient: etcdClient,
		indexes:    indexes,
		template:   template,
		keyCheck:   keyCheck,
	}
}

func (c *collection) ReadWrite(stm STM) ReadWriteCollection {
	return &readWriteCollection{
		collection: c,
		stm:        stm,
	}
}

func (c *collection) ReadWriteInt(stm STM) ReadWriteIntCollection {
	return &readWriteIntCollection{
		collection: c,
		stm:        stm,
	}
}

func (c *collection) ReadOnly(ctx context.Context) ReadonlyCollection {
	return &readonlyCollection{
		collection: c,
		ctx:        ctx,
	}
}

// Path returns the full path of a key in the etcd namespace
func (c *collection) Path(key string) string {
	return path.Join(c.prefix, key)
}

// See the documentation for `Index` for details.
func (c *collection) indexDir(index Index, indexVal string) string {
	indexDir := c.prefix
	// remove trailing slash
	indexDir = strings.TrimRight(indexDir, "/")
	return fmt.Sprintf("%s__index_%s/%s", indexDir, index, indexVal)
}

// See the documentation for `Index` for details.
func (c *collection) indexPath(index Index, indexVal string, key string) string {
	return path.Join(c.indexDir(index, indexVal), key)
}

type readWriteCollection struct {
	*collection
	stm STM
}

func (c *readWriteCollection) Get(key string, val proto.Unmarshaler) error {
	valStr := c.stm.Get(c.Path(key))
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	return val.Unmarshal([]byte(valStr))
}

func cloneProtoMsg(original proto.Marshaler) proto.Unmarshaler {
	val := reflect.ValueOf(original)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	return reflect.New(val.Type()).Interface().(proto.Unmarshaler)
}

// Giving a value, an index, and the key of the item, return the path
// under which the new index item should be stored.
func (c *readWriteCollection) getIndexPath(val interface{}, index Index, key string) string {
	r := reflect.ValueOf(val)
	f := reflect.Indirect(r).FieldByName(index.Field).Interface()
	indexKey := fmt.Sprintf("%s", f)
	return c.indexPath(index, indexKey, key)
}

// Giving a value, a multi-index, and the key of the item, return the
// paths under which the multi-index items should be stored.
func (c *readWriteCollection) getMultiIndexPaths(val interface{}, index Index, key string) []string {
	var indexPaths []string
	f := reflect.Indirect(reflect.ValueOf(val)).FieldByName(index.Field)
	for i := 0; i < f.Len(); i++ {
		indexKey := fmt.Sprintf("%s", f.Index(i).Interface())
		indexPaths = append(indexPaths, c.indexPath(index, indexKey, key))
	}
	return indexPaths
}

func (c *readWriteCollection) Put(key string, val proto.Marshaler) error {
	return c.PutTTL(key, val, 0)
}

func (c *readWriteCollection) PutTTL(key string, val proto.Marshaler, ttl int64) error {
	var options []etcd.OpOption
	if ttl > 0 {
		lease, err := c.collection.etcdClient.Grant(context.Background(), ttl)
		if err != nil {
			return fmt.Errorf("error granting lease: %v", err)
		}
		options = append(options, etcd.WithLease(lease.ID))
	}

	if c.collection.keyCheck != nil {
		if err := c.collection.keyCheck(key); err != nil {
			return err
		}
	}
	if c.indexes != nil {
		clone := cloneProtoMsg(val)

		// Put the appropriate record in any of c's secondary indexes
		for _, index := range c.indexes {
			if index.Multi {
				indexPaths := c.getMultiIndexPaths(val, index, key)
				for _, indexPath := range indexPaths {
					// Only put the index if it doesn't already exist; otherwise
					// we might trigger an unnecessary event if someone is
					// watching the index
					if c.stm.Get(indexPath) == "" {
						c.stm.Put(indexPath, key, options...)
					}
				}
				// If we can get the original value, we remove the original indexes
				if err := c.Get(key, clone); err == nil {
					for _, originalIndexPath := range c.getMultiIndexPaths(clone, index, key) {
						var found bool
						for _, indexPath := range indexPaths {
							if originalIndexPath == indexPath {
								found = true
							}
						}
						if !found {
							c.stm.Del(originalIndexPath)
						}
					}
				}
			} else {
				indexPath := c.getIndexPath(val, index, key)
				// If we can get the original value, we remove the original indexes
				if err := c.Get(key, clone); err == nil {
					originalIndexPath := c.getIndexPath(clone, index, key)
					if originalIndexPath != indexPath {
						c.stm.Del(originalIndexPath)
					}
				}
				// Only put the index if it doesn't already exist; otherwise
				// we might trigger an unnecessary event if someone is
				// watching the index
				if c.stm.Get(indexPath) == "" {
					c.stm.Put(indexPath, key, options...)
				}
			}
		}
	}
	bytes, _ := val.Marshal()
	c.stm.Put(c.Path(key), string(bytes), options...)
	return nil
}

func (c *readWriteCollection) Create(key string, val proto.Marshaler) error {
	fullKey := c.Path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.Put(key, val)
	return nil
}

func (c *readWriteCollection) Delete(key string) error {
	fullKey := c.Path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	if c.indexes != nil && c.template != nil {
		val := proto.Clone(c.template)
		for _, index := range c.indexes {
			if err := c.Get(key, val.(proto.Unmarshaler)); err == nil {
				if index.Multi {
					indexPaths := c.getMultiIndexPaths(val, index, key)
					for _, indexPath := range indexPaths {
						c.stm.Del(indexPath)
					}
				} else {
					indexPath := c.getIndexPath(val, index, key)
					c.stm.Del(indexPath)
				}
			}
		}
	}
	c.stm.Del(fullKey)
	return nil
}

func (c *readWriteCollection) DeleteAll() {
	for _, index := range c.indexes {
		// Delete indexes
		indexDir := c.prefix
		indexDir = strings.TrimRight(indexDir, "/")
		c.stm.DelAll(fmt.Sprintf("%s__index_%s/", indexDir, index))
	}
	c.stm.DelAll(c.prefix)
}

type readWriteIntCollection struct {
	*collection
	stm STM
}

func (c *readWriteIntCollection) Create(key string, val int) error {
	fullKey := c.Path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.stm.Put(fullKey, strconv.Itoa(val))
	return nil
}

func (c *readWriteIntCollection) Get(key string) (int, error) {
	valStr := c.stm.Get(c.Path(key))
	if valStr == "" {
		return 0, ErrNotFound{c.prefix, key}
	}
	return strconv.Atoi(valStr)
}

func (c *readWriteIntCollection) Increment(key string) error {
	return c.IncrementBy(key, 1)
}

func (c *readWriteIntCollection) IncrementBy(key string, n int) error {
	fullKey := c.Path(key)
	valStr := c.stm.Get(fullKey)
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return ErrMalformedValue{c.prefix, key, valStr}
	}
	c.stm.Put(fullKey, strconv.Itoa(val+n))
	return nil
}

func (c *readWriteIntCollection) Decrement(key string) error {
	return c.DecrementBy(key, 1)
}

func (c *readWriteIntCollection) DecrementBy(key string, n int) error {
	fullKey := c.Path(key)
	valStr := c.stm.Get(fullKey)
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return ErrMalformedValue{c.prefix, key, valStr}
	}
	c.stm.Put(fullKey, strconv.Itoa(val-n))
	return nil
}

func (c *readWriteIntCollection) Delete(key string) error {
	fullKey := c.Path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	c.stm.Del(fullKey)
	return nil
}

type readonlyCollection struct {
	*collection
	ctx context.Context
}

func (c *readonlyCollection) Get(key string, val proto.Unmarshaler) error {
	resp, err := c.etcdClient.Get(c.ctx, c.Path(key))
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound{c.prefix, key}
	}

	return val.Unmarshal(resp.Kvs[0].Value)
}

// an indirect iterator goes through a list of keys and retrieve those
// items from the collection.
type indirectIterator struct {
	index int
	resp  *etcd.GetResponse
	col   *readonlyCollection
}

func (i *indirectIterator) Next(key *string, val proto.Unmarshaler) (ok bool, retErr error) {
	if i.index < len(i.resp.Kvs) {
		kv := i.resp.Kvs[i.index]
		i.index++

		*key = path.Base(string(kv.Key))
		if err := i.col.Get(*key, val); err != nil {
			return false, err
		}

		return true, nil
	}
	return false, nil
}

func (c *readonlyCollection) GetByIndex(index Index, val interface{}) (Iterator, error) {
	valStr := fmt.Sprintf("%s", val)
	resp, err := c.etcdClient.Get(c.ctx, c.indexDir(index, valStr), etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortDescend))
	if err != nil {
		return nil, err
	}
	return &indirectIterator{
		resp: resp,
		col:  c,
	}, nil
}

// List returns an iteraor that can be used to iterate over the collection.
// The objects are sorted by revision time in descending order, i.e. newer
// objects are returned first.
func (c *readonlyCollection) List() (Iterator, error) {
	resp, err := c.etcdClient.Get(c.ctx, c.prefix, etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortDescend))
	if err != nil {
		return nil, err
	}
	return &iterator{
		resp: resp,
	}, nil
}

type iterator struct {
	index int
	resp  *etcd.GetResponse
}

func (c *readonlyCollection) Count() (int64, error) {
	resp, err := c.etcdClient.Get(c.ctx, c.prefix, etcd.WithPrefix(), etcd.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return resp.Count, err
}

func (i *iterator) Next(key *string, val proto.Unmarshaler) (ok bool, retErr error) {
	if i.index < len(i.resp.Kvs) {
		kv := i.resp.Kvs[i.index]
		i.index++

		*key = path.Base(string(kv.Key))
		if err := val.Unmarshal(kv.Value); err != nil {
			return false, err
		}

		return true, nil
	}
	return false, nil
}

// Watch a collection, returning the current content of the collection as
// well as any future additions.
func (c *readonlyCollection) Watch() (watch.Watcher, error) {
	return watch.NewWatcher(c.ctx, c.etcdClient, c.prefix)
}

func (c *readonlyCollection) WatchWithPrev() (watch.Watcher, error) {
	return watch.NewWatcherWithPrev(c.ctx, c.etcdClient, c.prefix)
}

// WatchByIndex watches items in a collection that match a particular index
func (c *readonlyCollection) WatchByIndex(index Index, val interface{}) (watch.Watcher, error) {
	eventCh := make(chan *watch.Event)
	done := make(chan struct{})
	watcher, err := watch.NewWatcher(c.ctx, c.etcdClient, c.indexDir(index, fmt.Sprintf("%s", val)))
	if err != nil {
		return nil, err
	}
	go func() (retErr error) {
		defer func() {
			if retErr != nil {
				eventCh <- &watch.Event{
					Type: watch.EventError,
					Err:  retErr,
				}
				watcher.Close()
			}
			close(eventCh)
		}()
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
				resp, err := c.etcdClient.Get(c.ctx, c.Path(path.Base(string(ev.Key))))
				if err != nil {
					return err
				}
				if len(resp.Kvs) == 0 {
					// this happens only if the item was deleted shortly after
					// we receive this event.
					continue
				}
				directEv = &watch.Event{
					Key:   []byte(path.Base(string(ev.Key))),
					Value: resp.Kvs[0].Value,
					Type:  ev.Type,
				}
			case watch.EventDelete:
				directEv = &watch.Event{
					Key:  []byte(path.Base(string(ev.Key))),
					Type: ev.Type,
				}
			}
			eventCh <- directEv
		}
	}()
	return watch.MakeWatcher(eventCh, done), nil
}

// WatchOne watches a given item.  The first value returned from the watch
// will be the current value of the item.
func (c *readonlyCollection) WatchOne(key string) (watch.Watcher, error) {
	return watch.NewWatcher(c.ctx, c.etcdClient, c.Path(key))
}
