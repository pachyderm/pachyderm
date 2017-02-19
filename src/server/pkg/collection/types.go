package collection

import (
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
)

type Index string

// Collection implements helper functions that makes common operations
// on top of etcd more pleasant to work with.  It's called collection
// because most of our data is modelled as collections, such as repos,
// commits, branches, etc.
type Collection interface {
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

type ReadWriteCollection interface {
	Get(key string, val proto.Message) error
	Put(key string, val proto.Message)
	Create(key string, val proto.Message) error
	Delete(key string) error
	DeleteAll()
}

type ReadWriteIntCollection interface {
	Create(key string, val int) error
	Get(key string) (int, error)
	Increment(key string) error
	Decrement(key string) error
	Delete(key string) error
}

type ReadonlyCollection interface {
	Get(key string, val proto.Message) error
	List() (Iterator, error)
	Watch() (Watcher, error)
	WatchOne(key string) (Watcher, error)
}

type Iterator interface {
	// Next is a function that, when called, serializes the key and value
	// of the next object in a collection.
	// ok is true if the serialization was successful.  It's false if the
	// collection has been exhausted.
	Next(key *string, val proto.Message) (ok bool, retErr error)
}

type EventType int

const (
	EventPut EventType = iota
	EventDelete
)

type Watcher interface {
	io.Closer
	Next() (Event, error)
}

type Event interface {
	Unmarshal(key *string, value proto.Message) error
	Type() EventType
}
