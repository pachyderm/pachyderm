package collection

import (
	"context"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

type postgresCollection struct {
	table   string
	indexes []*Index
	// keyCheck is a function that checks if a key is valid.  Invalid keys
	// cannot be created.
	keyCheck func(string) error

	// valCheck is a function that checks if a value is valid.
	valCheck func(proto.Message) error
}

// NewPostgresCollection creates a new collection backed by postgres.
func NewPostgresCollection(table string, indexes []*Index, keyCheck func(string) error, valCheck func(proto.Message) error) Collection {
	return &postgresCollection{
		table:    table,
		indexes:  indexes,
		keyCheck: keyCheck,
		valCheck: valCheck,
	}
}

func (c *postgresCollection) ReadWrite(stm STM) ReadWriteCollection {
	return &postgresReadWriteCollection{c}
}

func (c *postgresCollection) ReadOnly(ctx context.Context) ReadOnlyCollection {
	return &postgresReadOnlyCollection{c}
}

func (c *postgresCollection) Claim(ctx context.Context, key string, val proto.Message, f func(context.Context) error) error {
	return errors.New("Claim is not supported on postgres collections")
}

type postgresReadOnlyCollection struct {
	*postgresCollection
}

func (c *postgresReadOnlyCollection) Get(key string, val proto.Message) error {
	return errors.New("Get is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) GetByIndex(index *Index, indexVal interface{}, val proto.Message, opts *Options, f func(key string) error) error {
	return errors.New("GetByIndex is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) GetBlock(key string, val proto.Message) error {
	return errors.New("GetBlock is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) TTL(key string) (int64, error) {
	return 0, errors.New("TTL is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) List(val proto.Message, opts *Options, f func(key string) error) error {
	return errors.New("List is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) ListRev(val proto.Message, opts *Options, f func(key string, createRev int64) error) error {
	return errors.New("ListRev is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) ListPrefix(prefix string, val proto.Message, opts *Options, f func(string) error) error {
	return errors.New("ListPrefix is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) Count() (int64, error) {
	return 0, errors.New("Count is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) Watch(opts ...watch.OpOption) (watch.Watcher, error) {
	return nil, errors.New("Watch is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) WatchF(f func(*watch.Event) error, opts ...watch.OpOption) error {
	return errors.New("WatchF is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) WatchOne(key string, opts ...watch.OpOption) (watch.Watcher, error) {
	return nil, errors.New("WatchOne is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) WatchOneF(key string, f func(*watch.Event) error, opts ...watch.OpOption) error {
	return errors.New("WatchOneF is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) WatchByIndex(index *Index, val interface{}) (watch.Watcher, error) {
	return nil, errors.New("WatchByIndex is not supported on read-only postgres collections")
}

type postgresReadWriteCollection struct {
	*postgresCollection
}

func (c *postgresReadWriteCollection) Get(key string, val proto.Message) error {
	return errors.New("Get is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) Put(key string, val proto.Message) error {
	return errors.New("Put is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) TTL(key string) (int64, error) {
	return 0, errors.New("TTL is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) PutTTL(key string, val proto.Message, ttl int64) error {
	return errors.New("PutTTL is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) Update(key string, val proto.Message, f func() error) error {
	return errors.New("Update is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) Upsert(key string, val proto.Message, f func() error) error {
	return errors.New("Upsert is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) Create(key string, val proto.Message) error {
	return errors.New("Create is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) Delete(key string) error {
	return errors.New("Delete is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) DeleteAll() {
	return
}

func (c *postgresReadWriteCollection) DeleteAllPrefix(prefix string) {
	return
}
