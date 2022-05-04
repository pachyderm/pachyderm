package sdata

import (
	"database/sql"
	"io"
	"reflect"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
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
	ColumnNames []string
	RowCount    uint64
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
	row, err := NewTupleFromColumnTypes(cTypes)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	var count uint64
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
		ColumnNames: colNames,
		RowCount:    count,
	}, nil
}

func NewTupleFromColumnTypes(cTypes []*sql.ColumnType) (Tuple, error) {
	row := make(Tuple, len(cTypes))
	for i, cType := range cTypes {
		dbType := cType.DatabaseTypeName()
		nullable, ok := cType.Nullable()
		if !ok {
			nullable = true
		}
		var err error
		row[i], err = makeTupleElement(dbType, nullable)
		if err != nil {
			return nil, err
		}
	}
	return row, nil
}

// Copy copies a tuple from r to w. Row is used to indicate the correct shape of read data.
func Copy(w TupleWriter, r TupleReader, row Tuple) (n int, _ error) {
	for {
		err := r.Next(row)
		if errors.Is(err, io.EOF) {
			w.Flush()
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

func NewTupleFromTableInfo(info *pachsql.TableInfo) (Tuple, error) {
	tuple := make(Tuple, len(info.Columns))
	for i, ci := range info.Columns {
		var err error
		tuple[i], err = makeTupleElement(ci.DataType, ci.IsNullable)
		if err != nil {
			return nil, err
		}
	}
	return tuple, nil
}

func makeTupleElement(dbType string, nullable bool) (interface{}, error) {
	switch dbType {
	case "BOOL", "BOOLEAN":
		if nullable {
			return new(sql.NullBool), nil
		} else {
			return new(bool), nil
		}
	case "SMALLINT", "INT2":
		if nullable {
			return new(sql.NullInt16), nil
		} else {
			return new(int16), nil
		}
	case "INTEGER", "INT", "INT4":
		if nullable {
			return new(sql.NullInt32), nil
		} else {
			return new(int32), nil
		}
	// TODO "NUMBER" type from Snowflake can vary in precision, but default to int64 for now.
	// https://docs.snowflake.com/en/sql-reference/data-types-numeric.html#number
	case "BIGINT", "INT8", "NUMBER":
		if nullable {
			return new(sql.NullInt64), nil
		} else {
			return new(int64), nil
		}
	// TODO account for precision and scale as well
	case "FLOAT", "FLOAT8", "REAL", "DOUBLE PRECISION", "FIXED":
		if nullable {
			return new(sql.NullFloat64), nil
		} else {
			return new(float64), nil
		}
	case "VARCHAR", "TEXT", "CHARACTER VARYING":
		if nullable {
			return new(sql.NullString), nil
		} else {
			return new(string), nil
		}
	case "DATE", "TIMESTAMP", "TIMESTAMP_NTZ", "TIMESTAMP WITHOUT TIME ZONE":
		if nullable {
			return new(sql.NullTime), nil
		} else {
			return new(time.Time), nil
		}
	default:
		return nil, errors.Errorf("unrecognized type: %v", dbType)
	}
}

// CloneTuple uses Go reflection to make a copy of a Tuple.
func CloneTuple(t Tuple) Tuple {
	newTuple := make(Tuple, len(t))
	for i := range t {
		v := reflect.New(reflect.TypeOf(t[i]).Elem())
		v.Elem().Set(reflect.ValueOf(t[i]).Elem())
		newTuple[i] = v.Interface()
	}
	return newTuple
}
