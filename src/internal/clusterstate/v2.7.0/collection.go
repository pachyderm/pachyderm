package v2_7_0

/*
 * This file imports just enough of src/internal/collections to create the
 * cluster defaults table.  It is literally a copy-paste of
 * collections.SetupPostgresCollections, with the supporting infrastructure it
 * requires.
 */

import (
	"context"
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

type index struct {
	Name string
}

type postgresCollection struct {
	table   string
	indexes []*index
}

func indexFieldName(idx *index) string {
	return "idx_" + idx.Name
}

func newPostgresCollection(name string, indexes []*index) *postgresCollection {
	col := &postgresCollection{
		table:   name,
		indexes: indexes,
	}
	return col
}

// This is a copy-paste of collections.SetupPostgresCollections, using local
// types.
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
