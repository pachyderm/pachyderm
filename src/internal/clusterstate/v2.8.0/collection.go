package v2_8_0

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type index struct {
	Name    string
	Extract func(val proto.Message) string
}

type postgresCollection struct {
	table    string
	indexes  []*index
	keyGen   func(interface{}) (string, error)
	keyCheck func(string) error
	notFound func(interface{}) string
	exists   func(interface{}) string
}

type postgresReadWriteCollection struct {
	*postgresCollection
	tx *pachsql.Tx
}

type colOption func(collection *postgresCollection)

func indexFieldName(idx *index) string {
	return "idx_" + idx.Name
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

// ErrNotFound indicates that a key was not found when it was expected to
// exist.
type ErrNotFound struct {
	Type          string
	Key           string
	customMessage string
}

func (err ErrNotFound) Is(other error) bool {
	_, ok := other.(ErrNotFound)
	return ok
}

func (err ErrNotFound) Error() string {
	if err.customMessage != "" {
		return err.customMessage
	}
	return fmt.Sprintf("%s %s not found", strings.TrimPrefix(err.Type, defaultPrefix), err.Key)
}

func (e ErrNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, e.Error())
}

// IsErrNotFound determines if an error is an ErrNotFound error
func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound{})
}

// ErrExists indicates that a key was found to exist when it was expected not
// to.
type ErrExists struct {
	Type          string
	Key           string
	customMessage string
}

func (err ErrExists) Is(other error) bool {
	_, ok := other.(ErrExists)
	return ok
}

const defaultPrefix string = "pachyderm/1.7.0"

func (err ErrExists) Error() string {
	if err.customMessage != "" {
		return err.customMessage
	}
	return fmt.Sprintf("%s %s already exists", strings.TrimPrefix(err.Type, defaultPrefix), err.Key)
}

func (e ErrExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, e.Error())
}

// IsErrExists determines if an error is an ErrExists error
func IsErrExists(err error) bool {
	return errors.Is(err, ErrExists{})
}

func (c *postgresCollection) withKey(key interface{}, query func(string) error) error {
	if str, ok := key.(string); ok {
		return query(str)
	}
	rawKey, err := c.keyGen(key)
	if err != nil {
		return err
	}
	err = query(rawKey)
	if err != nil {
		var notFound ErrNotFound
		var exists ErrExists
		if c.notFound != nil && errors.As(err, &notFound) {
			notFound.customMessage = c.notFound(key)
			return errors.EnsureStack(notFound)
		}
		if c.exists != nil && errors.As(err, &exists) {
			exists.customMessage = c.exists(key)
			return errors.EnsureStack(exists)
		}
		return err
	}
	return nil
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
			return errors.WithStack(ErrNotFound{Type: c.table, Key: key})
		} else if isDuplicateKeyError(err) {
			return errors.WithStack(ErrExists{Type: c.table, Key: key})
		}
		return errors.EnsureStack(err)
	}
	return nil
}

func migratePostgreSQLCollection(ctx context.Context, tx *pachsql.Tx, name string, indices []*index, oldVal proto.Message, f func(oldKey string) (newKey string, newVal proto.Message, err error), opts ...colOption) (retErr error) {
	var col = postgresReadWriteCollection{
		postgresCollection: &postgresCollection{
			table:   name,
			indexes: indices,
		},
		tx: tx,
	}
	for _, o := range opts {
		o(col.postgresCollection)
	}
	log.Info(ctx, fmt.Sprintf("Retrieving rows from collection.%s", name))
	rr, err := tx.QueryContext(ctx, fmt.Sprintf(`SELECT key, proto FROM collections.%s`, name))
	if err != nil {
		return errors.Wrap(err, "could not read table")
	}
	defer errors.Close(&retErr, rr, "close collections select")
	type pair struct {
		key string
		val proto.Message
	}
	var vals = make(map[string]pair)
	for rr.Next() {
		var (
			oldKey string
			pb     []byte
		)
		if err := rr.Err(); err != nil {
			return errors.Wrap(err, "row error")
		}
		if err := rr.Scan(&oldKey, &pb); err != nil {
			return errors.Wrap(err, "could not scan row")
		}
		if err := proto.Unmarshal(pb, oldVal); err != nil {
			return errors.Wrapf(err, "could not unmarshal proto")
		}
		newKey, newVal, err := f(oldKey)
		if err != nil {
			return errors.Wrapf(err, "could not convert %q", oldKey)
		}
		vals[oldKey] = pair{newKey, proto.Clone(newVal)}

	}
	log.Info(ctx, fmt.Sprintf("Migrating collection.%s", name))
	i := 0
	for oldKey, pair := range vals {
		if err := col.withKey(oldKey, func(oldKey string) error {
			return col.withKey(pair.key, func(newKey string) error {
				params, err := col.getWriteParams(newKey, pair.val)
				if err != nil {
					return err
				}

				updateFields := []string{}
				for k := range params {
					updateFields = append(updateFields, fmt.Sprintf("%s = :%s", k, k))
				}
				params["oldKey"] = oldKey

				query := fmt.Sprintf("update collections.%s set %s where key = :oldKey", col.table, strings.Join(updateFields, ", "))

				_, err = col.tx.NamedExecContext(ctx, query, params)
				i++
				log.Info(ctx, fmt.Sprintf("Migrating collection.%s", name), zap.String("old", oldKey), zap.String("new", newKey), zap.String("progress", fmt.Sprintf("%d/%d", i, len(vals))))
				return col.mapSQLError(err, oldKey)
			})
		}); err != nil {
			return errors.Wrapf(err, "could not update %q to %q", oldKey, pair.key)
		}
	}
	return nil
}

func setupPostgresCollections(ctx context.Context, sqlTx *pachsql.Tx, collections ...*postgresCollection) error {
	for _, col := range collections {
		columns := []string{
			"createdat timestamp with time zone default current_timestamp",
			"updatedat timestamp with time zone default current_timestamp",
			"proto bytea",
			"version text",
			"key text primary key",
		}

		indexFields := []string{"'key'"}
		for _, idx := range col.indexes {
			name := indexFieldName(idx)
			columns = append(columns, name+" text")
			indexFields = append(indexFields, "'"+name+"'")
		}

		log.Info(ctx, fmt.Sprintf("Creating collections.%s table", col.table))
		createTable := fmt.Sprintf("create table collections.%s (%s);", col.table, strings.Join(columns, ", "))
		if _, err := sqlTx.Exec(createTable); err != nil {
			return errors.EnsureStack(err)
		}

		for _, idx := range col.indexes {
			createIndex := fmt.Sprintf("create index on collections.%s (%s);", col.table, indexFieldName(idx))
			if _, err := sqlTx.Exec(createIndex); err != nil {
				return errors.EnsureStack(err)
			}
		}

		updatedatTrigger := fmt.Sprintf(`
	create trigger updatedat_trigger
		before insert or update or delete on collections.%s
		for each row execute procedure collections.updatedat_trigger_fn();
	`, col.table)
		if _, err := sqlTx.ExecContext(ctx, updatedatTrigger); err != nil {
			return errors.EnsureStack(err)
		}

		notifyTrigger := fmt.Sprintf(`
	create trigger notify_trigger
		after insert or update or delete on collections.%s
		for each row execute procedure collections.notify_trigger_fn(%s);
	`, col.table, strings.Join(indexFields, ", "))
		if _, err := sqlTx.ExecContext(ctx, notifyTrigger); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// newPostgresCollection creates a new collection backed by postgres.
func newPostgresCollection(name string) *postgresCollection {
	col := &postgresCollection{
		table: name,
	}
	return col
}
