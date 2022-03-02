package pachsql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// SQLTableInfo contains info about a SQL table.
type SQLTableInfo struct {
	Name        string
	ColumnNames []string
	ColumnTypes []*sql.ColumnType
}

// LookupTableInfo looks up column name and type information using INFORMATION_SCHEMA.
func LookupTableInfo(tx *Tx, tableName string) (*SQLTableInfo, error) {
	// Query information schema to get collum
	type infoRow struct {
	}
	infoRows := []infoRow{}
	if err := tx.Select(infoRows, "SELECT * FROM INFORMATION_SCHEMA."); err != nil {
		return nil, err
	}
	colNames := []string{}
	return &SQLTableInfo{
		Name:        tableName,
		ColumnNames: colNames,
	}, nil
}

func MakeInsertStatement(driverName string, tableName string, colNames []string) string {
	return fmt.Sprintf("INSERT INTO %s %s VALUES %s", tableName, formatColumnNames(colNames), formatPlaceHolder(driverName, len(colNames)))
}

func formatColumnNames(colNames []string) string {
	return "(" + strings.Join(colNames, ",") + ")"
}

func formatPlaceHolder(driver string, n int) string {
	var placeholder func(i int) string
	switch driver {
	case "pgx":
		placeholder = func(i int) string { return "$" + strconv.Itoa(i+1) }
	case "mysql":
		placeholder = func(int) string { return "?" }
	default:
		panic("unknown driver " + driver)
	}
	var ret string
	for i := 0; i < n; i++ {
		if i > 0 {
			ret += ", "
		}
		ret += placeholder(i)
	}
	return "(" + ret + ")"
}
