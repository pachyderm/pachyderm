package server

import (
	"archive/tar"
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func (s *debugServer) collectDatabaseDump(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	if err := s.collectDatabaseStats(ctxWithTimeout, tw, prefix...); err != nil {
		return err
	}
	if err := s.collectDatabaseTables(ctxWithTimeout, tw, prefix...); err != nil {
		return err
	}
	return nil
}

func (s *debugServer) collectDatabaseStats(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	queries := map[string]string{
		"table-sizes": `
			SELECT nspname AS "schemaname", relname, pg_total_relation_size(C.oid) AS "total_size"
			FROM pg_class C INNER JOIN pg_namespace N ON (N.oid = C.relnamespace)
			WHERE nspname NOT IN ('pg_catalog', 'information_schema') AND C.relkind <> 'i' AND nspname !~ '^pg_toast'
			ORDER BY nspname, relname;
		`,
		"activities": `
			SELECT current_timestamp - query_start as runtime, datname, usename, client_addr, query
			FROM pg_stat_activity WHERE state != 'idle' ORDER BY runtime DESC;
		`,
		"row-counts": `
			SELECT schemaname, relname, n_live_tup, seq_scan, idx_scan
			FROM pg_stat_user_tables ORDER BY schemaname, relname;
		`,
	}
	var errs []error
	for filename, query := range queries {
		if err := collectDebugFile(tw, filename, "json", func(w io.Writer) error {
			rows, err := s.database.QueryContext(ctx, query)
			if err != nil {
				return errors.EnsureStack(err)
			}
			return s.writeRowsToJSON(rows, w)
		}, prefix...); err != nil {
			errs = append(errs, errors.Wrapf(err, "execute query %v", filename))
		}
	}
	if len(errs) > 0 {
		return errors.EnsureStack(fmt.Errorf("%v", errs))
	}
	return nil
}

func (s *debugServer) collectDatabaseTables(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	tables, err := pachsql.ListTables(ctx, s.database)
	if err != nil {
		return errors.EnsureStack(err)
	}
	for _, table := range tables {
		fullPrefix := strings.Join(append(append([]string{}, prefix...), "tables", table.SchemaName), "/")
		if err := collectDebugFile(tw, table.TableName, "json", func(rawWriter io.Writer) error {
			w := bufio.NewWriter(rawWriter)
			if err := s.collectTable(ctx, w, &table); err != nil {
				return errors.Wrapf(err, "collect table %s.%s", table.SchemaName, table.TableName)
			}
			if err := w.Flush(); err != nil {
				return errors.Wrap(err, "flush buffer")
			}
			return nil
		}, fullPrefix); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func (s *debugServer) collectTable(ctx context.Context, w io.Writer, table *pachsql.SchemaTable) error {
	sanitizedTableName := strings.ReplaceAll(table.SchemaName+"."+table.TableName, "'", "''")
	rows, err := s.database.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", sanitizedTableName))
	if err != nil {
		return errors.Wrap(err, "execute query")
	}
	defer rows.Close()
	return s.writeRowsToJSON(rows, w)
}

func (s *debugServer) writeRowsToJSON(rows *sql.Rows, w io.Writer) error {
	if _, err := w.Write([]byte("[\n\t")); err != nil {
		return errors.Wrap(err, "write rows to JSON")
	}
	wroteFirstRow := false
	for rows.Next() {
		if wroteFirstRow {
			if _, err := w.Write([]byte(",\n\t")); err != nil {
				return errors.Wrap(err, "write rows to JSON")
			}
		}
		row := pachsql.RowMap{}
		if err := sqlx.MapScan(rows, row); err != nil {
			return errors.Wrap(err, "map scan rows")
		}
		jsonRow, err := json.Marshal(row)
		if err != nil {
			return errors.Wrap(err, "write rows to JSON")
		}
		if _, err := w.Write(jsonRow); err != nil {
			return errors.Wrap(err, "write rows to JSON")
		}
		wroteFirstRow = true
	}
	if err := rows.Err(); err != nil {
		truncation := fmt.Sprintf("\n\t{\"message\": \"result set was truncated.\", \"error\":%q}\n]",
			err.Error())
		w.Write([]byte(truncation)) //nolint:errcheck
		return nil
	}
	if _, err := w.Write([]byte("\n]")); err != nil {
		return errors.Wrap(err, "write rows to JSON")
	}
	return nil
}
