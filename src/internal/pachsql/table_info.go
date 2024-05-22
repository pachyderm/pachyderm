package pachsql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// TableInfo contains information about a SQL table
type TableInfo struct {
	Driver  string
	Name    string
	Schema  string
	Columns []ColumnInfo
}

// ColumnInfo is information about a single column in a SQL table
type ColumnInfo struct {
	Name       string
	DataType   string
	IsNullable bool
}

// GetTableInfo looks up information about the table using INFORMATION_SCHEMA
func GetTableInfo(ctx context.Context, db *DB, tableName string) (*TableInfo, error) {
	readonly := true
	if db.DriverName() == "snowflake" {
		readonly = false
	}
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: readonly}) // TODO Snowflake doesn't support ReadOnly
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer tx.Rollback()
	ti, err := GetTableInfoTx(tx, tableName)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return ti, errors.EnsureStack(tx.Rollback())
}

// GetTableInfoTx looks up information about the table using INFORMATION_SCHEMA
func GetTableInfoTx(tx *Tx, tablePath string) (*TableInfo, error) {
	schemaName, tableName := SplitTableSchema(tx.DriverName(), tablePath)
	if schemaName == "" {
		// Check whether table is unique, and infer schema_name
		if rows, err := tx.Query(fmt.Sprintf(`
		SELECT lower(table_schema) as schema_name
		FROM information_schema.tables
		WHERE lower(table_name) = lower('%s')`, tableName)); err != nil {
			return nil, errors.EnsureStack(err)
		} else {
			count := 0
			for rows.Next() {
				count++
				if count > 1 {
					return nil, errors.Errorf("table %s is not unique, please specify schema name", tableName)
				}
				if err = rows.Scan(&schemaName); err != nil {
					return nil, errors.EnsureStack(err)
				}
			}
			if err = rows.Err(); err != nil {
				return nil, errors.EnsureStack(err)
			}
		}
	}
	q := fmt.Sprintf(`
	SELECT
		column_name,
	    upper(data_type),
	    CASE upper(is_nullable)
			WHEN 'YES' THEN true
		   	ELSE false
		END,
		numeric_precision,
		numeric_scale
	FROM information_schema.columns
	WHERE upper(table_name) = upper('%s') AND upper(table_schema) = upper('%s')
	ORDER BY ordinal_position`, tableName, schemaName)
	// We use tx.Query, not tx.Select here because MySQL and Postgres have conflicting capitalization
	// and sqlx complains about scanning using struct tags.
	rows, err := tx.Query(q)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer rows.Close()

	var cinfos []ColumnInfo
	var precision, scale sql.NullInt64
	for rows.Next() {
		var ci ColumnInfo
		if err := rows.Scan(&ci.Name, &ci.DataType, &ci.IsNullable, &precision, &scale); err != nil {
			return nil, errors.EnsureStack(err)
		}
		cinfos = append(cinfos, ci)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &TableInfo{Driver: tx.DriverName(), Name: tableName, Schema: schemaName, Columns: cinfos}, nil
}

func (t *TableInfo) ColumnNames() []string {
	columns := make([]string, len(t.Columns))
	for i := range t.Columns {
		columns[i] = t.Columns[i].Name
	}
	return columns
}
