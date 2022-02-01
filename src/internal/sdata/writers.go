// package sdata deals with structured data.
package sdata

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// ErrTupleFields is returned by writers to indicate that
// a tuple is not the right shape to be written to them.
type ErrTupleFields struct {
	Writer TupleWriter
	Tuple  Tuple
	Fields []string
}

func (e ErrTupleFields) Error() string {
	return fmt.Sprintf("tuple has invalid fields for this writer (%T). expected: (%v), only have %d", e.Writer, e.Fields, (e.Tuple))
}

// TupleWriter is the type of Writers for structured data.
type TupleWriter interface {
	WriteTuple(row Tuple) error
	Flush() error
}

type Tuple = []interface{}

// CSVWriter writes Tuples in CSV format.
type CSVWriter struct {
	cw      *csv.Writer
	headers []string

	headersWritten bool
	record         []string
}

// NewCSVWriter returns a CSVWriter writing to w.
// If len(headers) == 0, then no headers will be written
func NewCSVWriter(w io.Writer, headers []string) *CSVWriter {
	return &CSVWriter{
		cw:      csv.NewWriter(w),
		headers: headers,
	}
}

func (m *CSVWriter) WriteTuple(row Tuple) error {
	if m.record == nil {
		m.record = make([]string, len(row))
	}
	if len(m.record) != len(row) {
		return ErrTupleFields{
			Writer: m,
			Tuple:  row,
		}
	}
	record := m.record
	if len(m.headers) > 0 && !m.headersWritten {
		if err := m.cw.Write(m.headers); err != nil {
			return errors.EnsureStack(err)
		}
		m.headersWritten = true
	}
	for i := range row {
		var err error
		record[i], err = m.format(row[i])
		if err != nil {
			return err
		}
	}
	return errors.EnsureStack(m.cw.Write(record))
}

func (m *CSVWriter) format(x interface{}) (string, error) {
	const null = "null"
	var y string
	switch x := x.(type) {
	case *int16:
		y = strconv.FormatInt(int64(*x), 10)
	case *int32:
		y = strconv.FormatInt(int64(*x), 10)
	case *int64:
		y = strconv.FormatInt(*x, 10)
	case *uint64:
		y = strconv.FormatUint(*x, 10)
	case *string:
		y = *x
	case *float32:
		y = strconv.FormatFloat(float64(*x), 'f', -1, 32)
	case *float64:
		y = strconv.FormatFloat(*x, 'f', -1, 64)
	case *sql.RawBytes:
		// TODO: what to do here? might not be printable.
		// Maybe have a list of base64 encoded columns.
		y = string(*x)
	case *sql.NullInt16:
		if x.Valid {
			y = strconv.FormatInt(int64(x.Int16), 10)
		} else {
			y = null
		}
	case *sql.NullInt32:
		if x.Valid {
			y = strconv.FormatInt(int64(x.Int32), 10)
		} else {
			y = null
		}
	case *sql.NullInt64:
		if x.Valid {
			y = strconv.FormatInt(x.Int64, 10)
		} else {
			y = null
		}
	case *sql.NullFloat64:
		if x.Valid {
			y = strconv.FormatFloat(x.Float64, 'f', -1, 64)
		} else {
			y = null
		}
	case *sql.NullString:
		if x.Valid {
			y = x.String
		} else {
			y = null
		}
	case *sql.NullTime:
		if x.Valid {
			y = x.Time.Format(time.RFC3339Nano)
		} else {
			y = null
		}
	default:
		return "", errors.Errorf("unrecognized value (%v: %T)", x, x)
	}
	return y, nil
}

func (m *CSVWriter) Flush() error {
	m.cw.Flush()
	return errors.EnsureStack(m.cw.Error())
}

// JSONWriter writes tuples as newline separated json objects.
type JSONWriter struct {
	bufw   *bufio.Writer
	enc    *json.Encoder
	fields []string
	record map[string]interface{}
}

// TODO: figure out some way to specify a projection so that we can write nested structures.
func NewJSONWriter(w io.Writer, fieldNames []string) *JSONWriter {
	bufw := bufio.NewWriter(w)
	enc := json.NewEncoder(bufw)
	return &JSONWriter{
		bufw:   bufw,
		enc:    enc,
		fields: fieldNames,
		record: make(map[string]interface{}, len(fieldNames)),
	}
}

func (m *JSONWriter) WriteTuple(row Tuple) error {
	if len(row) != len(m.fields) {
		return ErrTupleFields{Writer: m, Fields: m.fields, Tuple: row}
	}
	record := m.record
	for i := range row {
		var y interface{}
		switch x := row[i].(type) {
		case *sql.NullInt16:
			if x.Valid {
				y = x.Int16
			} else {
				y = nil
			}
		case *sql.NullInt32:
			if x.Valid {
				y = x.Int32
			} else {
				y = nil
			}
		case *sql.NullInt64:
			if x.Valid {
				y = x.Int64
			} else {
				y = nil
			}
		case *sql.NullFloat64:
			if x.Valid {
				y = x.Float64
			} else {
				y = nil
			}
		case *sql.NullString:
			if x.Valid {
				y = x.String
			} else {
				y = nil
			}
		case *sql.NullTime:
			if x.Valid {
				y = x.Time
			} else {
				y = nil
			}
		default:
			y = row[i]
		}
		record[m.fields[i]] = y
	}
	return errors.EnsureStack(m.enc.Encode(record))
}

func (m *JSONWriter) Flush() error {
	return errors.EnsureStack(m.bufw.Flush())
}
