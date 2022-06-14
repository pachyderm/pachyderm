package server

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

type page string

const (
	firstPage  page = "firstPage"
	middlePage page = "middlePage"
	lastPage   page = "lastPage"
	pageSize        = 500
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
	if err := s.collectDatabaseActivities(ctx, tw, prefix...); err != nil {
		return err
	}
	if err := s.collectDatabaseRowCount(ctx, tw, prefix...); err != nil {
		return err
	}
	if err := s.collectDatabaseTableSizes(ctx, tw, prefix...); err != nil {
		return err
	}
	return nil
}

func (s *debugServer) collectDatabaseActivities(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	return s.collectDatabaseQuery(ctx, tw, "activities", s.queryDatabaseActivities, prefix...)
}

func (s *debugServer) collectDatabaseRowCount(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	return s.collectDatabaseQuery(ctx, tw, "row-counts", s.queryDatabaseRowCounts, prefix...)
}

func (s *debugServer) collectDatabaseTableSizes(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	return s.collectDatabaseQuery(ctx, tw, "table-sizes", s.queryDatabaseTableSizes, prefix...)
}

func (s *debugServer) collectDatabaseQuery(ctx context.Context, tw *tar.Writer, fileName string,
	queryFunc func(ctx context.Context, w io.Writer) error, prefix ...string) error {
	return collectDebugFile(tw, fileName, "json", func(w io.Writer) error {
		if err := queryFunc(ctx, w); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}, prefix...)
}

func (s *debugServer) queryDatabaseRowCounts(ctx context.Context, w io.Writer) error {
	query := `
		SELECT schemaname, relname, n_live_tup, seq_scan, idx_scan
		FROM pg_stat_user_tables
		ORDER BY schemaname, relname;
	`
	return s.writeQueryResponseToJSON(ctx, query, w)
}

func (s *debugServer) queryDatabaseActivities(ctx context.Context, w io.Writer) error {
	query := `
		SELECT current_timestamp - query_start as runtime, datname, usename, client_addr, query
		FROM pg_stat_activity
		WHERE state != 'idle'
		ORDER by runtime DESC;
	`
	return s.writeQueryResponseToJSON(ctx, query, w)
}

func (s *debugServer) queryDatabaseTableSizes(ctx context.Context, w io.Writer) error {
	query := `
		SELECT nspname AS "schemaname", relname, pg_total_relation_size(C.oid) AS "total_size"
		FROM pg_class C
		  INNER JOIN pg_namespace N ON (N.oid = C.relnamespace)
		WHERE nspname NOT IN ('pg_catalog', 'information_schema')
		  AND C.relkind <> 'i'
		  AND nspname !~ '^pg_toast'
		ORDER BY nspname, relname;
	`
	return s.writeQueryResponseToJSON(ctx, query, w)
}

func (s *debugServer) collectDatabaseTables(ctx context.Context, tw *tar.Writer, prefix ...string) error {
	tables, err := pachsql.ListTables(ctx, s.database)
	if err != nil {
		return errors.EnsureStack(err)
	}
	for _, table := range tables {
		table := table // intentional shadow to safely pass by pointer
		fullPrefix := strings.Join(append(append([]string{}, prefix...), "tables", table.SchemaName), "/")
		if err := collectDebugFile(tw, table.TableName, "json", func(w io.Writer) error {
			if err := s.collectTable(ctx, w, &table); err != nil {
				return errors.EnsureStack(err)
			}
			return nil
		}, fullPrefix); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

func (s *debugServer) collectTable(ctx context.Context, w io.Writer, table *pachsql.SchemaTable) error {
	column, err := pachsql.GetMainColumn(ctx, s.database, table)
	if err != nil {
		return errors.EnsureStack(err)
	}
	rows, err := s.getFirstPage(ctx, column, pageSize, table)
	if err != nil {
		return errors.EnsureStack(err)
	}
	if len(rows) == 0 {
		return writeAllRowsToJSON(rows, w)
	} else if err := writeRowPageToJSON(rows, w, firstPage); err != nil {
		return errors.EnsureStack(err)
	}
	queryTemplate := fmt.Sprintf(`
		SELECT * FROM %s.%s WHERE %s > {{lastitem}} 
		ORDER BY %s FETCH FIRST %d ROWS ONLY;`,
		table.SchemaName, table.TableName, column, column, pageSize)
	toString := setToStringBasedOnType(rows[len(rows)-1][column])
	for len(rows) == pageSize {
		lastItem := rows[len(rows)-1][column]
		query := strings.ReplaceAll(queryTemplate, "{{lastitem}}", toString(lastItem))
		rows, err = pachsql.GetScannedRows(ctx, s.database, query)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				// truncate the table that was being dumped.
				return writeRowPageToJSON(rows, w, lastPage)
			}
			return errors.EnsureStack(err)
		}
		if len(rows) == 0 {
			break
		}
		if err := writeRowPageToJSON(rows, w, middlePage); err != nil {
			lastPageErr := writeRowPageToJSON(rows, w, lastPage)
			return errors.New(fmt.Sprintf("error writing page: %s: "+
				"error writing last page: %s", err.Error(), lastPageErr.Error()))
		}
	}
	return writeRowPageToJSON(rows, w, lastPage)
}

func (s *debugServer) getFirstPage(ctx context.Context, column string, pageSize int, table *pachsql.SchemaTable) ([]pachsql.RowMap, error) {
	query := fmt.Sprintf(`
		SELECT * FROM %s.%s
		ORDER BY %s 
		FETCH FIRST %d ROWS ONLY;
	`, table.SchemaName, table.TableName, column, pageSize)
	pageRows, err := pachsql.GetScannedRows(ctx, s.database, query)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return pageRows, nil
}

func setToStringBasedOnType(item interface{}) func(item interface{}) string {
	switch fmt.Sprint(reflect.TypeOf(item)) {
	case "time.Time":
		return func(item interface{}) string {
			asTime, ok := item.(time.Time)
			if !ok {
				return ""
			}
			formattedTime := asTime.Format("2006-01-2 15:04:05.000000 -07:00")
			return fmt.Sprintf("'%v'", formattedTime)
		}
	case "[]uint8":
		return func(item interface{}) string {
			asByteArray, ok := item.([]uint8)
			if !ok {
				return ""
			}
			asString := ""
			for _, num := range asByteArray {
				asString += fmt.Sprintf("%02X", num)
			}
			return fmt.Sprintf("byteain('\\x%s')", asString)
		}
	default:
		return func(item interface{}) string {
			return fmt.Sprintf("'%v'", item)
		}
	}
}

func (s *debugServer) writeQueryResponseToJSON(ctx context.Context, query string, w io.Writer) error {
	scannedRows, err := pachsql.GetScannedRows(ctx, s.database, query)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return writeAllRowsToJSON(scannedRows, w)
}

func writeAllRowsToJSON(rowMap []pachsql.RowMap, w io.Writer) error {
	return writeRowsToJSON(rowMap, w, "[{", "[\n\t{", "}]", "}\n]\n", "},", "},\n\t")
}

func writeRowPageToJSON(rowMap []pachsql.RowMap, w io.Writer, p page) error {
	var replacerOpts []string
	switch p {
	case firstPage:
		replacerOpts = []string{"[{", "[\n\t{", "}]", "}", "},", "},\n\t"}
	case middlePage:
		replacerOpts = []string{"[{", ",\n\t{", "}]", "}", "},", "},\n\t"}
	case lastPage: // in this case, just write the end of the dictionary
		if _, err := w.Write([]byte("\n]")); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	default:
		return errors.New("page should be either firstPage, middlePage, or lastPage")
	}
	return writeRowsToJSON(rowMap, w, replacerOpts...)
}

func writeRowsToJSON(rowMap []pachsql.RowMap, w io.Writer, opts ...string) error {
	marshalledRows, err := json.Marshal(rowMap)
	if err != nil {
		return errors.EnsureStack(err)
	}
	formatter := strings.NewReplacer(opts...)
	formattedJson := []byte(formatter.Replace(string(marshalledRows)))
	if _, err = w.Write(formattedJson); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}
