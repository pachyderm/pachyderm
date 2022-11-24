package collection

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// MigratePostgreSQLCollection_v2_4_0 is used to migrate a PostgreSQL collection.
// Given a collection named name, with the given indices and options, it will
// fetch each row and call f on it, then update the current row with the new
// key, value and indices, all in one transaction which caller is responsible
// for committing.
//
// # DO NOT MODIFY THIS FUNCTION
//
// IT HAS BEEN USED IN A RELEASED MIGRATION
func MigratePostgreSQLCollection_v2_4_0(ctx context.Context, tx *pachsql.Tx, name string, indices []*Index, oldVal proto.Message, f func(oldKey string) (newKey string, newVal proto.Message, err error), opts ...Option) error {
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
				return col.mapSQLError(err, oldKey)
			})
		}); err != nil {
			return errors.Wrapf(err, "could not update %q to %q", oldKey, pair.key)
		}
	}
	return nil
}
