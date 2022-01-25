package sdata

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
		return nil, errors.EnsureStack(err)
	}
	cTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	var dbTypes []string
	for _, cType := range cTypes {
		dbTypes = append(dbTypes, cType.DatabaseTypeName())
	}
	var count uint64
	row := newTupleFromSQL(cTypes)
	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			return nil, errors.EnsureStack(err)
		}
		if err := tw.WriteTuple(row); err != nil {
			return nil, errors.EnsureStack(err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	if err := tw.Flush(); err != nil {
		return nil, errors.EnsureStack(err)
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
		v := reflect.New(rType).Interface()
		if nullable, ok := cType.Nullable(); !ok || nullable {
			switch v.(type) {
			case *int16:
				v = &sql.NullInt16{}
			case *int32:
				v = &sql.NullInt32{}
			case *int64:
				v = &sql.NullInt64{}
			case *float64:
				v = &sql.NullFloat64{}
			case *string:
				v = &sql.NullString{}
			case *time.Time:
				v = &sql.NullTime{}
			}
		}
		row[i] = v
	}
	return row
}
