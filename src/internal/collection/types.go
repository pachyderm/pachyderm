package collection

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"

	"github.com/gogo/protobuf/proto"
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
	ReadWrite(tx *pachsql.Tx) PostgresReadWriteCollection

	// For read-only operations, use the ReadOnly for better performance
	ReadOnly(ctx context.Context) PostgresReadOnlyCollection
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

	WithRenewer(ctx context.Context, cb func(context.Context, *Renewer) error) error
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
	Get(key interface{}, val proto.Message) error
	Put(key interface{}, val proto.Message) error
	// Update reads the current value associated with 'key', calls 'f' to update
	// the value, and writes the new value back to the collection. 'key' must be
	// present in the collection, or a 'Not Found' error is returned
	Update(key interface{}, val proto.Message, f func() error) error
	// Upsert is like Update but 'key' is not required to be present
	Upsert(key interface{}, val proto.Message, f func() error) error
	Create(key interface{}, val proto.Message) error
	Delete(key interface{}) error
	DeleteAll() error
}

type PostgresReadWriteCollection interface {
	ReadWriteCollection

	DeleteByIndex(index *Index, indexVal string) error
	// GetByIndex can have a large impact on database contention if used to retrieve
	// a large number of rows. Consider using a read-only collection if possible
	GetByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(string) error) error

	// GetUniqueByIndex is identical to GetByIndex except it is an error if
	// exactly one row is not found.
	// TODO: decide if we should merge this with GetByIndex and use an `Options`.
	GetUniqueByIndex(index *Index, indexVal string, val proto.Message) error
	// NOTE: List scans the collection over multiple queries,
	// making this method susceptible to inconsistent reads
	List(val proto.Message, opts *Options, f func(string) error) error
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
	Get(key interface{}, val proto.Message) error
	GetByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(string) error) error
	List(val proto.Message, opts *Options, f func(string) error) error
	ListRev(val proto.Message, opts *Options, f func(string, int64) error) error
	Count() (int64, error)
	Watch(opts ...watch.Option) (watch.Watcher, error)
	WatchF(f func(*watch.Event) error, opts ...watch.Option) error
	WatchOne(key interface{}, opts ...watch.Option) (watch.Watcher, error)
	WatchOneF(key interface{}, f func(*watch.Event) error, opts ...watch.Option) error
	WatchByIndex(index *Index, val string, opts ...watch.Option) (watch.Watcher, error)
	WatchByIndexF(index *Index, val string, f func(*watch.Event) error, opts ...watch.Option) error
}

type PostgresReadOnlyCollection interface {
	ReadOnlyCollection

	GetRevByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(string, int64) error) error

	// GetUniqueByIndex is identical to GetByIndex except it is an error if
	// exactly one row is not found.
	// TODO: decide if we should merge this with GetByIndex and use an `Options`.
	GetUniqueByIndex(index *Index, indexVal string, val proto.Message) error
}

type EtcdReadOnlyCollection interface {
	ReadOnlyCollection

	// TTL returns the number of seconds that 'key' will continue to exist in the
	// collection, or '0' if 'key' will remain in the collection indefinitely
	// TODO: TTL might be unused
	TTL(key string) (int64, error)
	// CountRev returns the number of items in the collection at a specific
	// revision, it's only in EtcdReadOnlyCollection because only etcd has
	// revs.
	// TODO: CountRev might be unused
	CountRev(int64) (int64, int64, error)
}
