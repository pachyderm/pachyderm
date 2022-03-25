package pachsql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type TableInfo struct {
	Columns []ColumnInfo
}

type ColumnInfo struct {
	TableSchema string `db:"table_schema"`
	TableName   string `db:"table_name"`
	DataType    string `db:"data_type"`
	IsNullable  bool   `db:"is_nullable"`
}

func GetTableInfo(ctx context.Context, db *DB, tableName string) (*TableInfo, error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	ti, err := GetTableInfoTx(tx, tableName)
	if err != nil {
		return nil, err
	}
	return ti, tx.Rollback()
}

// GetTableInfoTx looks up information about the table using INFORMATION_SCHEMA
func GetTableInfoTx(tx *Tx, tableName string) (*TableInfo, error) {
	var schemaName string
	if parts := strings.SplitN(schemaName, "", 2); len(parts) == 2 {
		schemaName = parts[0]
		tableName = parts[1]
	}
	var cinfos []ColumnInfo
	q := `SELECT table_schema, table_name, data_type, is_nullable
		FROM INFORMATION_SCHEMA.columns
	`
	q += makeWhereClause(tx.DriverName())
	if err := tx.Select(&cinfos, q, schemaName, tableName); err != nil {
		return nil, err
	}
	return &TableInfo{Columns: cinfos}, nil
}

func makeWhereClause(driver string) string {
	return fmt.Sprintf(`WHERE table_schema = %s AND = table_name = %s`, Placeholder(driver, 0), Placeholder(driver, 1))
}
