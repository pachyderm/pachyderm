package collection

import (
	"context"
	"fmt"
	"io"
	"path"
	"strconv"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

type ErrNotFound struct {
	Type string
	Key  string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s %s not found", e.Type, e.Key)
}

type ErrExists struct {
	Type string
	Key  string
}

func (e ErrExists) Error() string {
	return fmt.Sprintf("%s %s already exists", e.Type, e.Key)
}

type ErrMalformedValue struct {
	Type string
	Key  string
	Val  string
}

func (e ErrMalformedValue) Error() string {
	return fmt.Sprintf("malformed value at %s/%s: %s", e.Type, e.Key, e.Val)
}

// Collection implements helper functions that makes common operations
// on top of etcd more pleasant to work with.  It's called collection
// because most of our data is modelled as collections, such as repos,
// commits, branches, etc.
type Collection struct {
	etcdClient *etcd.Client
	prefix     string
	stm        STM
}

func NewCollection(etcdClient *etcd.Client, prefix string, stm STM) *Collection {
	return &Collection{
		prefix:     prefix,
		etcdClient: etcdClient,
		stm:        stm,
	}
}

type IntCollection struct {
	etcdClient *etcd.Client
	prefix     string
	stm        STM
}

func NewIntCollection(etcdClient *etcd.Client, prefix string, stm STM) *IntCollection {
	return &IntCollection{
		prefix:     prefix,
		etcdClient: etcdClient,
		stm:        stm,
	}
}

// CollectionFactory generates collections.  It's mainly used for
// namespaced collections, such as /commits/foo, i.e. commits in
// repo foo.
type CollectionFactory func(string) *Collection

// path returns the full path of a key in the etcd namespace
func (c *Collection) path(key string) string {
	return path.Join(c.prefix, key)
}

func (c *Collection) Get(key string, val proto.Message) error {
	valStr := c.stm.Get(c.path(key))
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	return proto.UnmarshalText(valStr, val)
}

func (c *Collection) Put(key string, val proto.Message) {
	c.stm.Put(c.path(key), val.String())
}

func (c *Collection) Create(key string, val proto.Message) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.stm.Put(fullKey, val.String())
	return nil
}

func (c *Collection) Delete(key string) error {
	fullKey := c.path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	c.stm.Del(fullKey)
	return nil
}

func (c *Collection) DeleteAll() {
	c.stm.DelAll(c.prefix)
}

// path returns the full path of a key in the etcd namespace
func (c *IntCollection) path(key string) string {
	return path.Join(c.prefix, key)
}

// Create creates an object if it doesn't already exist
func (c *IntCollection) Create(key string, val int) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.stm.Put(fullKey, strconv.Itoa(val))
	return nil
}

// Get gets the value of a key
func (c *IntCollection) Get(key string) (int, error) {
	valStr := c.stm.Get(c.path(key))
	if valStr == "" {
		return 0, ErrNotFound{c.prefix, key}
	}
	return strconv.Atoi(valStr)
}

// Increment atomically increments the value of a key.
func (c *IntCollection) Increment(key string) error {
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

// Decrement atomically decrements the value of a key.
func (c *IntCollection) Decrement(key string) error {
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

// Delete deletes an object
func (c *IntCollection) Delete(key string) error {
	fullKey := c.path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	c.stm.Del(fullKey)
	return nil
}

// Following are read-only collections
type ReadonlyCollection struct {
	ctx        context.Context
	etcdClient *etcd.Client
	prefix     string
}

func NewReadonlyCollection(ctx context.Context, etcdClient *etcd.Client, prefix string) *ReadonlyCollection {
	return &ReadonlyCollection{
		ctx:        ctx,
		prefix:     prefix,
		etcdClient: etcdClient,
	}
}

type ReadonlyCollectionFactory func(string) *ReadonlyCollection

// path returns the full path of a key in the etcd namespace
func (c *ReadonlyCollection) path(key string) string {
	return path.Join(c.prefix, key)
}

func (c *ReadonlyCollection) Get(key string, val proto.Message) error {
	resp, err := c.etcdClient.Get(c.ctx, c.path(key))
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		return ErrNotFound{c.prefix, key}
	}

	return proto.UnmarshalText(string(resp.Kvs[0].Value), val)
}

type Iterator interface {
	// Next is a function that, when called, serializes the key and value
	// of the next object in a collection.
	// ok is true if the serialization was successful.  It's false if the
	// collection has been exhausted.
	Next(key *string, val proto.Message) (ok bool, retErr error)
}

type iterator struct {
	index int
	resp  *etcd.GetResponse
}

func (i *iterator) Next(key *string, val proto.Message) (ok bool, retErr error) {
	if i.index < len(i.resp.Kvs) {
		kv := i.resp.Kvs[i.index]
		i.index += 1

		*key = string(kv.Key)
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
func (c *ReadonlyCollection) List() (Iterator, error) {
	resp, err := c.etcdClient.Get(c.ctx, c.path(""), etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortDescend))
	if err != nil {
		return nil, err
	}
	return &iterator{
		resp: resp,
	}, nil
}

type IterateCloser interface {
	Iterator
	io.Closer
}

type iterateCloser struct {
	events  []*etcd.Event
	watcher etcd.Watcher
	rch     etcd.WatchChan
}

func (i *iterateCloser) Next(key *string, val proto.Message) (ok bool, retErr error) {
	if len(i.events) == 0 {
		ev, ok := <-i.rch
		if !ok {
			return false, nil
		}

		i.events = append(i.events, ev.Events...)
	}

	kv := i.events[0].Kv
	i.events = i.events[1:]

	*key = string(kv.Key)
	if err := proto.UnmarshalText(string(kv.Value), val); err != nil {
		return false, err
	}

	return true, nil
}

func (i *iterateCloser) Close() error {
	return i.watcher.Close()
}

// Watch watches new items added to this collection
// TODO: handle deletion events; right now if an item is deleted from
// this collection, we treat the event as if it's an addition event.
func (c *ReadonlyCollection) Watch() (IterateCloser, error) {
	watcher := etcd.NewWatcher(c.etcdClient)
	rch := watcher.Watch(c.ctx, c.path(""), etcd.WithPrefix(), etcd.WithRev(1))
	return &iterateCloser{
		watcher: watcher,
		rch:     rch,
	}, nil
}

// WatchOne watches for the new values of a certain item
func (c *ReadonlyCollection) WatchOne(key string) (IterateCloser, error) {
	watcher := etcd.NewWatcher(c.etcdClient)
	rch := watcher.Watch(c.ctx, c.path(key))
	return &iterateCloser{
		watcher: watcher,
		rch:     rch,
	}, nil
}
