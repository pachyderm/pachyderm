package collection

import (
	"context"
	"fmt"
	"path"
	"strconv"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

type collection struct {
	etcdClient *etcd.Client
	prefix     string
}

func NewCollection(etcdClient *etcd.Client, prefix string) Collection {
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

// path returns the full path of a key in the etcd namespace
func (c *collection) path(key string) string {
	return path.Join(c.prefix, key)
}

type readWriteCollection struct {
	*collection
	stm STM
}

func (c *readWriteCollection) Get(key string, val proto.Message) error {
	valStr := c.stm.Get(c.path(key))
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	return proto.UnmarshalText(valStr, val)
}

func (c *readWriteCollection) Put(key string, val proto.Message) {
	c.stm.Put(c.path(key), val.String())
}

func (c *readWriteCollection) Create(key string, val proto.Message) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.Put(key, val)
	return nil
}

func (c *readWriteCollection) Delete(key string) error {
	fullKey := c.path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	c.stm.Del(fullKey)
	return nil
}

func (c *readWriteCollection) DeleteAll() {
	c.stm.DelAll(c.prefix)
}

type readWriteIntCollection struct {
	*collection
	stm STM
}

func (c *readWriteIntCollection) Create(key string, val int) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.stm.Put(fullKey, strconv.Itoa(val))
	return nil
}

func (c *readWriteIntCollection) Get(key string) (int, error) {
	valStr := c.stm.Get(c.path(key))
	if valStr == "" {
		return 0, ErrNotFound{c.prefix, key}
	}
	return strconv.Atoi(valStr)
}

func (c *readWriteIntCollection) Increment(key string) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return ErrMalformedValue{c.prefix, key, valStr}
	}
	c.stm.Put(fullKey, strconv.Itoa(val+1))
	return nil
}

func (c *readWriteIntCollection) Decrement(key string) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return ErrMalformedValue{c.prefix, key, valStr}
	}
	c.stm.Put(fullKey, strconv.Itoa(val-1))
	return nil
}

func (c *readWriteIntCollection) Delete(key string) error {
	fullKey := c.path(key)
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

func (c *readonlyCollection) Get(key string, val proto.Message) error {
	resp, err := c.etcdClient.Get(c.ctx, c.path(key))
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound{c.prefix, key}
	}

	return proto.UnmarshalText(string(resp.Kvs[0].Value), val)
}

type iterator struct {
	index int
	resp  *etcd.GetResponse
}

func (i *iterator) Next(key *string, val proto.Message) (ok bool, retErr error) {
	if i.index < len(i.resp.Kvs) {
		kv := i.resp.Kvs[i.index]
		i.index += 1

		*key = path.Base(string(kv.Key))
		if err := proto.UnmarshalText(string(kv.Value), val); err != nil {
			return false, err
		}

		return true, nil
	}
	return false, nil
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

type event struct {
	key   []byte
	value []byte
	typ   EventType
}

func (e event) Unmarshal(key *string, value proto.Message) error {
	*key = path.Base(string(e.key))
	return proto.UnmarshalText(string(e.value), value)
}

func (e event) Type() EventType {
	return e.typ
}

type watcher struct {
	events  []event
	watcher etcd.Watcher
	rch     etcd.WatchChan
}

func (w *watcher) Next() (Event, error) {
	if len(w.events) == 0 {
		ev, ok := <-w.rch
		if !ok {
			return nil, fmt.Errorf("stream has been closed")
		}

		for _, etcdEv := range ev.Events {
			ev := event{
				key:   etcdEv.Kv.Key,
				value: etcdEv.Kv.Value,
			}
			switch etcdEv.Type {
			case etcd.EventTypePut:
				ev.typ = EventPut
			case etcd.EventTypeDelete:
				ev.typ = EventDelete
			}
			w.events = append(w.events, ev)
		}
	}

	// pop the first element
	event := w.events[0]
	w.events = w.events[1:]

	return event, nil
}

func (w *watcher) Close() error {
	return w.watcher.Close()
}

// Watch a collection, returning the current content of the collection as
// well as any future additions.
// TODO: handle deletion events; right now if an item is deleted from
// this collection, we treat the event as if it's an addition event.
func (c *readonlyCollection) Watch() (Watcher, error) {
	// Firstly we list the collection to get the current items
	// Sort them by ascending order because that's how the items would have
	// been returned if we watched them from the beginning.
	resp, err := c.etcdClient.Get(c.ctx, c.path(""), etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortAscend))
	if err != nil {
		return nil, err
	}

	etcdWatcher := etcd.NewWatcher(c.etcdClient)
	// Now we issue a watch that uses the revision timestamp returned by the
	// Get request earlier.  That way even if some items are added between
	// when we list the collection and when we start watching the collection,
	// we won't miss any items.
	rch := etcdWatcher.Watch(c.ctx, c.path(""), etcd.WithPrefix(), etcd.WithRev(resp.Header.Revision))
	w := &watcher{
		watcher: etcdWatcher,
		rch:     rch,
	}

	for _, etcdKv := range resp.Kvs {
		w.events = append(w.events, event{
			key:   etcdKv.Key,
			value: etcdKv.Value,
			typ:   EventPut,
		})
	}
	return w, nil
}

// WatchOne watches a given item.  The first value returned from the watch
// will be the current value of the item.
func (c *readonlyCollection) WatchOne(key string) (Watcher, error) {
	// Firstly we list the collection to get the current items
	resp, err := c.etcdClient.Get(c.ctx, c.path(key))
	if err != nil {
		return nil, err
	}

	etcdWatcher := etcd.NewWatcher(c.etcdClient)
	rch := etcdWatcher.Watch(c.ctx, c.path(key), etcd.WithRev(resp.Header.Revision))
	w := &watcher{
		watcher: etcdWatcher,
		rch:     rch,
	}
	if resp.Count > 0 {
		kv := resp.Kvs[0]
		w.events = append(w.events, event{
			key:   kv.Key,
			value: kv.Value,
			typ:   EventPut,
		})
	}
	return w, nil
}
