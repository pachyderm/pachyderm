package drive

import (
	"context"
	"fmt"
	"io"
	"path"
	"strconv"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

const (
	reposPrefix         = "/repos"
	repoRefCountsPrefix = "/repoRefCounts"
	commitsPrefix       = "/commits"
	branchesPrefix      = "/branches"
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

// collection implements helper functions that makes common operations
// on top of etcd more pleasant to work with.  It's called collection
// because most of our data is modelled as collections, such as repos,
// commits, branches, etc.
type collection struct {
	etcdClient *etcd.Client
	prefix     string
	stm        STM
}

type intCollection struct {
	etcdClient *etcd.Client
	prefix     string
	stm        STM
}

// collectionFactory generates collections.  It's mainly used for
// namespaced collections, such as /commits/foo, i.e. commits in
// repo foo.
type collectionFactory func(string) *collection

// repos returns a collection of repos
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /repos
//     /foo
//     /bar
func (d *driver) repos(stm STM) *collection {
	return &collection{
		prefix:     path.Join(d.prefix, reposPrefix),
		etcdClient: d.etcdClient,
		stm:        stm,
	}
}

// repoRefCounts returns a collection of repo reference counters
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /repoRefCounts
//     /foo
//     /bar
func (d *driver) repoRefCounts(stm STM) *intCollection {
	return &intCollection{
		prefix:     path.Join(d.prefix, repoRefCountsPrefix),
		etcdClient: d.etcdClient,
		stm:        stm,
	}
}

// commits returns a collection of commits
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /commits
//     /foo
//       /UUID1
//       /UUID2
//     /bar
//       /UUID3
//       /UUID4
func (d *driver) commits(stm STM) collectionFactory {
	return func(repo string) *collection {
		return &collection{
			prefix:     path.Join(d.prefix, commitsPrefix, repo),
			etcdClient: d.etcdClient,
			stm:        stm,
		}
	}
}

// commits returns a collection of commits
// Example etcd structure, assuming we have two repos "foo" and "bar",
// each of which has two branches:
//   /branches
//     /foo
//       /master
//       /test
//     /bar
//       /master
//       /test
func (d *driver) branches(stm STM) collectionFactory {
	return func(repo string) *collection {
		return &collection{
			prefix:     path.Join(d.prefix, branchesPrefix, repo),
			etcdClient: d.etcdClient,
			stm:        stm,
		}
	}
}

// path returns the full path of a key in the etcd namespace
func (c *collection) path(key string) string {
	return path.Join(c.prefix, key)
}

func (c *collection) Get(key string, val proto.Message) error {
	valStr := c.stm.Get(c.path(key))
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	return proto.UnmarshalText(valStr, val)
}

func (c *collection) Put(key string, val proto.Message) {
	c.stm.Put(c.path(key), val.String())
}

func (c *collection) Create(key string, val proto.Message) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.stm.Put(fullKey, val.String())
	return nil
}

func (c *collection) Delete(key string) error {
	fullKey := c.path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	c.stm.Del(fullKey)
	return nil
}

func (c *collection) DeleteAll() {
	c.stm.DelAll(c.prefix)
}

// itoa converts an integer to a string
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

// path returns the full path of a key in the etcd namespace
func (c *intCollection) path(key string) string {
	return path.Join(c.prefix, key)
}

// Create creates an object if it doesn't already exist
func (c *intCollection) Create(key string, val int) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr != "" {
		return ErrExists{c.prefix, key}
	}
	c.stm.Put(fullKey, itoa(val))
	return nil
}

// Get gets the value of a key
func (c *intCollection) Get(key string) (int, error) {
	valStr := c.stm.Get(c.path(key))
	if valStr == "" {
		return 0, ErrNotFound{c.prefix, key}
	}
	return strconv.Atoi(valStr)
}

// Increment atomically increments the value of a key.
func (c *intCollection) Increment(key string) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return ErrMalformedValue{c.prefix, key, valStr}
	}
	c.stm.Put(fullKey, itoa(val+1))
	return nil
}

// Decrement atomically decrements the value of a key.
func (c *intCollection) Decrement(key string) error {
	fullKey := c.path(key)
	valStr := c.stm.Get(fullKey)
	if valStr == "" {
		return ErrNotFound{c.prefix, key}
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return ErrMalformedValue{c.prefix, key, valStr}
	}
	c.stm.Put(fullKey, itoa(val-1))
	return nil
}

// Delete deletes an object
func (c *intCollection) Delete(key string) error {
	fullKey := c.path(key)
	if c.stm.Get(fullKey) == "" {
		return ErrNotFound{c.prefix, key}
	}
	c.stm.Del(fullKey)
	return nil
}

// Following are read-only collections
type readonlyCollection struct {
	ctx        context.Context
	etcdClient *etcd.Client
	prefix     string
}

type readonlyCollectionFactory func(string) *readonlyCollection

func (d *driver) reposReadonly(ctx context.Context) *readonlyCollection {
	return &readonlyCollection{
		ctx:        ctx,
		prefix:     path.Join(d.prefix, reposPrefix),
		etcdClient: d.etcdClient,
	}
}

func (d *driver) commitsReadonly(ctx context.Context) readonlyCollectionFactory {
	return func(repo string) *readonlyCollection {
		return &readonlyCollection{
			ctx:        ctx,
			prefix:     path.Join(d.prefix, commitsPrefix, repo),
			etcdClient: d.etcdClient,
		}
	}
}

func (d *driver) branchesReadonly(ctx context.Context) readonlyCollectionFactory {
	return func(repo string) *readonlyCollection {
		return &readonlyCollection{
			ctx:        ctx,
			prefix:     path.Join(d.prefix, branchesPrefix, repo),
			etcdClient: d.etcdClient,
		}
	}
}

// path returns the full path of a key in the etcd namespace
func (c *readonlyCollection) path(key string) string {
	return path.Join(c.prefix, key)
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
func (c *readonlyCollection) List() (Iterator, error) {
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
func (c *readonlyCollection) Watch() (IterateCloser, error) {
	watcher := etcd.NewWatcher(c.etcdClient)
	rch := watcher.Watch(c.ctx, c.path(""), etcd.WithPrefix(), etcd.WithRev(1))
	return &iterateCloser{
		watcher: watcher,
		rch:     rch,
	}, nil
}

// WatchOne watches for the new values of a certain item
func (c *readonlyCollection) WatchOne(key string) (IterateCloser, error) {
	watcher := etcd.NewWatcher(c.etcdClient)
	rch := watcher.Watch(c.ctx, c.path(key))
	return &iterateCloser{
		watcher: watcher,
		rch:     rch,
	}, nil
}
