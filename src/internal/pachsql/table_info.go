package pachsql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// TableInfo contains information about a SQL table
type TableInfo struct {
	Columns []ColumnInfo
}

// ColumnInfo is information about a single column in a SQL table
type ColumnInfo struct {
	TableSchema string
	TableName   string
	DataType    string
	IsNullable  bool
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
	var cinfos []ColumnInfo
	where := fmt.Sprintf("lower(table_name) = lower('%s')", tableName)
	if schemaName != "" {
		where += fmt.Sprintf(" AND lower(table_schema) = lower('%s')", schemaName)
	}
	q := fmt.Sprintf(`
	SELECT table_schema, table_name, data_type, is_nullable
	FROM INFORMATION_SCHEMA.columns
	WHERE %s
	ORDER BY ordinal_position`, where)
	// We use tx.Query, not tx.Select here because MySQL and Postgres have conflicting capitalization
	// and sqlx complains about scanning using struct tags.
	rows, err := tx.Query(q)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	defer rows.Close()
	for rows.Next() {
		var ci ColumnInfo
		var isNullableS string
		if err := rows.Scan(&ci.TableSchema, &ci.TableName, &ci.DataType, &isNullableS); err != nil {
			return nil, errors.EnsureStack(err)
		}
		switch strings.ToLower(isNullableS) {
		case "yes":
			ci.IsNullable = true
		}
		cinfos = append(cinfos, ci)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &TableInfo{Columns: cinfos}, nil
}
