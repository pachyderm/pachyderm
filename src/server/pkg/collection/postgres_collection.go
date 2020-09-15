package collection

import (
	"context"
	"fmt"
	"reflect"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/jinzhu/gorm"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

// PostgresModel is the interface that all models must fulfill to be used in a postgres collection
type PostgresModel interface {
	TableName() string
	WriteToProtobuf(proto.Message) error
}

type postgresCollection struct {
	db         *gorm.DB
	model      PostgresModel
	pkeyField  string
	withFields map[string]interface{}
}

func getPrimaryKeyField(model interface{}) string {
	mt := reflect.TypeOf(model)
	// Iterate over all available fields and read the tag value
	for i := 0; i < mt.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := mt.Field(i)

		// Get the field tag value
		tag := field.Tag.Get("gorm")

		fmt.Printf("%d. %v (%v), tag: '%v'\n", i+1, field.Name, field.Type.Name(), tag)
	}
	return ""
}

// NewPostgresCollection creates a new collection backed by postgres.
func NewPostgresCollection(db *gorm.DB, model PostgresModel) Collection {
	return &postgresCollection{
		db:         db,
		model:      model,
		pkeyField:  getPrimaryKeyField(model),
		withFields: make(map[string]interface{}),
	}
}

func (c *postgresCollection) newQuery() *gorm.DB {
	query := c.db.Model(c.model)
	for field, value := range c.withFields {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}
	return query
}

func (c *postgresCollection) With(field string, value interface{}) Collection {
	newWithFields := make(map[string]interface{})
	for k, v := range c.withFields {
		newWithFields[k] = v
	}

	return &postgresCollection{
		db:         c.db,
		model:      c.model,
		pkeyField:  c.pkeyField,
		withFields: newWithFields,
	}
}

func (c *postgresCollection) ReadOnly(ctx context.Context) ReadOnlyCollection {
	return &postgresReadOnlyCollection{c, ctx}
}

func (c *postgresCollection) ReadWrite(stm STM) ReadWriteCollection {
	return &postgresReadWriteCollection{c, stm}
}

func (c *postgresCollection) Claim(ctx context.Context, key string, val proto.Message, f func(context.Context) error) error {
	return errors.New("Claim is not supported on postgres collections")
}

type postgresReadOnlyCollection struct {
	*postgresCollection
	ctx context.Context
}

func writeToProtobuf(result reflect.Value, val proto.Message) error {
	writeResults := result.MethodByName("WriteToProtobuf").Call([]reflect.Value{reflect.ValueOf(val)})
	if !writeResults[0].IsNil() {
		return writeResults[0].Interface().(error)
	}
	return nil
}

func (c *postgresReadOnlyCollection) Get(key string, val proto.Message) error {
	result := reflect.New(reflect.TypeOf(c.model))
	if err := c.newQuery().Where(fmt.Sprintf("%s = ?", c.pkeyField), key).First(&result).Error; err != nil {
		return err
	}
	return writeToProtobuf(result, val)
}

func (c *postgresReadOnlyCollection) GetByIndex(index *Index, indexVal interface{}, val proto.Message, opts *Options, f func(key string) error) error {
	return c.With(index.Field, indexVal).ReadOnly(c.ctx).List(val, opts, f)
}

func orderToSQL(order etcd.SortOrder) (string, error) {
	switch order {
	case etcd.SortAscend:
		return "asc", nil
	case etcd.SortDescend:
		return "desc", nil
	}
	return "", errors.Errorf("unsupported sort order: %d", order)
}

func targetToSQL(target etcd.SortTarget) (string, error) {
	switch target {
	case etcd.SortByCreateRevision:
		return "created_at", nil
	case etcd.SortByModRevision:
		return "updated_at", nil
	}
	return "", errors.Errorf("unsupported sort target: %d", target)
}

func (c *postgresReadOnlyCollection) List(val proto.Message, opts *Options, f func(key string) error) error {
	return c.listInternal(c.newQuery(), val, opts, f)
}

func (c *postgresReadOnlyCollection) ListPrefix(prefix string, val proto.Message, opts *Options, f func(string) error) error {
	query := c.newQuery()
	query = query.Where(fmt.Sprintf("%s LIKE ?", c.pkeyField), prefix+"%")
	return c.listInternal(query, val, opts, f)
}

func (c *postgresReadOnlyCollection) listInternal(query *gorm.DB, val proto.Message, opts *Options, f func(key string) error) error {
	if opts.Order != etcd.SortNone {
		if order, err := orderToSQL(opts.Order); err != nil {
			return err
		} else if target, err := targetToSQL(opts.Target); err != nil {
			return err
		} else {
			query = query.Order(fmt.Sprintf("%s %s", target, order))
		}
	}

	result := reflect.New(reflect.TypeOf(c.model))
	rows, err := query.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(result); err != nil {
			return err
		}

		if err := writeToProtobuf(result, val); err != nil {
			return err
		}

		// TODO: get the pkey from the model or just remove this arg, why is it here - they've already got the model
		if err := f(""); err != nil {
			return err
		}
	}

	return rows.Close()
}

func (c *postgresReadOnlyCollection) Count() (int64, error) {
	var result int64
	if err := c.newQuery().Count(&result).Error; err != nil {
		return 0, err
	}
	return result, nil
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

// These operations will (probably) never be supported for a postgres collection

func (c *postgresReadOnlyCollection) GetBlock(key string, val proto.Message) error {
	return errors.New("GetBlock is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) TTL(key string) (int64, error) {
	return 0, errors.New("TTL is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) ListRev(val proto.Message, opts *Options, f func(key string, createRev int64) error) error {
	return errors.New("ListRev is not supported on read-only postgres collections")
}

type postgresReadWriteCollection struct {
	*postgresCollection
	stm STM
}

func (c *postgresReadWriteCollection) Get(key string, val proto.Message) error {
	return errors.New("Get is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) Put(key string, val proto.Message) error {
	return errors.New("Put is not supported on read-write postgres collections")
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

// These operations will (probably) never be supported on a postgres collection

func (c *postgresReadWriteCollection) TTL(key string) (int64, error) {
	return 0, errors.New("TTL is not supported on read-write postgres collections")
}

func (c *postgresReadWriteCollection) PutTTL(key string, val proto.Message, ttl int64) error {
	return errors.New("PutTTL is not supported on read-write postgres collections")
}
