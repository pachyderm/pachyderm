package collection

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/watch"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
)

// Collection implements helper functions that makes common operations
// on top of etcd more pleasant to work with.  It's called collection
// because most of our data is modelled as collections, such as repos,
// commits, branches, etc.

type PostgresCollection interface {
	// ReadWrite enables reads and writes on a collection in a
	// transactional manner.  Specifically, all writes are applied
	// atomically, and writes are only applied if reads have not been
	// invalidated at the end of the transaction.  Basically, it's
	// software transactional memory.  See this blog post for details:
	// https://coreos.com/blog/transactional-memory-with-etcd3.html
	ReadWrite(tx *sqlx.Tx) PostgresReadWriteCollection

	// For read-only operations, use the ReadOnly for better performance
	ReadOnly(ctx context.Context) PostgresReadOnlyCollection

	// With returns a new collection with the given predicate
	With(index *Index, val string) PostgresCollection
}

type EtcdCollection interface {
	// ReadWrite enables reads and writes on a collection in a
	// transactional manner.  Specifically, all writes are applied
	// atomically, and writes are only applied if reads have not been
	// invalidated at the end of the transaction.  Basically, it's
	// software transactional memory.  See this blog post for details:
	// https://coreos.com/blog/transactional-memory-with-etcd3.html
	ReadWrite(stm STM) EtcdReadWriteCollection

	// For read-only operations, use the ReadOnly for better performance
	ReadOnly(ctx context.Context) EtcdReadOnlyCollection

	// Claim attempts to claim a key and run the passed in callback with
	// the context for the claim.
	Claim(ctx context.Context, key string, val proto.Message, f func(context.Context) error) error
}

// Index specifies a secondary index on a collection.
//
// Indexes are created in a transactional manner thanks to etcd's
// transactional support.
//
// A secondary index for collection "foo" on field "bar" will reside under
// the path `/foo__index_bar`.  Each item under the path is in turn a
// directory whose name is the value of the field `bar`.  For instance,
// if you have a object in collection `foo` whose `bar` field is `test`,
// then you will see a directory at path `/foo__index_bar/test`.
//
// Under that directory, you have keys that point to items in the collection.
// For instance, if the aforementioned object has the key "buzz", then you
// will see an item at `/foo__index_bar/test/buzz`.  The value of this item
// is empty.  Thus, to get all items in collection `foo` whose values of
// field `bar` is `test`, we issue a query for all items under
// `foo__index_bar/test`.
type Index struct {
	Name    string
	Extract func(val proto.Message) string

	// `limit` is an internal implementation detail for etcd collections to avoid list operations overflowing the max message size
	limit int64
}

// ReadWriteCollection is a collection interface that supports read,write and delete
// operations.
type ReadWriteCollection interface {
	Get(key string, val proto.Message) error
	Put(key string, val proto.Message) error
	// Update reads the current value associated with 'key', calls 'f' to update
	// the value, and writes the new value back to the collection. 'key' must be
	// present in the collection, or a 'Not Found' error is returned
	Update(key string, val proto.Message, f func() error) error
	// Upsert is like Update but 'key' is not required to be present
	Upsert(key string, val proto.Message, f func() error) error
	Create(key string, val proto.Message) error
	Delete(key string) error
	DeleteAll() error
}

type PostgresReadWriteCollection interface {
	ReadWriteCollection

	// Unsupported operations - only here during migration so we can compile
	// TODO: remove these before merging into master
	TTL(key string) (int64, error)
	PutTTL(key string, val proto.Message, ttl int64) error
}

type EtcdReadWriteCollection interface {
	ReadWriteCollection

	// TTL returns the amount of time that 'key' will continue to exist in the
	// collection, or '0' if 'key' will remain in the collection indefinitely
	TTL(key string) (int64, error)
	// PutTTL is the same as Put except that the object is removed after
	// TTL seconds.
	// WARNING: using PutTTL with a collection that has secondary indices
	// can result in inconsistency, as the indices are removed at roughly
	// but not exactly the same time as the documents.
	PutTTL(key string, val proto.Message, ttl int64) error

	DeleteAllPrefix(prefix string) error
}

// ReadOnlyCollection is a collection interface that only supports read ops.
type ReadOnlyCollection interface {
	Get(key string, val proto.Message) error
	GetByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func() error) error
	List(val proto.Message, opts *Options, f func() error) error
	Count() (int64, error)
	Watch(opts ...watch.Option) (watch.Watcher, error)
	WatchF(f func(*watch.Event) error, opts ...watch.Option) error
	WatchOne(key string, opts ...watch.Option) (watch.Watcher, error)
	WatchOneF(key string, f func(*watch.Event) error, opts ...watch.Option) error
}

type PostgresReadOnlyCollection interface {
	ReadOnlyCollection

	// Unsupported operation - only here during migration so we can compile
	// TODO: remove this before merging into master
	TTL(key string) (int64, error)
}

type EtcdReadOnlyCollection interface {
	ReadOnlyCollection

	// TTL returns the number of seconds that 'key' will continue to exist in the
	// collection, or '0' if 'key' will remain in the collection indefinitely
	TTL(key string) (int64, error)

	ListRev(val proto.Message, opts *Options, f func(createRev int64) error) error

	WatchByIndex(index *Index, val string) (watch.Watcher, error)
}
