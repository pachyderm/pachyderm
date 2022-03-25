package sdata

import (
	"context"
	"database/sql"
	"io"
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

// NewTupleFromSQL returns a new Tuple based on column types from the sql package.
func newTupleFromSQL(colTypes []*sql.ColumnType) Tuple {
	row := make([]interface{}, len(colTypes))
	for i, cType := range colTypes {
		var err error
		isNull, nullOk := cType.Nullable()
		// If the driver doesn't support null, then use a nullable type just to be safe
		row[i], err = makeTupleElement(cType.DatabaseTypeName(), isNull || !nullOk)
		if err != nil {
			panic(err)
		}
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

func NewTupleFromTable(ctx context.Context, db *pachsql.DB, tableName string) (Tuple, error) {
	info, err := pachsql.GetTableInfo(ctx, db, tableName)
	if err != nil {
		return nil, err
	}
	tuple := make(Tuple, len(info.Columns))
	for i, ci := range info.Columns {
		var err error
		tuple[i], err = makeTupleElement(ci.DataType, ci.IsNullable)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func makeTupleElement(dbType string, nullable bool) (interface{}, error) {
	switch dbType {
	case "BOOL":
		if nullable {
			return &sql.NullBool{}, nil
		} else {
			return new(bool), nil
		}
	case "SMALLINT", "INT2":
		if nullable {
			return &sql.NullInt16{}, nil
		} else {
			return new(int16), nil
		}
	case "INT", "INT4":
		if nullable {
			return &sql.NullInt32{}, nil
		} else {
			return new(int32), nil
		}
	case "BIGINT", "INT8":
		if nullable {
			return &sql.NullInt64{}, nil
		} else {
			return new(int64), nil
		}
	case "FLOAT", "FLOAT8":
		if nullable {
			return &sql.NullFloat64{}, nil
		} else {
			return new(float64), nil
		}
	case "VARCHAR", "TEXT":
		if nullable {
			return &sql.NullString{}, nil
		} else {
			return new(string), nil
		}
	case "TIMESTAMP":
		if nullable {
			return &sql.NullTime{}, nil
		} else {
			return new(time.Time), nil
		}
	default:
		return nil, errors.Errorf("unrecognized type: %v", dbType)
	}
}
