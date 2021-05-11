package collection

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
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

type postgresCollection struct {
	table    string
	db       *sqlx.DB
	listener *PostgresListener
	template proto.Message
	indexes  []*Index
	keyCheck func(string) error
}

func indexFieldName(idx *Index) string {
	return "idx_" + idx.Name
}

type model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   string
	Key       string
	Proto     []byte
}

// NewPostgresCollection creates a new collection backed by postgres.
func NewPostgresCollection(name string, db *sqlx.DB, listener *PostgresListener, template proto.Message, indexes []*Index, keyCheck func(string) error) PostgresCollection {
	return &postgresCollection{
		table:    name,
		db:       db,
		listener: listener,
		template: template,
		indexes:  indexes,
		keyCheck: keyCheck,
	}
}

// Indexes passed into queries are required to be the same object used at
// construction time to ensure that their Name field and Extract method are
// identical.
func (c *postgresCollection) validateIndex(index *Index) error {
	found := false
	for _, idx := range c.indexes {
		if idx == index {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("Unknown collection index: %s", index.Name)
	}
	return nil
}

func (c *postgresCollection) tableWatchChannel() string {
	return "pwc_" + c.table
}

func (c *postgresCollection) indexWatchChannel(field string, value string) string {
	data := fmt.Sprintf("%s %s", field, value)
	return c.tableWatchChannel() + "_" + fmt.Sprintf("%x", md5.Sum([]byte(data)))
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
// back.  This will reattempt the transaction forever.
func NewSQLTx(ctx context.Context, db *sqlx.DB, apply func(*sqlx.Tx) error) error {
	attemptTx := func() error {
		tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return errors.EnsureStack(err)
		}

		err = apply(tx)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = errors.EnsureStack(tx.Commit())
		if err != nil {
			return err
		}
		return nil
	}

	for {
		if err := attemptTx(); err != nil {
			if !isTransactionError(err) {
				return err
			}
		} else {
			return nil
		}
	}
}

// NewDryrunSQLTx is identical to NewSQLTx except it will always roll back the
// transaction instead of committing it.
func NewDryrunSQLTx(ctx context.Context, db *sqlx.DB, apply func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer tx.Rollback()

	err = apply(tx)
	if err != nil {
		return err
	}
	return nil
}

func (c *postgresCollection) Claim(ctx context.Context, key string, val proto.Message, f func(context.Context) error) error {
	return errors.New("Claim is not supported on postgres collections")
}

func isTransactionError(err error) bool {
	pqerr := &pq.Error{}
	if errors.As(err, pqerr) {
		return pgerrcode.IsTransactionRollback(string(pqerr.Code))
	}
	return false
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
			return errors.WithStack(ErrNotFound{c.table, key})
		} else if isDuplicateKeyError(err) {
			return errors.WithStack(ErrExists{c.table, key})
		}
		return errors.EnsureStack(err)
	}
	return nil
}

type postgresReadOnlyCollection struct {
	*postgresCollection
	ctx context.Context
}

func (c *postgresCollection) get(ctx context.Context, key string, q sqlx.QueryerContext) (*model, error) {
	result := &model{}
	queryString := fmt.Sprintf("select proto, updatedat from collections.%s where key = $1;", c.table)
	if err := sqlx.GetContext(ctx, q, result, queryString, key); err != nil {
		return nil, c.mapSQLError(err, key)
	}
	return result, nil
}

func (c *postgresReadOnlyCollection) Get(key string, val proto.Message) error {
	result, err := c.get(c.ctx, key, c.db)
	if err != nil {
		return err
	}
	return errors.EnsureStack(proto.Unmarshal(result.Proto, val))
}

func (c *postgresCollection) getByIndex(ctx context.Context, q sqlx.ExtContext, index *Index, indexVal string, val proto.Message, opts *Options, f func(string) error) error {
	if err := c.validateIndex(index); err != nil {
		return err
	}
	return c.list(ctx, map[string]string{indexFieldName(index): indexVal}, opts, q, func(m *model) error {
		if err := proto.Unmarshal(m.Proto, val); err != nil {
			return errors.EnsureStack(err)
		}
		return f(m.Key)
	})
}

func (c *postgresReadOnlyCollection) GetByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(string) error) error {
	return c.getByIndex(c.ctx, c.db, index, indexVal, val, opts, f)
}

func (c *postgresReadWriteCollection) GetByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(string) error) error {
	return c.getByIndex(context.Background(), c.tx, index, indexVal, val, opts, f)
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

func targetToSQL(target etcd.SortTarget) (string, error) {
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

func (c *postgresCollection) list(
	ctx context.Context,
	withFields map[string]string,
	opts *Options,
	q sqlx.ExtContext,
	f func(*model) error) error {
	query := fmt.Sprintf("select key, createdat, updatedat, proto from collections.%s", c.table)

	params := map[string]interface{}{}
	if len(withFields) > 0 {
		fields := []string{}
		for k, v := range withFields {
			fields = append(fields, fmt.Sprintf("%s = :%s", k, k))
			params[k] = v
		}
		query += " where " + strings.Join(fields, " and ")
	}

	if opts.Order != SortNone {
		if order, err := orderToSQL(opts.Order); err != nil {
			return err
		} else if target, err := targetToSQL(opts.Target); err != nil {
			return err
		} else {
			query += fmt.Sprintf(" order by %s %s", target, order)
		}
	}

	rows, err := sqlx.NamedQueryContext(ctx, q, query, params)
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
			if errors.Is(err, errutil.ErrBreak) {
				return nil
			}
			return err
		}
	}

	return c.mapSQLError(rows.Close(), "")
}

func (c *postgresReadOnlyCollection) list(withFields map[string]string, opts *Options, f func(*model) error) error {
	return c.postgresCollection.list(c.ctx, withFields, opts, c.db, f)
}

func (c *postgresReadOnlyCollection) List(val proto.Message, opts *Options, f func(string) error) error {
	return c.list(nil, opts, func(m *model) error {
		if err := proto.Unmarshal(m.Proto, val); err != nil {
			return errors.EnsureStack(err)
		}
		return f(m.Key)
	})
}

func (c *postgresReadOnlyCollection) listRev(withFields map[string]string, val proto.Message, opts *Options, f func(string, int64) error) error {
	fakeRev := int64(0)
	lastTimestamp := time.Time{}

	updateRev := func(t time.Time) {
		if t.After(lastTimestamp) {
			lastTimestamp = t
			fakeRev++
		}
	}

	return c.list(withFields, opts, func(m *model) error {
		if err := proto.Unmarshal(m.Proto, val); err != nil {
			return errors.EnsureStack(err)
		}

		if opts.Target == SortByCreateRevision {
			updateRev(m.CreatedAt)
		} else if opts.Target == SortByModRevision {
			updateRev(m.UpdatedAt)
		}

		return f(m.Key, fakeRev)
	})
}

// ListRev emulates the behavior of etcd collection's ListRev, but doesn't
// reproduce it exactly. The revisions returned are not from the database -
// postgres uses 32-bit transaction ids and doesn't include one for the creating
// transaction of a row. So, we fake a revision id by sorting rows by their
// create/update timestamp and incrementing a fake revision id every time the
// timestamp changes. Note that the etcd implementation always returns the
// create revision, but that only works here if you also sort by the create
// revision.
func (c *postgresReadOnlyCollection) ListRev(val proto.Message, opts *Options, f func(string, int64) error) error {
	return c.listRev(nil, val, opts, f)
}

// GetRevByIndex is identical to ListRev except that it filters the results
// according to a predicate on the given index.
func (c *postgresReadOnlyCollection) GetRevByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(string, int64) error) error {
	if err := c.validateIndex(index); err != nil {
		return err
	}
	return c.listRev(map[string]string{indexFieldName(index): indexVal}, val, opts, f)
}

func (c *postgresReadOnlyCollection) Count() (int64, error) {
	query := fmt.Sprintf("select count(*) from collections.%s", c.table)
	row := c.db.QueryRowContext(c.ctx, query)

	var result int64
	err := row.Scan(&result)
	return result, c.mapSQLError(err, "")
}

func (c *postgresReadOnlyCollection) Watch(opts ...watch.Option) (watch.Watcher, error) {
	options := watch.SumOptions(opts...)

	watcher, err := c.listener.listen(c.tableWatchChannel(), c.template, nil, nil, options)
	if err != nil {
		return nil, err
	}

	go func() {
		// Do a list of the collection to get the initial state
		lastUpdated := time.Time{}
		val := cloneProtoMsg(c.template)
		if err := c.list(nil, &Options{Target: options.SortTarget, Order: options.SortOrder}, func(m *model) error {
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
			watcher.listener.unregister(watcher)
			return
		}

		// Forward all buffered notifications until the watcher is closed
		watcher.forwardNotifications(c.ctx, lastUpdated)
	}()

	return watcher, nil
}

func (c *postgresReadOnlyCollection) WatchF(f func(*watch.Event) error, opts ...watch.Option) error {
	watcher, err := c.Watch(opts...)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchF(c.ctx, watcher, f)
}

func (c *postgresReadOnlyCollection) WatchOne(key string, opts ...watch.Option) (watch.Watcher, error) {
	options := watch.SumOptions(opts...)

	watcher, err := c.listener.listen(c.indexWatchChannel("key", key), c.template, nil, nil, options)
	if err != nil {
		return nil, err
	}

	go func() {
		// Load the initial state of the row
		lastUpdated := time.Time{}
		if m, err := c.get(c.ctx, key, c.db); err != nil {
			if !errors.Is(err, ErrNotFound{}) {
				watcher.sendInitial(&watch.Event{Type: watch.EventError, Err: err})
				watcher.listener.unregister(watcher)
				return
			}
		} else {
			lastUpdated = m.UpdatedAt
			if err := watcher.sendInitial(&watch.Event{
				Key:      []byte(key),
				Value:    m.Proto,
				Type:     watch.EventPut,
				Template: c.template,
			}); err != nil {
				watcher.listener.unregister(watcher)
				return
			}
		}

		// Forward all buffered notifications until the watcher is closed
		watcher.forwardNotifications(c.ctx, lastUpdated)
	}()
	return watcher, nil
}

func (c *postgresReadOnlyCollection) WatchOneF(key string, f func(*watch.Event) error, opts ...watch.Option) error {
	watcher, err := c.WatchOne(key, opts...)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchF(c.ctx, watcher, f)
}

func (c *postgresReadOnlyCollection) WatchByIndex(index *Index, indexVal string, opts ...watch.Option) (watch.Watcher, error) {
	if err := c.validateIndex(index); err != nil {
		return nil, err
	}

	options := watch.SumOptions(opts...)

	channelName := c.indexWatchChannel(indexFieldName(index), indexVal)
	watcher, err := c.listener.listen(channelName, c.template, nil, nil, options)
	if err != nil {
		return nil, err
	}

	go func() {
		// Do a list of the collection to get the initial state
		lastUpdated := time.Time{}
		val := cloneProtoMsg(c.template)
		opts := &Options{Target: options.SortTarget, Order: options.SortOrder}
		withFields := map[string]string{indexFieldName(index): indexVal}
		if err := c.list(withFields, opts, func(m *model) error {
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
			watcher.listener.unregister(watcher)
			return
		}

		// Forward all buffered notifications until the watcher is closed
		watcher.forwardNotifications(c.ctx, lastUpdated)
	}()

	return watcher, nil
}

func (c *postgresReadOnlyCollection) WatchByIndexF(index *Index, indexVal string, f func(*watch.Event) error, opts ...watch.Option) error {
	watcher, err := c.WatchByIndex(index, indexVal, opts...)
	if err != nil {
		return err
	}
	defer watcher.Close()
	return watchF(c.ctx, watcher, f)
}

func (c *postgresReadOnlyCollection) TTL(key string) (int64, error) {
	return 0, errors.New("TTL is not supported in postgres collections")
}

type postgresReadWriteCollection struct {
	*postgresCollection
	tx *sqlx.Tx
}

func (c *postgresReadWriteCollection) Get(key string, val proto.Message) error {
	result, err := c.get(context.Background(), key, c.tx)
	if err != nil {
		return err
	}
	return errors.EnsureStack(proto.Unmarshal(result.Proto, val))
}

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

	for _, idx := range c.indexes {
		params[indexFieldName(idx)] = idx.Extract(val)
	}

	return params, nil
}

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

	query := fmt.Sprintf("update collections.%s set %s where key = :key", c.table, strings.Join(updateFields, ", "))

	_, err = c.tx.NamedExec(query, params)
	return c.mapSQLError(err, key)
}

func (c *postgresReadWriteCollection) insert(key string, val proto.Message, upsert bool) error {
	if c.keyCheck != nil {
		if err := c.keyCheck(key); err != nil {
			return err
		}
	}

	paramMap, err := c.getWriteParams(key, val)
	if err != nil {
		return err
	}

	columns := []string{}
	paramNames := []string{}
	params := []interface{}{}
	for k, v := range paramMap {
		columns = append(columns, k)
		paramNames = append(paramNames, fmt.Sprintf("$%d", len(paramNames)+1))
		params = append(params, v)
	}
	columnList := strings.Join(columns, ", ")
	paramList := strings.Join(paramNames, ", ")

	query := fmt.Sprintf("insert into collections.%s (%s) values (%s)", c.table, columnList, paramList)
	if upsert {
		query = fmt.Sprintf("%s on conflict (key) do update set (%s) = (%s)", query, columnList, paramList)
	} else {
		// On a normal insert, an error would invalidate the transaction, so do
		// nothing and check the number of rows affected afterwards.
		query += " on conflict do nothing"
	}

	result, err := c.tx.Exec(query, params...)
	if err != nil {
		return c.mapSQLError(err, key)
	}

	if !upsert {
		count, err := result.RowsAffected()
		if err != nil {
			return c.mapSQLError(err, key)
		}

		if count != int64(1) {
			return errors.WithStack(ErrExists{c.table, key})
		}
	}
	return nil
}

func (c *postgresReadWriteCollection) Upsert(key string, val proto.Message, f func() error) error {
	if err := c.Get(key, val); err != nil && !IsErrNotFound(err) {
		return err
	}
	if err := f(); err != nil {
		return err
	}
	return c.Put(key, val)
}

func (c *postgresReadWriteCollection) Create(key string, val proto.Message) error {
	return c.insert(key, val, false)
}

func (c *postgresReadWriteCollection) Delete(key string) error {
	query := fmt.Sprintf("delete from collections.%s where key = $1", c.table)
	res, err := c.tx.Exec(query, key)
	if err != nil {
		return c.mapSQLError(err, key)
	}

	if count, err := res.RowsAffected(); err != nil {
		return c.mapSQLError(err, key)
	} else if count == 0 {
		return errors.WithStack(ErrNotFound{c.table, key})
	}
	return nil
}

func (c *postgresReadWriteCollection) DeleteAll() error {
	query := fmt.Sprintf("delete from collections.%s", c.table)
	_, err := c.tx.Exec(query)
	return c.mapSQLError(err, "")
}

func (c *postgresReadWriteCollection) DeleteByIndex(index *Index, indexVal string) error {
	if err := c.validateIndex(index); err != nil {
		return err
	}
	query := fmt.Sprintf("delete from collections.%s where %s = $1", c.table, indexFieldName(index))
	_, err := c.tx.Exec(query, indexVal)
	return c.mapSQLError(err, "")
}

func (c *postgresReadWriteCollection) TTL(key string) (int64, error) {
	return 0, errors.New("TTL is not supported in postgres collections")
}

func (c *postgresReadWriteCollection) PutTTL(key string, val proto.Message, ttl int64) error {
	return errors.New("PutTTL is not supported in postgres collections")
}
