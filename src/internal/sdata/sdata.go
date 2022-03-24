package sdata

import (
	"database/sql"
	"io"
	"reflect"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Tuple is an alias for []interface{}.
// It is used for passing around rows of data.
// The elements of a tuple will always be pointers so the Tuple can
// be passed to sql.Rows.Scan
type Tuple = []interface{}

// TupleWriter is the type of Writers for structured data.
type TupleWriter interface {
	WriteTuple(row Tuple) error
	Flush() error
}

// TupleReader is a stream of Tuples
type TupleReader interface {
	// Next attempts to read one Tuple into x.
	// If the next data is the wrong shape for x then an error is returned.
	Next(x Tuple) error
}

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
	row := NewTupleFromSQL(cTypes)
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

// NewTupleFromSQL returns a new Tuple based on column types from the sql package.
func NewTupleFromSQL(colTypes []*sql.ColumnType) Tuple {
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

// Copy copies a tuple from w to r.  Row is used to indicate the correct shape of read data.
func Copy(w TupleWriter, r TupleReader, row Tuple) (n int, _ error) {
	for {
		err := r.Next(row)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return n, errors.EnsureStack(err)
		}
		if err := w.WriteTuple(row); err != nil {
			return n, errors.EnsureStack(err)
		}
		n++
	}
	return n, nil
}
