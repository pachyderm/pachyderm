package collection

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

const (
	modifiedTriggerName  = "update_modified_trigger"
	modifiedFunctionName = "update_modified_time"
	watchBaseName        = "pwc" // "Pachyderm Watch Channel"
	pgIdentBase64Values  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_$"
)

var (
	pgIdentBase64Encoding = base64.NewEncoding(pgIdentBase64Values).WithPadding(base64.NoPadding)
)

// PostgresModel is the interface that all models must fulfill to be used in a postgres collection
type PostgresModel interface {
	TableName() string
	Indexes() []*Index
	WriteToProtobuf(proto.Message) error
	LoadFromProtobuf(proto.Message) error
}

type postgresCollection struct {
	db         *sqlx.DB
	model      PostgresModel
	sqlInfo    *SQLInfo
	withFields map[string]interface{}
}

func toSQLName(name string) string {
	return strings.ToLower(name)
}

var sqlTypes = map[string]string{
	"string":    "varchar",
	"time.Time": "timestamp",
	"bool":      "bool",
}

func toSQLType(gotype string) (string, error) {
	if result, ok := sqlTypes[gotype]; ok {
		return result, nil
	}
	return "", errors.Errorf("No SQL type for %s", gotype)
}

type SQLIndex struct {
	Index
}

type SQLField struct {
	SQLType string
	SQLName string
	GoType  string
	GoName  string
}

// TODO: would be better if we just serialize this field in the protobuf to a byte array?
func (sf *SQLField) fromRow(row reflect.Value) interface{} {
	return reflect.Indirect(row).FieldByName(sf.GoName).Interface()
}

type SQLInfo struct {
	Table   string
	Pkey    *SQLField
	Fields  []SQLField
	Indexes []*Index
}

func forEachField(modelType reflect.Type, cb func(reflect.StructField) error) error {
	for i := 0; i < modelType.NumField(); i++ {
		if err := cb(modelType.Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func parseModel(model PostgresModel) (*SQLInfo, error) {
	modelType := reflect.TypeOf(model).Elem()

	pkey := []*SQLField{}
	sqlFields := []SQLField{}
	if err := forEachField(modelType, func(field reflect.StructField) error {
		// Get the field tag value
		tag := field.Tag.Get("collection")

		goType := fmt.Sprintf("%v", field.Type)
		sqlType, err := toSQLType(goType)
		if err != nil {
			return err
		}
		sqlFields = append(sqlFields, SQLField{
			SQLName: toSQLName(field.Name),
			SQLType: sqlType,
			GoName:  field.Name,
			GoType:  goType,
		})

		for _, prop := range strings.Split(tag, ",") {
			switch prop {
			case "primaryKey":
				pkey = append(pkey, &sqlFields[len(sqlFields)-1])
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(pkey) == 0 {
		return nil, errors.Errorf("%v has no field tagged with collection:\"primaryKey\"", modelType)
	} else if len(pkey) != 1 {
		return nil, errors.Errorf("%v has multiple fields tagged with collection:\"primaryKey\": %v", modelType, pkey[0].GoName)
	}

	return &SQLInfo{model.TableName(), pkey[0], sqlFields, model.Indexes()}, nil
}

func ensureModifiedTrigger(tx *sqlx.Tx, info *SQLInfo) error {
	// Create trigger to update the modified timestamp (no way to do this without recreating it each time?)
	dropTrigger := fmt.Sprintf("drop trigger if exists %s on %s;", modifiedTriggerName, info.Table)
	if _, err := tx.Exec(dropTrigger); err != nil {
		return errors.EnsureStack(err)
	}

	createFunction := fmt.Sprintf("create or replace function %s() returns trigger as $$ begin new.updatedat = now(); return new; end; $$ language 'plpgsql';", modifiedFunctionName)
	if _, err := tx.Exec(createFunction); err != nil {
		return errors.EnsureStack(err)
	}

	createTrigger := fmt.Sprintf("create trigger %s before update or insert on %s for each row execute procedure %s();", modifiedTriggerName, info.Table, modifiedFunctionName)
	if _, err := tx.Exec(createTrigger); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

// Ensure the table and all indices exist
func ensureCollection(db *sqlx.DB, info *SQLInfo) error {
	columns := []string{}
	for _, field := range info.Fields {
		if field.GoName == "CreatedAt" || field.GoName == "UpdatedAt" {
			columns = append(columns, fmt.Sprintf("%s %s default current_timestamp", field.SQLName, field.SQLType))
		} else {
			columns = append(columns, fmt.Sprintf("%s %s", field.SQLName, field.SQLType))
		}
	}
	columns = append(columns, fmt.Sprintf("primary key(%s)", info.Pkey.SQLName))

	// TODO: use actual context
	return NewSQLTx(db, context.Background(), func(tx *sqlx.Tx) error {
		createTable := fmt.Sprintf("create table if not exists %s (%s);", info.Table, strings.Join(columns, ", "))
		if _, err := tx.Exec(createTable); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	})
}

// NewPostgresCollection creates a new collection backed by postgres.
func NewPostgresCollection(db *sqlx.DB, model PostgresModel) PostgresCollection {
	sqlInfo, err := parseModel(model)
	if err != nil {
		panic(err)
	}

	if err := ensureCollection(db, sqlInfo); err != nil {
		panic(err)
	}

	// TODO: handle error

	c := &postgresCollection{
		db:         db,
		model:      model,
		sqlInfo:    sqlInfo,
		withFields: make(map[string]interface{}),
	}
	return c
}

func (c *postgresCollection) With(field string, value interface{}) PostgresCollection {
	newWithFields := make(map[string]interface{})
	for k, v := range c.withFields {
		newWithFields[k] = v
	}

	return &postgresCollection{
		db:         c.db,
		model:      c.model,
		sqlInfo:    c.sqlInfo,
		withFields: newWithFields,
	}
}

func (c *postgresCollection) ReadOnly(ctx context.Context) PostgresReadOnlyCollection {
	return &postgresReadOnlyCollection{c, ctx}
}

func (c *postgresCollection) ReadWrite(tx *sqlx.Tx) PostgresReadWriteCollection {
	return &postgresReadWriteCollection{c, tx}
}

func NewSQLTx(db *sqlx.DB, ctx context.Context, apply func(*sqlx.Tx) error) error {
	errs := []error{}

	attemptTx := func() (bool, error) {
		tx, err := db.BeginTxx(context.Background(), nil)
		if err != nil {
			return true, errors.EnsureStack(err)
		}

		// TODO: log something on failed rollback?
		defer tx.Rollback()

		err = apply(tx)
		if err != nil {
			return true, err
		}

		return true, errors.EnsureStack(tx.Commit())
	}

	// TODO: try indefinitely?  time out?
	for i := 0; i < 3; i++ {
		if done, err := attemptTx(); done {
			return err
		} else {
			errs = append(errs, err)
		}
	}

	return errors.Errorf("sql transaction rolled back too many times: %v", errs)
}

func (c *postgresCollection) Claim(ctx context.Context, key string, val proto.Message, f func(context.Context) error) error {
	return errors.New("Claim is not supported on postgres collections")
}

func isDuplicateKeyError(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code == "23505" // Postgres error code: unique_violation
	}
	return false
}

func (c *postgresCollection) mapSQLError(err error, key string) error {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.WithStack(ErrNotFound{c.sqlInfo.Table, key})
		} else if isDuplicateKeyError(err) {
			return errors.WithStack(ErrExists{c.sqlInfo.Table, key})
		}
		return errors.EnsureStack(err)
	}
	return nil
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

func (c *postgresCollection) getInternal(key string, val proto.Message, q sqlx.Queryer) error {
	queryString := fmt.Sprintf("select * from %s where %s = $1;", c.sqlInfo.Table, c.sqlInfo.Pkey.SQLName)
	result := reflect.New(reflect.ValueOf(c.model).Elem().Type())

	if err := sqlx.Get(q, result.Interface(), queryString, key); err != nil {
		return c.mapSQLError(err, key)
	}
	return writeToProtobuf(result, val)
}

func (c *postgresReadOnlyCollection) Get(key string, val proto.Message) error {
	return c.getInternal(key, val, c.db)
}

func (c *postgresReadOnlyCollection) GetByIndex(index *Index, indexVal interface{}, val proto.Message, opts *Options, f func() error) error {
	return c.With(index.Field, indexVal).ReadOnly(c.ctx).List(val, opts, f)
}

func orderToSQL(order etcd.SortOrder) (string, error) {
	switch order {
	case SortAscend:
		return "asc", nil
	case SortDescend:
		return "desc", nil
	}
	return "", errors.Errorf("unsupported sort order: %d", order)
}

func (c *postgresReadOnlyCollection) targetToSQL(target etcd.SortTarget) (string, error) {
	switch target {
	case SortByCreateRevision:
		return "createdat", nil
	case SortByModRevision:
		return "updatedat", nil
	case SortByKey:
		return c.sqlInfo.Pkey.SQLName, nil
	}
	return "", errors.Errorf("unsupported sort target for postgres collections: %d", target)
}

func (c *postgresReadOnlyCollection) List(val proto.Message, opts *Options, f func() error) error {
	query := fmt.Sprintf("select * from %s", c.sqlInfo.Table)

	if opts.Order != SortNone {
		if order, err := orderToSQL(opts.Order); err != nil {
			return err
		} else if target, err := c.targetToSQL(opts.Target); err != nil {
			return err
		} else {
			query += fmt.Sprintf(" order by %s %s", target, order)
		}
	}

	rows, err := c.db.Queryx(query)
	if err != nil {
		return c.mapSQLError(err, "")
	}
	defer rows.Close()

	result := reflect.New(reflect.ValueOf(c.model).Elem().Type())
	for rows.Next() {
		if err := rows.StructScan(result.Interface()); err != nil {
			return c.mapSQLError(err, "")
		}

		if err := writeToProtobuf(result, val); err != nil {
			return err
		}

		if err := f(); err != nil {
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}

	return c.mapSQLError(rows.Close(), "")
}

func (c *postgresReadOnlyCollection) Count() (int64, error) {
	query := fmt.Sprintf("select count(*) from %s", c.sqlInfo.Table)
	row := c.db.QueryRow(query)

	var result int64
	err := row.Scan(&result)
	return result, c.mapSQLError(err, "")
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
	tx *sqlx.Tx
}

func (c *postgresReadWriteCollection) Get(key string, val proto.Message) error {
	return c.getInternal(key, val, c.tx)
}

// Upsert without the get and callback
func (c *postgresReadWriteCollection) Put(key string, val proto.Message) error {
	return c.insert(key, val, true)
}

// Get then update all values
func (c *postgresReadWriteCollection) Update(key string, val proto.Message, f func() error) error {
	if err := c.Get(key, val); err != nil {
		return err
	}
	if err := f(); err != nil {
		return err
	}

	row := reflect.New(reflect.ValueOf(c.model).Elem().Type())
	row.MethodByName("LoadFromProtobuf").Call([]reflect.Value{reflect.ValueOf(val)})

	params := map[string]interface{}{}
	updateFields := []string{}
	for _, field := range c.sqlInfo.Fields {
		if field.GoName == "UpdatedAt" || field.GoName == "CreatedAt" {
			continue
		}
		params[field.SQLName] = reflect.Indirect(row).FieldByName(field.GoName).Interface()
		updateFields = append(updateFields, fmt.Sprintf("%s = :%s", field.SQLName, field.SQLName))
	}

	query := fmt.Sprintf("update %s set %s where %s = :%s", c.sqlInfo.Table, strings.Join(updateFields, ", "), c.sqlInfo.Pkey.SQLName, c.sqlInfo.Pkey.SQLName)

	if _, err := c.tx.NamedExec(query, params); err != nil {
		return c.mapSQLError(err, key)
	}

	return c.notify(key, row, NotifyPayload_PUT)
}

func (c *postgresReadWriteCollection) notify(key string, row reflect.Value, kind NotifyPayload_Type) error {
	payload := &NotifyPayload{Info: &NotifyInfo{Table: c.sqlInfo.Table}, Key: key, Type: kind}
	payloadData, err := payload.Marshal()
	if err != nil {
		return errors.EnsureStack(err)
	}
	payloadString := pgIdentBase64Encoding.EncodeToString(payloadData)
	query := fmt.Sprintf("notify %s_%s, '%s';\n", watchBaseName, c.sqlInfo.Table, payloadString)
	if _, err := c.tx.Exec(query); err != nil {
		return c.mapSQLError(err, key)
	}

	for _, idx := range c.sqlInfo.Indexes {
		payload.Info.Fields = []string{}
		payload.Info.Values = []string{}

		// TODO: allow compound indexes?
		// TODO: would be nice if we could just serialize the field from the
		// original protobuf? But this might be fine since we limit the types that
		// the model can use.
		value := reflect.Indirect(row).FieldByName(idx.Field).Interface()
		payload.Info.Fields = append(payload.Info.Fields, idx.Field)
		payload.Info.Values = append(payload.Info.Values, fmt.Sprintf("%v", value))

		data, err := payload.Info.Marshal()
		if err != nil {
			return errors.EnsureStack(err)
		}
		hash := pachhash.Sum(data)
		hashBase64 := pgIdentBase64Encoding.EncodeToString(hash[:])

		payloadData, err = payload.Marshal()
		if err != nil {
			return errors.EnsureStack(err)
		}
		payloadString = pgIdentBase64Encoding.EncodeToString(payloadData)

		query = fmt.Sprintf("notify %s_%s, '%s';\n", watchBaseName, hashBase64, payloadString)
		if _, err := c.tx.Exec(query); err != nil {
			return c.mapSQLError(err, key)
		}
	}
	return nil
}

func (c *postgresReadWriteCollection) insert(key string, val proto.Message, upsert bool) error {
	row := reflect.New(reflect.ValueOf(c.model).Elem().Type())
	row.MethodByName("LoadFromProtobuf").Call([]reflect.Value{reflect.ValueOf(val)})

	columns := []string{}
	paramNames := []string{}
	params := map[string]interface{}{}
	for _, field := range c.sqlInfo.Fields {
		if field.GoName == "UpdatedAt" || field.GoName == "CreatedAt" {
			continue
		}
		columns = append(columns, field.SQLName)
		paramNames = append(paramNames, ":"+field.SQLName)
		params[field.SQLName] = reflect.Indirect(row).FieldByName(field.GoName).Interface()
	}

	columnList := strings.Join(columns, ", ")
	paramList := strings.Join(paramNames, ", ")

	query := fmt.Sprintf("insert into %s (%s) values (%s)", c.sqlInfo.Table, columnList, paramList)
	if upsert {
		upsertFields := []string{}
		for _, column := range columns {
			upsertFields = append(upsertFields, fmt.Sprintf("%s = :%s", column, column))
		}
		query = fmt.Sprintf("%s on conflict (%s) do update set (%s) = (%s)", query, c.sqlInfo.Pkey.SQLName, columnList, paramList)
	}

	if _, err := c.tx.NamedExec(query, params); err != nil {
		return c.mapSQLError(err, key)
	}

	return c.notify(key, row, NotifyPayload_PUT)
}

// Insert on conflict update all values (except createdat)
func (c *postgresReadWriteCollection) Upsert(key string, val proto.Message, f func() error) error {
	if err := c.Get(key, val); err != nil && !IsErrNotFound(err) {
		return err
	}
	if err := f(); err != nil {
		return err
	}
	return c.Put(key, val)
}

// Insert
func (c *postgresReadWriteCollection) Create(key string, val proto.Message) error {
	// TODO: require that the proto pkey matches key or override it in the insert
	return c.insert(key, val, false)
}

func (c *postgresReadWriteCollection) Delete(key string) error {
	// For delete notifications we need to load the deleted row
	query := fmt.Sprintf("delete from %s where %s = $1 returning *;", c.sqlInfo.Table, c.sqlInfo.Pkey.SQLName)
	row := reflect.New(reflect.ValueOf(c.model).Elem().Type())

	if err := sqlx.Get(c.tx, row.Interface(), query, key); err != nil {
		// TODO: what happens if there is no row?
		return c.mapSQLError(err, key)
	}

	return c.notify(key, row, NotifyPayload_DELETE)

	/*
		res, err := c.tx.NamedExec(query, key)
		if err != nil {
			return c.mapSQLError(err, key)
		}

		if count, err := res.RowsAffected(); err != nil {
			return c.mapSQLError(err, key)
		} else if count == 0 {
			return errors.WithStack(ErrNotFound{c.sqlInfo.Table, key})
		}
		return nil
	*/
}

func (c *postgresReadWriteCollection) DeleteAll() error {
	// TODO: delete notifications are tough here, can we use a trigger?
	// Otherwise we need to load each row to notify the right channel.
	// It's easier for collection-level watches but not for indexed watches.
	query := fmt.Sprintf("delete from %s;", c.sqlInfo.Table)
	_, err := c.db.Exec(query)
	return c.mapSQLError(err, "")
}
