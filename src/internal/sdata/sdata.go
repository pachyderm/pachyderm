package sdata

import (
	"database/sql"
	"reflect"
)

// MaterializationResult is returned by MaterializeSQL
type MaterializationResult struct {
	ColumnNames   []string
	ColumnDBTypes []string
	RowCount      uint64
}

// MaterializeSQL reads all the rows from a *sql.Rows, and writes them to tw.
// It flushes tw and returns a MaterializationResult
func MaterializeSQL(tw TupleWriter, rows *sql.Rows) (*MaterializationResult, error) {
	colNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	cTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	var dbTypes []string
	for _, cType := range cTypes {
		dbTypes = append(dbTypes, cType.DatabaseTypeName())
	}
	var count uint64
	row := newTupleFromSQL(cTypes)
	for rows.Next() {
		if err := readTuple(row, rows); err != nil {
			return nil, err
		}
		if err := tw.WriteTuple(row); err != nil {
			return nil, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tw.Flush(); err != nil {
		return nil, err
	}
	return &MaterializationResult{
		ColumnNames:   colNames,
		ColumnDBTypes: dbTypes,
		RowCount:      count,
	}, nil
}

func newTupleFromSQL(colTypes []*sql.ColumnType) Tuple {
	row := make([]interface{}, len(colTypes))
	for i, cType := range colTypes {
		var rType reflect.Type
		switch cType.DatabaseTypeName() {
		case "VARCHAR", "TEXT":
			// force scan type to be string
			rType = reflect.TypeOf("")
		default:
			rType = cType.ScanType()
		}
		row[i] = reflect.New(rType).Interface()
	}
	return row
}

func readTuple(x Tuple, rows *sql.Rows) error {
	return rows.Scan(x...)
}
