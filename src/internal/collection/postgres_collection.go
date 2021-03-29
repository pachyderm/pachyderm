package collection

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/version"
)

const (
	modifiedTriggerName  = "update_modified_trigger"
	modifiedFunctionName = "update_modified_time"
	watchBaseName        = "pwc" // "Pachyderm Watch Channel"
	indexBaseName        = "idx"
	pgIdentBase64Values  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_$"
)

var (
	pgIdentBase64Encoding = base64.NewEncoding(pgIdentBase64Values).WithPadding(base64.NoPadding)
)

type postgresCollection struct {
	db         *sqlx.DB
	listener   *PostgresListener
	template   proto.Message
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

type SQLInfo struct {
	Table   string
	Indexes []*SQLField
}

func indexFieldName(field *SQLField) string {
	return indexBaseName + "_" + field.SQLName
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

	createTrigger := fmt.Sprintf("create trigger %s before insert or update or delete on %s for each row execute procedure %s();", modifiedTriggerName, info.Table, modifiedFunctionName)
	if _, err := tx.Exec(createTrigger); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

func ensureNotifyTrigger(tx *sqlx.Tx, info *SQLInfo) error {
	triggerName := "notify_watches"
	functionName := "func_" + triggerName
	dropTrigger := fmt.Sprintf("drop trigger if exists %s on %s;", triggerName, info.Table)
	if _, err := tx.Exec(dropTrigger); err != nil {
		return errors.EnsureStack(err)
	}

	createFunction := fmt.Sprintf(`
create or replace function %s() returns trigger as $$
declare
  row record;
	payload text;
	base_channel text;
	channel text;
	field text;
	value text;
begin
  case tg_op
	when 'INSERT', 'UPDATE' then
	  row := new;
	when 'DELETE' then
	  row := old;
	else
	  raise exception 'Unrecognized tg_op: "%%"', tg_op;
	end case;

	payload := row.key || ' ' || tg_op || ' ' || date_part('epoch', now())::text || ' ' || encode(row.proto, 'base64');
	base_channel := '%s_' || tg_table_name;

	if tg_argv is not null then
		foreach field in array tg_argv loop
			execute 'select ($1).' || field || '::text;' into value using new;
			if value is not null then
				channel := base_channel || '_' || md5(value);
				perform pg_notify(channel, field || ' ' || value || ' ' || payload);
			end if;
		end loop;
	end if;

  perform pg_notify(base_channel, 'key' || ' ' || row.key || ' ' || payload);
	return row;
end;
$$ language 'plpgsql';
	`, functionName, watchBaseName)
	if _, err := tx.Exec(createFunction); err != nil {
		return errors.EnsureStack(err)
	}

	indexFields := []string{}
	for _, field := range info.Indexes {
		indexFields = append(indexFields, "'"+indexFieldName(field)+"'")
	}

	createTrigger := fmt.Sprintf("create trigger %s after insert or update or delete on %s for each row execute procedure %s(%s);", triggerName, info.Table, functionName, strings.Join(indexFields, ", "))
	if _, err := tx.Exec(createTrigger); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

type model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   string
	Key       string
	Proto     []byte
}

// Ensure the table and all indices exist
func ensureCollection(db *sqlx.DB, info *SQLInfo) error {
	columns := []string{
		"createdat timestamp with time zone default current_timestamp",
		"updatedat timestamp with time zone default current_timestamp",
		"proto bytea",
		"version text",
		"key text primary key",
	}

	for _, field := range info.Indexes {
		columns = append(columns, fmt.Sprintf("%s text", indexFieldName(field)))
	}
	// TODO: create indexes on table

	// TODO: use actual context
	return NewSQLTx(context.Background(), db, func(tx *sqlx.Tx) error {
		createTable := fmt.Sprintf("create table if not exists %s (%s);", info.Table, strings.Join(columns, ", "))
		if _, err := tx.Exec(createTable); err != nil {
			return errors.EnsureStack(err)
		}
		// TODO: drop all existing triggers/functions on this collection to make
		// sure we don't leave any hanging around due to a rename
		if err := ensureModifiedTrigger(tx, info); err != nil {
			return err
		}
		if err := ensureNotifyTrigger(tx, info); err != nil {
			return err
		}
		return nil
	})
}

func parseProto(template proto.Message, indexes []*Index) (*SQLInfo, error) {
	templateType := reflect.TypeOf(template).Elem()

	sqlIndexes := []*SQLField{}
	for _, idx := range indexes {
		field, ok := templateType.FieldByName(idx.Field)
		if !ok {
			return nil, errors.Errorf("Index field '%v.%s' not found", templateType, idx.Field)
		}

		goType := fmt.Sprintf("%v", field.Type)
		sqlType, err := toSQLType(goType)
		if err != nil {
			return nil, err
		}
		sqlIndexes = append(sqlIndexes, &SQLField{
			SQLType: sqlType,
			SQLName: toSQLName(field.Name),
			GoType:  goType,
			GoName:  field.Name,
		})
	}

	return &SQLInfo{
		Table:   toSQLName(templateType.Name()) + "s",
		Indexes: sqlIndexes,
	}, nil
}

// NewPostgresCollection creates a new collection backed by postgres.
func NewPostgresCollection(db *sqlx.DB, listener *PostgresListener, template proto.Message, indexes []*Index) (PostgresCollection, error) {
	sqlInfo, err := parseProto(template, indexes)
	if err != nil {
		return nil, err
	}

	if err := ensureCollection(db, sqlInfo); err != nil {
		return nil, err
	}

	c := &postgresCollection{
		db:         db,
		listener:   listener,
		template:   template,
		sqlInfo:    sqlInfo,
		withFields: make(map[string]interface{}),
	}
	return c, nil
}

func (c *postgresCollection) With(field string, value interface{}) PostgresCollection {
	newWithFields := make(map[string]interface{})
	for k, v := range c.withFields {
		newWithFields[k] = v
	}

	return &postgresCollection{
		db:         c.db,
		listener:   c.listener,
		template:   c.template,
		sqlInfo:    c.sqlInfo,
		withFields: newWithFields,
	}
}

func (c *postgresCollection) tableWatchChannel() string {
	return watchBaseName + "_" + c.sqlInfo.Table
}

func (c *postgresCollection) ReadOnly(ctx context.Context) PostgresReadOnlyCollection {
	return &postgresReadOnlyCollection{c, ctx}
}

func (c *postgresCollection) ReadWrite(tx *sqlx.Tx) PostgresReadWriteCollection {
	return &postgresReadWriteCollection{c, tx}
}

// NewSQLTx starts a transaction on the given DB, passes it to the callback, and
// finishes the transaction afterwards. If the callback was successful, the
// transaction is committed. If any errors occur, the transaction is rolled
// back.  This will reattempt the transaction up to three times.
func NewSQLTx(ctx context.Context, db *sqlx.DB, apply func(*sqlx.Tx) error) error {
	errs := []error{}

	attemptTx := func() (bool, error) {
		tx, err := db.BeginTxx(context.Background(), nil)
		if err != nil {
			return true, errors.EnsureStack(err)
		}

		defer tx.Rollback()

		err = apply(tx)
		if err != nil {
			return true, err
		}

		return true, errors.EnsureStack(tx.Commit())
	}

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
	pqerr := &pq.Error{}
	if errors.As(err, pqerr) {
		return pqerr.Code == pgerrcode.UniqueViolation
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

func (c *postgresCollection) getInternal(key string, val proto.Message, q sqlx.Queryer) error {
	queryString := fmt.Sprintf("select proto from %s where key = $1;", c.sqlInfo.Table)
	result := &model{}

	if err := sqlx.Get(q, result, queryString, key); err != nil {
		return c.mapSQLError(err, key)
	}

	return errors.EnsureStack(proto.Unmarshal(result.Proto, val))
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
		return "key", nil
	}
	return "", errors.Errorf("unsupported sort target for postgres collections: %d", target)
}

func (c *postgresReadOnlyCollection) list(opts *Options, f func(*model) error) error {
	query := fmt.Sprintf("select key, createdat, updatedat, proto from %s", c.sqlInfo.Table)

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

	result := &model{}
	for rows.Next() {
		if err := rows.StructScan(result); err != nil {
			return c.mapSQLError(err, "")
		}

		if err := f(result); err != nil {
			return err
		}
	}

	return c.mapSQLError(rows.Close(), "")
}

func (c *postgresReadOnlyCollection) List(val proto.Message, opts *Options, f func() error) error {
	err := c.list(opts, func(m *model) error {
		if err := proto.Unmarshal(m.Proto, val); err != nil {
			return errors.EnsureStack(err)
		}

		return f()
	})

	if errors.Is(err, errutil.ErrBreak) {
		return nil
	}
	return err
}

func (c *postgresReadOnlyCollection) Count() (int64, error) {
	query := fmt.Sprintf("select count(*) from %s", c.sqlInfo.Table)
	row := c.db.QueryRow(query)

	var result int64
	err := row.Scan(&result)
	return result, c.mapSQLError(err, "")
}

func (c *postgresReadOnlyCollection) Watch(opts ...watch.Option) (watch.Watcher, error) {
	options := watch.SumOptions(opts...)

	channelNames := []string{c.tableWatchChannel()}
	watcher, err := c.listener.listen(channelNames, c.template, nil, nil, options)
	if err != nil {
		return nil, err
	}

	go func() {
		// Do a List of the collection to get the initial state
		lastUpdated := time.Time{}
		val := cloneProtoMsg(c.template)
		if err := c.list(&Options{Target: options.SortTarget, Order: options.SortOrder}, func(m *model) error {
			if err := proto.Unmarshal(m.Proto, val); err != nil {
				return errors.EnsureStack(err)
			}

			if lastUpdated.Before(m.UpdatedAt) {
				lastUpdated = m.UpdatedAt
			}

			return watcher.sendInitial(&watch.Event{
				Key:      []byte(m.Key),
				Value:    m.Proto,
				Type:     watch.EventPut,
				Template: c.template,
			})
		}); err != nil {
			// Ignore any additional error here - we're already attempting to send an error to the user
			watcher.sendInitial(&watch.Event{Type: watch.EventError, Err: err})
			return
		}

		// Forward all buffered notifications until the watcher is closed
		watcher.forwardNotifications(lastUpdated)
	}()

	return watcher, nil
}

func (c *postgresReadOnlyCollection) WatchF(f func(*watch.Event) error, opts ...watch.Option) error {
	options := watch.SumOptions(opts...)
	channelNames := []string{c.tableWatchChannel()}
	watcher, err := c.listener.listen(channelNames, c.template, nil, nil, options)
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Do a List of the collection to get the initial state
	lastUpdated := time.Time{}
	val := cloneProtoMsg(c.template)
	if err := c.list(&Options{Target: options.SortTarget, Order: options.SortOrder}, func(m *model) error {
		if err := proto.Unmarshal(m.Proto, val); err != nil {
			return errors.EnsureStack(err)
		}

		if lastUpdated.Before(m.UpdatedAt) {
			lastUpdated = m.UpdatedAt
		}

		// watcher may be closed by the listener due to a connection error, parse
		// error, or full channel - check if it's closed and abort
		if watcher.isClosed() {
			return errors.New("watcher was closed")
		}

		return f(&watch.Event{
			Key:      []byte(m.Key),
			Value:    m.Proto,
			Type:     watch.EventPut,
			Template: c.template,
		})
	}); err != nil {
		if errors.Is(err, errutil.ErrBreak) {
			return nil
		}
		return err
	}

	// Forward all buffered notifications until the watcher is closed
	go watcher.forwardNotifications(lastUpdated)
	return watchF(context.Background(), watcher, f)
}

func (c *postgresReadOnlyCollection) WatchOne(key string, opts ...watch.Option) (watch.Watcher, error) {
	return nil, errors.New("WatchOne is not supported on read-only postgres collections")
}

func (c *postgresReadOnlyCollection) WatchOneF(key string, f func(*watch.Event) error, opts ...watch.Option) error {
	return errors.New("WatchOneF is not supported on read-only postgres collections")
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

func (c *postgresReadWriteCollection) getWriteParams(key string, val proto.Message) (map[string]interface{}, error) {
	data, err := proto.Marshal(val)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	params := map[string]interface{}{
		"version": version.PrettyVersion(),
		"key":     key,
		"proto":   data,
	}

	reflVal := reflect.ValueOf(val)
	for _, idx := range c.sqlInfo.Indexes {
		fieldValue := reflect.Indirect(reflVal).FieldByName(idx.GoName).Interface()
		params[indexFieldName(idx)] = fieldValue
	}

	return params, nil
}

// Get then update all values
func (c *postgresReadWriteCollection) Update(key string, val proto.Message, f func() error) error {
	if err := c.Get(key, val); err != nil {
		return err
	}
	if err := f(); err != nil {
		return err
	}

	params, err := c.getWriteParams(key, val)
	if err != nil {
		return err
	}

	updateFields := []string{}
	for k := range params {
		updateFields = append(updateFields, fmt.Sprintf("%s = :%s", k, k))
	}

	query := fmt.Sprintf("update %s set %s where key = :key", c.sqlInfo.Table, strings.Join(updateFields, ", "))

	_, err = c.tx.NamedExec(query, params)
	return c.mapSQLError(err, key)
}

func (c *postgresReadWriteCollection) insert(key string, val proto.Message, upsert bool) error {
	params, err := c.getWriteParams(key, val)
	if err != nil {
		return err
	}

	columns := []string{}
	paramNames := []string{}
	for k := range params {
		columns = append(columns, k)
		paramNames = append(paramNames, ":"+k)
	}
	columnList := strings.Join(columns, ", ")
	paramList := strings.Join(paramNames, ", ")

	query := fmt.Sprintf("insert into %s (%s) values (%s)", c.sqlInfo.Table, columnList, paramList)
	if upsert {
		upsertFields := []string{}
		for _, column := range columns {
			upsertFields = append(upsertFields, fmt.Sprintf("%s = :%s", column, column))
		}
		query = fmt.Sprintf("%s on conflict (key) do update set (%s) = (%s)", query, columnList, paramList)
	}

	_, err = c.tx.NamedExec(query, params)
	return c.mapSQLError(err, key)
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
	return c.insert(key, val, false)
}

func (c *postgresReadWriteCollection) Delete(key string) error {
	query := fmt.Sprintf("delete from %s where key = $1;", c.sqlInfo.Table)
	res, err := c.tx.Exec(query, key)
	if err != nil {
		return c.mapSQLError(err, key)
	}

	if count, err := res.RowsAffected(); err != nil {
		return c.mapSQLError(err, key)
	} else if count == 0 {
		return errors.WithStack(ErrNotFound{c.sqlInfo.Table, key})
	}
	return nil
}

func (c *postgresReadWriteCollection) DeleteAll() error {
	query := fmt.Sprintf("delete from %s;", c.sqlInfo.Table)
	_, err := c.tx.Exec(query)
	return c.mapSQLError(err, "")
}
