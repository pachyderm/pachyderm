package v2_5_0

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/version"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
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

const DefaultPrefix string = "pachyderm/1.7.0"

func (err ErrNotFound) Error() string {
	if err.customMessage != "" {
		return err.customMessage
	}
	return fmt.Sprintf("%s %s not found", strings.TrimPrefix(err.Type, DefaultPrefix), err.Key)
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

func (err ErrExists) Error() string {
	if err.customMessage != "" {
		return err.customMessage
	}
	return fmt.Sprintf("%s %s already exists", strings.TrimPrefix(err.Type, DefaultPrefix), err.Key)
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

func withKeyCheck(check func(string) error) colOption {
	return func(c *postgresCollection) {
		c.keyCheck = check
	}
}

func withKeyGen(gen func(interface{}) (string, error)) colOption {
	return func(c *postgresCollection) {
		c.keyGen = gen
	}
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
func newPostgresCollection(name string, indexes []*index, opts ...colOption) *postgresCollection {
	col := &postgresCollection{
		table:   name,
		indexes: indexes,
	}
	for _, opt := range opts {
		opt(col)
	}
	return col
}

func migratePostgreSQLCollection(ctx context.Context, tx *pachsql.Tx, name string, indices []*index, oldVal proto.Message, f func(oldKey string) (newKey string, newVal proto.Message, err error), opts ...colOption) error {
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
	rr, err := tx.QueryContext(ctx, fmt.Sprintf(`SELECT key, proto FROM collections.%s`, name))
	if err != nil {
		return errors.Wrap(err, "could not read table")
	}
	defer rr.Close()
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

	// Optimized code for building the UPDATEs and sending them to Postgres in batches
	// NOTE: This code necessitated upgrading from sqlx 1.2.0 to sqlx 1.3.x to get batching support for named variables

	log.Info(ctx, fmt.Sprintf("Migrating %s [%d rows]", name, len(vals)))

	batch_size := 1000 // Maximum size of each batch
	batch_count := 0   // Incrementing count of the number of rows in the current batch
	total_count := 0   // Incrementing count of the total number of rows
	query := ""        // String to hold the UPDATE query

	// Create an array that will contain the substring of the SET clause for each column in the UPDATE statement.
	// For example, the substring for the "idx_name" column would be "idx_name = :idx_name"
	// In that substring, "idx_name" is the column name, and ":idx_name" is a placeholder for the actual value.
	columnSubstrings := []string{}

	// Loop through each row that needs to be updated
	for oldKey, pair := range vals {

		// These two function calls handle updating an oldKey (with no project reference) to a newKey (with project reference)
		if err := col.withKey(oldKey, func(oldKey string) error {
			return col.withKey(pair.key, func(newKey string) error {

				// Increment our counters
				batch_count++
				total_count++

				// Get the parameters for the current row
				params, err := col.getWriteParams(pair.key, pair.val)

				if err != nil {
					return err
				}

				// Create the query string. This only needs to be done on the first time through, as only the values change with each row,
				// not the query string itself.  The query string has placeholders for the values that change from one row to the next.
				if total_count == 1 {

					// Loop through each parameter in the current row
					for k := range params {
						// Add the UPDATE/SET substring for this column
						columnSubstrings = append(columnSubstrings, fmt.Sprintf("%s = :%s", k, k))
					}

					query = fmt.Sprintf("update collections.%s set %s where key = :oldKey", col.table, strings.Join(columnSubstrings, ", "))
				}

				// Add the old key to the parameters. There isn't a column for it (which is why it isn't in params[] to begin with), by it
				// gets used in the WHERE clause.
				params["oldKey"] = oldKey

				// Add the params object to the batch of parameters that we are building
				param_batch := []any{}
				param_batch = append(param_batch, params)

				// If we've reached the batch limit, or if we're on the last row, send the updates to Postgres
				if batch_count == batch_size || total_count == len(vals) {

					// Execute the UPDATE query with the collected batch of parameters
					_, err = tx.NamedExecContext(ctx, query, param_batch)

					if err != nil {
						return col.mapSQLError(err, oldKey)
					}

					log.Debug(ctx, "Sent batch", zap.String("Sent", fmt.Sprintf("%d", batch_count)), zap.String("Progress", fmt.Sprintf("%d/%d", total_count, len(vals))))

					// Reset the count in preparation for the next batch
					batch_count = 0
				}

				return col.mapSQLError(err, oldKey)
			})
		}); err != nil {
			return errors.Wrapf(err, "could not update %q to %q", oldKey, pair.key)
		}

	}

	return nil
}
