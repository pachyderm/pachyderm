package drive

import (
	"fmt"
	"path"
	"strconv"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
)

const (
	reposPrefix         = "/repos"
	repoRefCountsPrefix = "/repoRefCounts"
	commitsPrefix       = "/commits"
	refsPrefix          = "/refs"
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
// commits, refs, etc.
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
// each of which has two refs:
//   /refs
//     /foo
//       /master
//       /test
//     /bar
//       /master
//       /test
func (d *driver) refs(stm STM) collectionFactory {
	return func(repo string) *collection {
		return &collection{
			prefix:     path.Join(d.prefix, refsPrefix, repo),
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

// iterate is a function that, when called, serializes the key and value
// of the next object in a collection.
// ok is true if the serialization was successful.  It's false if the
// collection has been exhausted.
type iterate func(key *string, val proto.Message) (ok bool, retErr error)

// List returns an iterate function that can be used to iterate over the
// collection.
func (c *collection) List() (iterate, error) {
	resp, err := c.etcdClient.Get(c.stm.Context(), c.path(""), etcd.WithPrefix(), etcd.WithSort(etcd.SortByModRevision, etcd.SortDescend))
	if err != nil {
		return nil, err
	}

	var i int
	return func(key *string, val proto.Message) (bool, error) {
		if i < len(resp.Kvs) {
			kv := resp.Kvs[i]
			i += 1

			*key = string(kv.Key)
			if err := proto.UnmarshalText(string(kv.Value), val); err != nil {
				return false, err
			}

			return true, nil
		}
		return false, nil
	}, nil
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
