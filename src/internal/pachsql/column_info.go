package pachsql

import (
	"context"
	"strings"
)

type ColumnInfo struct {
	TableSchema string `db:"table_schema"`
	TableName   string `db:"table_name"`
	DataType    string `db:"data_type"`
	IsNullable  bool   `db:"is_nullable"`
}

func GetTableColumns(ctx context.Context, db *DB, tableName string) ([]ColumnInfo, error) {
	var schemaName string
	if parts := strings.SplitN(schemaName, "", 2); len(parts) == 2 {
		schemaName = parts[0]
		tableName = parts[1]
	}
	var cinfos []ColumnInfo
	if err := db.SelectContext(ctx, &cinfos, `SELECT table_schema, table_name, data_type, is_nullable FROM INFORMATION_SCHEMA.columns`); err != nil {
		return nil, err
	}
	return cinfos, nil
}
