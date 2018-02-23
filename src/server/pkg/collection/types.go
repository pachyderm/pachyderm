package collection

import (
	"context"

	"github.com/pachyderm/pachyderm/src/server/pkg/watch"

	"github.com/gogo/protobuf/proto"
)

// Collection implements helper functions that makes common operations
// on top of etcd more pleasant to work with.  It's called collection
// because most of our data is modelled as collections, such as repos,
// commits, branches, etc.
type Collection interface {
	// Path returns the full etcd path of the given key in the collection
	Path(string) string
	// ReadWrite enables reads and writes on a collection in a
	// transactional manner.  Specifically, all writes are applied
	// atomically, and writes are only applied if reads have not been
	// invalidated at the end of the transaction.  Basically, it's
	// software transactional memory.  See this blog post for details:
	// https://coreos.com/blog/transactional-memory-with-etcd3.html
	ReadWrite(stm STM) ReadWriteCollection
	// ReadWriteInt is the same as ReadWrite except that it operates on
	// integral items, as opposed to protobuf items
	ReadWriteInt(stm STM) ReadWriteIntCollection
	// For read-only operatons, use the ReadOnly for better performance
	ReadOnly(ctx context.Context) ReadonlyCollection
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
//
// Multi specifies whether this is a multi-index.  A multi-index is an index
// on a field that's a slice.  The item is then indexed on each element of
// the slice.
type Index struct {
	Field string
	Multi bool
}

// ReadWriteCollection is a collection interface that supports read,write and delete
// operations.
type ReadWriteCollection interface {
	Get(key string, val proto.Message) error
	Put(key string, val proto.Message) error
	// PutTTL is the same as Put except that the object is removed after
	// TTL seconds.
	// WARNING: using PutTTL with a collection that has secondary indices
	// can result in inconsistency, as the indices are removed at roughly
	// but not exactly the same time as the documents.
	PutTTL(key string, val proto.Message, ttl int64) error
	// Update reads the current value associated with 'key', calls 'f' to update
	// the value, and writes the new value back to the collection. 'key' must be
	// present in the collection, or a 'Not Found' error is returned
	Update(key string, val proto.Message, f func() error) error
	// Upsert is like Update but 'key' is not required to be present
	Upsert(key string, val proto.Message, f func() error) error
	Create(key string, val proto.Message) error
	Delete(key string) error
	DeleteAll()
	DeleteAllPrefix(prefix string)
}

// ReadWriteIntCollection is a ReadonlyCollection interface specifically for ints.
type ReadWriteIntCollection interface {
	Create(key string, val int) error
	Get(key string) (int, error)
	Increment(key string) error
	IncrementBy(key string, n int) error
	Decrement(key string) error
	DecrementBy(key string, n int) error
	Delete(key string) error
}

// ReadonlyCollection is a collection interface that only supports read ops.
type ReadonlyCollection interface {
	Get(key string, val proto.Message) error
	GetByIndex(index Index, val interface{}) (Iterator, error)
	// GetBlock is like Get but waits for the key to exist if it doesn't already.
	GetBlock(key string, val proto.Message) error
	List() (Iterator, error)
	ListPrefix(prefix string) (Iterator, error)
	Count() (int64, error)
	Watch() (watch.Watcher, error)
	// WatchWithPrev is like Watch, but the events will include the previous
	// versions of the key/value.
	WatchWithPrev() (watch.Watcher, error)
	WatchOne(key string) (watch.Watcher, error)
	WatchByIndex(index Index, val interface{}) (watch.Watcher, error)
}

// Iterator is an interface for iterating protobufs.
type Iterator interface {
	// Next is a function that, when called, serializes the key and value
	// of the next object in a collection.
	// ok is true if the serialization was successful.  It's false if the
	// collection has been exhausted.
	Next(key *string, val proto.Message) (ok bool, retErr error)
}
