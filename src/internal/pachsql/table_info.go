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
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: false}) // TODO Snowflake doesn't support ReadOnly
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
	q := `SELECT table_schema, table_name, data_type, is_nullable
		FROM INFORMATION_SCHEMA.columns
	`
	var args []interface{}
	// table_name
	q += fmt.Sprintf(" WHERE lower(table_name) = lower(%s)", Placeholder(tx.DriverName(), len(args)))
	args = append(args, tableName)
	// schema_name
	if schemaName != "" {
		q += " AND lower(table_schema) = " + fmt.Sprintf("lower(%s)", Placeholder(tx.DriverName(), len(args)))
		args = append(args, schemaName)
	}
	// We use tx.Query, not tx.Select here because MySQL and Postgres have conflicting capitalization
	// and sqlx complains about scanning using struct tags.
	rows, err := tx.Query(q, args...)
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
